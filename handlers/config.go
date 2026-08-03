package handlers

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"party2026/models"
)

type ConfigHandler struct {
	config *models.ConfigStore
	themes []string
}

func NewConfigHandler(cfg *models.ConfigStore, themes []string) *ConfigHandler {
	return &ConfigHandler{config: cfg, themes: themes}
}

// GET /admin/config
func (h *ConfigHandler) Page(c echo.Context) error {
	cfg, err := h.config.All()
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "config.html", map[string]interface{}{
		"Config": cfg,
		"Themes": h.themes,
		"Saved":  false,
	})
}

// POST /admin/config
func (h *ConfigHandler) Save(c echo.Context) error {
	keys := []string{
		"party_date", "party_time_start",
		"party_name_de", "party_name_en",
		"rsvp_deadline",
		"smtp_from_name", "invite_message_de", "invite_message_en",
	}
	for _, k := range keys {
		if v := c.FormValue(k); v != "" {
			_ = h.config.Set(k, v)
		}
	}

	if c.FormValue("calendar_enabled") == "1" {
		_ = h.config.Set("calendar_enabled", "1")
	} else {
		_ = h.config.Set("calendar_enabled", "0")
	}
	for _, k := range []string{
		"calendar_time_end", "calendar_location",
		"calendar_description_de", "calendar_description_en",
	} {
		_ = h.config.Set(k, c.FormValue(k))
	}

	// Feature toggles (default ON) — persist explicit 1/0 from checkboxes.
	for _, k := range []string{"gallery_enabled", "carpool_enabled", "rsvp_enabled"} {
		if c.FormValue(k) == "1" {
			_ = h.config.Set(k, "1")
		} else {
			_ = h.config.Set(k, "0")
		}
	}

	_ = h.config.Set("hero_image_url", sanitizeImageURL(c.FormValue("hero_image_url")))
	_ = h.config.Set("rsvp_notify_email", strings.TrimSpace(c.FormValue("rsvp_notify_email")))

	// Theme selectors — allow empty selection (keep existing)
	for _, k := range []string{"theme_login", "theme_me"} {
		if v := c.FormValue(k); v != "" {
			_ = h.config.Set(k, v)
		}
	}

	cfg, _ := h.config.All()
	return c.Render(http.StatusOK, "config.html", map[string]interface{}{
		"Config": cfg,
		"Themes": h.themes,
		"Saved":  true,
	})
}
