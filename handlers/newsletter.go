package handlers

import (
	"bytes"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/yuin/goldmark"
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

	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(bodyMD), &buf); err != nil {
		buf.WriteString(bodyMD)
	}
	bodyHTML := buf.String()

	recipients, err := h.guests.NewsletterRecipients()
	if err != nil {
		return err
	}

	cfg, _ := h.config.All()
	sendErr := h.mailer.SendNewsletter(recipients, subject, bodyHTML, cfg)

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
