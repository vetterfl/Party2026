package handlers

import (
	"net/http"
	"strings"

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

func (h *NewsletterHandler) pageData(c echo.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	recipients, err := h.guests.NewsletterRecipients()
	if err != nil {
		return nil, err
	}
	cfg, err := h.config.All()
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{
		"RecipientCount": len(recipients),
		"AdminEmail":     strings.TrimSpace(cfg["rsvp_notify_email"]),
		"Subject":        c.FormValue("subject"),
		"Body":           c.FormValue("body"),
		"Sent":           false,
		"TestSent":       false,
		"Error":          "",
		"TestError":      "",
	}
	for k, v := range extra {
		data[k] = v
	}
	return data, nil
}

// GET /admin/newsletter
func (h *NewsletterHandler) Page(c echo.Context) error {
	data, err := h.pageData(c, nil)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "newsletter.html", data)
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

	extra := map[string]interface{}{"Sent": sendErr == nil}
	if sendErr != nil {
		extra["Error"] = sendErr.Error()
	}

	data, err := h.pageData(c, extra)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "newsletter.html", data)
}

// POST /admin/newsletter/test
func (h *NewsletterHandler) SendTest(c echo.Context) error {
	subject := c.FormValue("subject")
	bodyMD := c.FormValue("body")

	cfg, _ := h.config.All()
	sendErr := h.mailer.SendNewsletterTest(cfg["rsvp_notify_email"], subject, bodyMD, cfg)

	extra := map[string]interface{}{"TestSent": sendErr == nil}
	if sendErr != nil {
		extra["TestError"] = sendErr.Error()
	}

	data, err := h.pageData(c, extra)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "newsletter.html", data)
}
