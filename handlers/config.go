package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
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
		"rsvp_deadline", "charity_name", "charity_url",
		"smtp_from_name", "invite_message_de", "invite_message_en",
		"admin_user",
	}
	for _, k := range keys {
		if v := c.FormValue(k); v != "" {
			_ = h.config.Set(k, v)
		}
	}

	// Theme selectors — allow empty selection (keep existing)
	for _, k := range []string{"theme_login", "theme_me"} {
		if v := c.FormValue(k); v != "" {
			_ = h.config.Set(k, v)
		}
	}

	// Update password only if provided
	if newPass := c.FormValue("admin_password"); newPass != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(newPass), bcrypt.DefaultCost)
		if err == nil {
			_ = h.config.Set("admin_password_hash", string(hash))
		}
	}

	cfg, _ := h.config.All()
	return c.Render(http.StatusOK, "config.html", map[string]interface{}{
		"Config": cfg,
		"Themes": h.themes,
		"Saved":  true,
	})
}
