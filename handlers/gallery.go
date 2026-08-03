package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
	"party2026/middleware"
	"party2026/models"
)

// photoNameRe matches the export's filenames: a UUID-ish base, an optional
// `_thumb` suffix, and a `.jpg` extension. Used to reject anything that could
// escape the photos directory (path traversal) before we touch the disk.
var photoNameRe = regexp.MustCompile(`^[a-zA-Z0-9-]+(_thumb)?\.jpg$`)

// Photo pairs a full-size image with its thumbnail. Both are served through the
// guarded /me/gallery/img route, never as raw static files.
type Photo struct {
	Full  string // filename of the full-size image, e.g. "abc.jpg"
	Thumb string // filename of the thumbnail; falls back to Full when absent
}

type GalleryHandler struct {
	guests  *models.GuestStore
	config  *models.ConfigStore
	content *models.ContentStore

	dir    string
	mu     sync.RWMutex
	photos []Photo
	valid  map[string]bool // set of servable filenames (full + thumb)
}

func NewGalleryHandler(g *models.GuestStore, cfg *models.ConfigStore, cnt *models.ContentStore) *GalleryHandler {
	dir := os.Getenv("PHOTOS_DIR")
	if dir == "" {
		dir = "photos"
	}
	h := &GalleryHandler{guests: g, config: cfg, content: cnt, dir: dir}
	h.scan()
	return h
}

// scan reads the photos directory once and builds the ordered photo list plus
// the allow-list of servable filenames. Photos are static, so a single scan at
// startup is enough.
func (h *GalleryHandler) scan() {
	entries, err := os.ReadDir(h.dir)
	if err != nil {
		return
	}

	valid := map[string]bool{}
	thumbs := map[string]bool{}
	var fulls []string

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".jpg") || !photoNameRe.MatchString(name) {
			continue
		}
		valid[name] = true
		if strings.HasSuffix(name, "_thumb.jpg") {
			thumbs[name] = true
		} else {
			fulls = append(fulls, name)
		}
	}

	sort.Strings(fulls)
	photos := make([]Photo, 0, len(fulls))
	for _, full := range fulls {
		thumb := strings.TrimSuffix(full, ".jpg") + "_thumb.jpg"
		if !thumbs[thumb] {
			thumb = full // no thumbnail on disk — use the full image
		}
		photos = append(photos, Photo{Full: full, Thumb: thumb})
	}

	h.mu.Lock()
	h.photos = photos
	h.valid = valid
	h.mu.Unlock()
}

func (h *GalleryHandler) loadBase(c echo.Context, g *models.Guest) (BaseData, error) {
	cfg, err := h.config.All()
	if err != nil {
		return BaseData{}, err
	}
	blocks, err := h.content.All()
	if err != nil {
		return BaseData{}, err
	}
	cm := map[string]models.ContentBlock{}
	for _, b := range blocks {
		cm[b.Key] = b
	}
	theme := cfg["theme_me"]
	if theme == "" {
		theme = "midnight-pool"
	}
	return newBase(c, cfg, cm, g, theme), nil
}

// GET /me/gallery — after-party photo gallery
func (h *GalleryHandler) Gallery(c echo.Context) error {
	guest, err := h.guests.FindByID(middleware.GetGuestID(c))
	if err != nil {
		_ = middleware.ClearGuestSession(c)
		return c.Redirect(http.StatusSeeOther, "/")
	}
	_ = h.guests.IncrementViewCount(guest.ID)

	bd, err := h.loadBase(c, guest)
	if err != nil {
		return err
	}

	h.mu.RLock()
	photos := h.photos
	h.mu.RUnlock()

	return c.Render(http.StatusOK, "gallery.html", map[string]interface{}{
		"Base":   bd,
		"Photos": photos,
	})
}

// GET /me/gallery/img/:name — serve a single photo to logged-in guests only.
// The name is validated against the on-disk allow-list, so path traversal and
// requests for arbitrary files are rejected.
func (h *GalleryHandler) Serve(c echo.Context) error {
	name := c.Param("name")
	if !photoNameRe.MatchString(name) {
		return echo.ErrNotFound
	}

	h.mu.RLock()
	ok := h.valid[name]
	h.mu.RUnlock()
	if !ok {
		return echo.ErrNotFound
	}

	// filepath.Base is belt-and-suspenders: the regex already forbids slashes.
	path := filepath.Join(h.dir, filepath.Base(name))

	// Photos never change, so let browsers cache them aggressively.
	c.Response().Header().Set("Cache-Control", "private, max-age=604800, immutable")
	return c.File(path)
}
