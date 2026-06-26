package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"party2026/middleware"
	"party2026/models"
)

type CarpoolHandler struct {
	guests  *models.GuestStore
	config  *models.ConfigStore
	content *models.ContentStore
	carpool *models.CarpoolStore
}

func NewCarpoolHandler(g *models.GuestStore, cfg *models.ConfigStore, cnt *models.ContentStore, cp *models.CarpoolStore) *CarpoolHandler {
	return &CarpoolHandler{guests: g, config: cfg, content: cnt, carpool: cp}
}

func (h *CarpoolHandler) loadBase(c echo.Context, g *models.Guest) (BaseData, error) {
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

// GET /me/carpool — carpool message board
func (h *CarpoolHandler) Board(c echo.Context) error {
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

	posts, err := h.carpool.List()
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "carpool.html", map[string]interface{}{
		"Base":  bd,
		"Posts": posts,
	})
}

// POST /me/carpool — create a carpool post for the current guest
func (h *CarpoolHandler) Create(c echo.Context) error {
	guest, err := h.guests.FindByID(middleware.GetGuestID(c))
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/")
	}

	kind := strings.ToLower(strings.TrimSpace(c.FormValue("kind")))
	if !models.ValidCarpoolKind(kind) {
		kind = "offer"
	}

	contact := strings.TrimSpace(c.FormValue("contact"))
	if contact == "" {
		// Contact is required — bounce back to the board without saving.
		return c.Redirect(http.StatusSeeOther, "/me/carpool")
	}

	seats := 0
	if s := strings.TrimSpace(c.FormValue("seats")); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			if n > 99 {
				n = 99
			}
			seats = n
		}
	}

	post := &models.CarpoolPost{
		GuestID:    guest.ID,
		Kind:       kind,
		Origin:     strings.TrimSpace(c.FormValue("origin")),
		TravelTime: strings.TrimSpace(c.FormValue("travel_time")),
		Seats:      seats,
		Note:       strings.TrimSpace(c.FormValue("note")),
		Contact:    contact,
	}
	if err := h.carpool.Create(post); err != nil {
		return err
	}
	_ = h.guests.IncrementInteractionCount(guest.ID)

	return c.Redirect(http.StatusSeeOther, "/me/carpool")
}

// POST /me/carpool/:id/delete — delete one of the current guest's posts
func (h *CarpoolHandler) Delete(c echo.Context) error {
	guest, err := h.guests.FindByID(middleware.GetGuestID(c))
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/")
	}
	if err := h.carpool.Delete(c.Param("id"), guest.ID); err != nil {
		return err
	}
	return c.Redirect(http.StatusSeeOther, "/me/carpool")
}
