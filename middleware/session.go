package middleware

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
)

const (
	GuestSessionName = "guest_session"
	GuestIDKey       = "guest_id"
	AdminSessionName = "admin_session"
	AdminIDKey       = "admin_id"
	LangCookie       = "lang"

	guestSessionMaxAge = 86400 * 30 // 30 days
	adminSessionMaxAge = 86400      // 24 hours
)

var (
	store      *sessions.CookieStore
	cookieOpts sessions.Options
)

func InitStore() {
	secret := os.Getenv("SESSION_SECRET")
	if isProduction() {
		if secret == "" || strings.HasPrefix(secret, "change-this") {
			log.Fatal("SESSION_SECRET must be set (>=32 random bytes hex) in production")
		}
	}
	if secret == "" {
		secret = "dev-secret-change-in-production-please"
	}
	store = sessions.NewCookieStore([]byte(secret))
	cookieOpts = sessions.Options{
		Path:     "/",
		HttpOnly: true,
		Secure:   isProduction(),
		SameSite: http.SameSiteLaxMode,
	}
	store.Options = &sessions.Options{
		Path:     cookieOpts.Path,
		MaxAge:   guestSessionMaxAge,
		HttpOnly: cookieOpts.HttpOnly,
		Secure:   cookieOpts.Secure,
		SameSite: cookieOpts.SameSite,
	}
}

func isProduction() bool {
	return strings.EqualFold(os.Getenv("GO_ENV"), "production")
}

func sessionOptions(maxAge int) *sessions.Options {
	return &sessions.Options{
		Path:     cookieOpts.Path,
		MaxAge:   maxAge,
		HttpOnly: cookieOpts.HttpOnly,
		Secure:   cookieOpts.Secure,
		SameSite: cookieOpts.SameSite,
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
	sess.Options = sessionOptions(guestSessionMaxAge)
	sess.Values[GuestIDKey] = id
	return sess.Save(c.Request(), c.Response().Writer)
}

func ClearGuestSession(c echo.Context) error {
	sess, err := store.Get(c.Request(), GuestSessionName)
	if err != nil {
		return err
	}
	sess.Options = sessionOptions(-1)
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
	sess.Options = sessionOptions(adminSessionMaxAge)
	sess.Values[AdminIDKey] = adminID
	return sess.Save(c.Request(), c.Response().Writer)
}

func ClearAdminSession(c echo.Context) error {
	sess, err := store.Get(c.Request(), AdminSessionName)
	if err != nil {
		return err
	}
	sess.Options = sessionOptions(-1)
	return sess.Save(c.Request(), c.Response().Writer)
}

func GetLang(c echo.Context) string {
	cookie, err := c.Cookie(LangCookie)
	if err == nil && (cookie.Value == "en" || cookie.Value == "de") {
		return cookie.Value
	}
	return "de"
}

// SafeReferer returns the Referer path when the Referer host matches the
// request host. Otherwise returns fallback. Prevents open-redirect via Referer.
func SafeReferer(c echo.Context, fallback string) string {
	ref := c.Request().Referer()
	if ref == "" {
		return fallback
	}
	u, err := url.Parse(ref)
	if err != nil || u.Host != c.Request().Host {
		return fallback
	}
	path := u.RequestURI()
	if path == "" || !strings.HasPrefix(path, "/") {
		return fallback
	}
	return path
}

// RequireGuest redirects to / if no guest session.
func RequireGuest(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if GetGuestID(c) == "" {
			return c.Redirect(http.StatusSeeOther, "/")
		}
		return next(c)
	}
}

// RequireAdmin redirects to /admin/login if not authed.
func RequireAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !IsAdminAuthed(c) {
			return c.Redirect(http.StatusSeeOther, "/admin/login")
		}
		return next(c)
	}
}
