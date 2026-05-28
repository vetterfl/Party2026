package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"party2026/models"
)

type NewsletterHandler struct {
	guests *models.GuestStore
	config *models.ConfigStore
	mailer *Mailer
}

func NewNewsletterHandler(g *models.GuestStore, cfg *models.ConfigStore, m *Mailer) *NewsletterHandler {
	return &NewsletterHandler{guests: g, config: cfg, mailer: m}
}

// GET /admin/newsletter
func (h *NewsletterHandler) Page(c echo.Context) error {
	recipients, err := h.guests.NewsletterRecipients()
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "newsletter.html", map[string]interface{}{
		"RecipientCount": len(recipients),
		"Sent":           false,
		"Error":          "",
	})
}

// POST /admin/newsletter
func (h *NewsletterHandler) Send(c echo.Context) error {
	subject := c.FormValue("subject")
	bodyMD := c.FormValue("body")

	recipients, err := h.guests.NewsletterRecipients()
	if err != nil {
		return err
	}

	cfg, _ := h.config.All()
	sendErr := h.mailer.SendNewsletter(recipients, subject, bodyMD, cfg)

	errMsg := ""
	if sendErr != nil {
		errMsg = sendErr.Error()
	}

	return c.Render(http.StatusOK, "newsletter.html", map[string]interface{}{
		"RecipientCount": len(recipients),
		"Sent":           sendErr == nil,
		"Error":          errMsg,
	})
}
