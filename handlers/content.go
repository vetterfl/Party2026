package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"party2026/models"
)

type ContentHandler struct {
	content *models.ContentStore
}

func NewContentHandler(c *models.ContentStore) *ContentHandler {
	return &ContentHandler{content: c}
}

// GET /admin/content
func (h *ContentHandler) List(c echo.Context) error {
	blocks, err := h.content.All()
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "content.html", map[string]interface{}{
		"Blocks": blocks,
	})
}

// GET /admin/content/:key
func (h *ContentHandler) Edit(c echo.Context) error {
	block, err := h.content.Get(c.Param("key"))
	if err != nil {
		return echo.ErrNotFound
	}
	return c.Render(http.StatusOK, "content_edit.html", map[string]interface{}{
		"Block": block,
	})
}

// POST /admin/content/:key
func (h *ContentHandler) Save(c echo.Context) error {
	key := c.Param("key")
	bodyDE := c.FormValue("body_de")
	bodyEN := c.FormValue("body_en")
	if err := h.content.Save(key, bodyDE, bodyEN); err != nil {
		return err
	}
	return c.Redirect(http.StatusSeeOther, "/admin/content")
}
