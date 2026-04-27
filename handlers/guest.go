package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"party2026/middleware"
	"party2026/models"
)

type GuestHandler struct {
	guests  *models.GuestStore
	config  *models.ConfigStore
	content *models.ContentStore
	mailer  *Mailer
}

func NewGuestHandler(g *models.GuestStore, cfg *models.ConfigStore, cnt *models.ContentStore, m *Mailer) *GuestHandler {
	return &GuestHandler{guests: g, config: cfg, content: cnt, mailer: m}
}

func (h *GuestHandler) loadBase(c echo.Context, g *models.Guest) (BaseData, error) {
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
	return newBase(c, cfg, cm, g), nil
}

// GET / — spell login page
func (h *GuestHandler) SpellPage(c echo.Context) error {
	// Already logged in → redirect to /me
	if middleware.GetGuestID(c) != "" {
		return c.Redirect(http.StatusSeeOther, "/me")
	}

	bd, err := h.loadBase(c, nil)
	if err != nil {
		return err
	}

	spell := c.QueryParam("spell")
	return c.Render(http.StatusOK, "spell.html", map[string]interface{}{
		"Base":      bd,
		"PrefillSpell": spell,
		"Error":     false,
	})
}

// POST /login — authenticate with spell
func (h *GuestHandler) Login(c echo.Context) error {
	code := strings.TrimSpace(strings.ToUpper(c.FormValue("spell")))

	guest, err := h.guests.FindByCode(code)
	if err != nil {
		bd, _ := h.loadBase(c, nil)
		return c.Render(http.StatusOK, "spell.html", map[string]interface{}{
			"Base":         bd,
			"PrefillSpell": "",
			"Error":        true,
		})
	}

	if err := middleware.SetGuestID(c, guest.ID); err != nil {
		return err
	}
	return c.Redirect(http.StatusSeeOther, "/me")
}

// GET /logout
func (h *GuestHandler) Logout(c echo.Context) error {
	_ = middleware.ClearGuestSession(c)
	return c.Redirect(http.StatusSeeOther, "/")
}

// GET /me — personal invite page
func (h *GuestHandler) Me(c echo.Context) error {
	guest, err := h.guests.FindByID(middleware.GetGuestID(c))
	if err != nil {
		_ = middleware.ClearGuestSession(c)
		return c.Redirect(http.StatusSeeOther, "/")
	}

	bd, err := h.loadBase(c, guest)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "index.html", map[string]interface{}{"Base": bd})
}

// GET /me/rsvp — RSVP form prefilled
func (h *GuestHandler) RSVPForm(c echo.Context) error {
	guest, err := h.guests.FindByID(middleware.GetGuestID(c))
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/")
	}
	bd, err := h.loadBase(c, guest)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "rsvp.html", map[string]interface{}{"Base": bd})
}

// POST /me/rsvp — submit RSVP
func (h *GuestHandler) RSVPSubmit(c echo.Context) error {
	guest, err := h.guests.FindByID(middleware.GetGuestID(c))
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/")
	}

	attending := c.FormValue("attending") == "yes"

	if attending {
		guest.Status = "accepted"
	} else {
		guest.Status = "declined"
	}

	guest.PlusOne = c.FormValue("plus_one") == "yes"
	if name := strings.TrimSpace(c.FormValue("plus_one_name")); name != "" && guest.PlusOne {
		guest.PlusOneName = &name
	} else {
		guest.PlusOneName = nil
	}

	children := 0
	if ch := c.FormValue("children"); ch != "" {
		for _, r := range ch {
			if r >= '0' && r <= '5' {
				children = int(r - '0')
				break
			}
		}
	}
	guest.Children = children

	if song := strings.TrimSpace(c.FormValue("song")); song != "" {
		guest.Song = &song
	} else {
		guest.Song = nil
	}
	if comment := strings.TrimSpace(c.FormValue("comment")); comment != "" {
		guest.Comment = &comment
	} else {
		guest.Comment = nil
	}
	if email := strings.TrimSpace(c.FormValue("email")); email != "" {
		guest.Email = &email
		guest.Newsletter = c.FormValue("newsletter") == "on"
	} else {
		guest.Email = nil
		guest.Newsletter = false
	}

	now := time.Now()
	guest.RSVPAt = &now

	if err := h.guests.Update(guest); err != nil {
		return err
	}

	// Send confirmation email asynchronously
	if guest.Email != nil && *guest.Email != "" {
		cfg, _ := h.config.All()
		go h.mailer.SendConfirmation(guest, cfg, middleware.GetLang(c))
	}

	return c.Redirect(http.StatusSeeOther, "/me/confirmed")
}

// GET /me/confirmed
func (h *GuestHandler) Confirmed(c echo.Context) error {
	guest, err := h.guests.FindByID(middleware.GetGuestID(c))
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/")
	}
	bd, err := h.loadBase(c, guest)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "confirmed.html", map[string]interface{}{"Base": bd})
}

// GET /unsubscribe?token=...
func (h *GuestHandler) Unsubscribe(c echo.Context) error {
	token := c.QueryParam("token")
	id := h.mailer.ValidateUnsubToken(token)
	if id == "" {
		return c.String(http.StatusBadRequest, "Invalid token.")
	}
	_ = h.guests.UnsubscribeByID(id)
	lang := middleware.GetLang(c)
	msg := "Du wurdest vom Newsletter abgemeldet."
	if lang == "en" {
		msg = "You have been unsubscribed."
	}
	return c.String(http.StatusOK, msg)
}

// POST /lang — set language cookie
func (h *GuestHandler) SetLang(c echo.Context) error {
	lang := c.FormValue("lang")
	if lang != "en" {
		lang = "de"
	}
	cookie := new(http.Cookie)
	cookie.Name = middleware.LangCookie
	cookie.Value = lang
	cookie.Path = "/"
	cookie.MaxAge = 86400 * 365
	c.SetCookie(cookie)
	ref := c.Request().Referer()
	if ref == "" {
		ref = "/"
	}
	return c.Redirect(http.StatusSeeOther, ref)
}
