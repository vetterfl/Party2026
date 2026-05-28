package middleware

import (
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
)

const (
	GuestSessionName = "guest_session"
	GuestIDKey       = "guest_id"
	AdminSessionName = "admin_session"
	AdminIDKey       = "admin_id"
	LangCookie       = "lang"
)

var store *sessions.CookieStore

func InitStore() {
	secret := os.Getenv("SESSION_SECRET")
	if secret == "" {
		secret = "dev-secret-change-in-production-please"
	}
	store = sessions.NewCookieStore([]byte(secret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}


func GetGuestID(c echo.Context) string {
	sess, err := store.Get(c.Request(), GuestSessionName)
	if err != nil {
		return ""
	}
	id, _ := sess.Values[GuestIDKey].(string)
	return id
}

func SetGuestID(c echo.Context, id string) error {
	sess, err := store.Get(c.Request(), GuestSessionName)
	if err != nil {
		return err
	}
	sess.Values[GuestIDKey] = id
	return sess.Save(c.Request(), c.Response().Writer)
}

func ClearGuestSession(c echo.Context) error {
	sess, err := store.Get(c.Request(), GuestSessionName)
	if err != nil {
		return err
	}
	sess.Options.MaxAge = -1
	return sess.Save(c.Request(), c.Response().Writer)
}

func GetAdminID(c echo.Context) string {
	sess, err := store.Get(c.Request(), AdminSessionName)
	if err != nil {
		return ""
	}
	id, _ := sess.Values[AdminIDKey].(string)
	return id
}

func IsAdminAuthed(c echo.Context) bool {
	return GetAdminID(c) != ""
}

func SetAdminAuthed(c echo.Context, adminID string) error {
	sess, err := store.Get(c.Request(), AdminSessionName)
	if err != nil {
		return err
	}
	sess.Values[AdminIDKey] = adminID
	return sess.Save(c.Request(), c.Response().Writer)
}

func ClearAdminSession(c echo.Context) error {
	sess, err := store.Get(c.Request(), AdminSessionName)
	if err != nil {
		return err
	}
	sess.Options.MaxAge = -1
	return sess.Save(c.Request(), c.Response().Writer)
}

func GetLang(c echo.Context) string {
	cookie, err := c.Cookie(LangCookie)
	if err == nil && (cookie.Value == "en" || cookie.Value == "de") {
		return cookie.Value
	}
	return "de"
}

// RequireGuest middleware — redirects to / if no guest session
func RequireGuest(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if GetGuestID(c) == "" {
			return c.Redirect(http.StatusSeeOther, "/")
		}
		return next(c)
	}
}

// RequireAdmin middleware — redirects to /admin/login if not authed
func RequireAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !IsAdminAuthed(c) {
			return c.Redirect(http.StatusSeeOther, "/admin/login")
		}
		return next(c)
	}
}
