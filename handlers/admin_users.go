package handlers

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"party2026/middleware"
	"party2026/models"
)

type AdminUsersHandler struct {
	admins *models.AdminStore
}

func NewAdminUsersHandler(admins *models.AdminStore) *AdminUsersHandler {
	return &AdminUsersHandler{admins: admins}
}

// GET /admin/admins
func (h *AdminUsersHandler) List(c echo.Context) error {
	admins, err := h.admins.All()
	if err != nil {
		return err
	}
	errMsg := ""
	switch c.QueryParam("error") {
	case "self":
		errMsg = "You cannot delete your own account."
	case "last":
		errMsg = "At least one admin must remain."
	}
	return c.Render(http.StatusOK, "admins.html", map[string]interface{}{
		"Admins":   admins,
		"Error":    errMsg,
		"Created":  false,
		"SelfID":   middleware.GetAdminID(c),
	})
}

// POST /admin/admins
func (h *AdminUsersHandler) Create(c echo.Context) error {
	username := strings.TrimSpace(c.FormValue("username"))
	password := c.FormValue("password")

	admins, _ := h.admins.All()
	render := func(errMsg string) error {
		return c.Render(http.StatusOK, "admins.html", map[string]interface{}{
			"Admins":  admins,
			"Error":   errMsg,
			"Created": false,
			"SelfID":  middleware.GetAdminID(c),
		})
	}

	if username == "" || password == "" {
		return render("Username and password are required.")
	}
	if len(password) < 8 {
		return render("Password must be at least 8 characters.")
	}
	if _, err := h.admins.FindByUsername(username); err == nil {
		return render("That username is already taken.")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if _, err := h.admins.Create(username, string(hash)); err != nil {
		return err
	}

	admins, _ = h.admins.All()
	return c.Render(http.StatusOK, "admins.html", map[string]interface{}{
		"Admins":  admins,
		"Error":   "",
		"Created": true,
		"SelfID":  middleware.GetAdminID(c),
	})
}

// POST /admin/admins/:id/delete
func (h *AdminUsersHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	if id == middleware.GetAdminID(c) {
		return c.Redirect(http.StatusSeeOther, "/admin/admins?error=self")
	}

	count, err := h.admins.Count()
	if err != nil {
		return err
	}
	if count <= 1 {
		return c.Redirect(http.StatusSeeOther, "/admin/admins?error=last")
	}

	if err := h.admins.Delete(id); err != nil {
		return err
	}
	return c.Redirect(http.StatusSeeOther, "/admin/admins")
}
