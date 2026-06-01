package main

import (
	"bytes"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/time/rate"

	embedded "party2026"
	"party2026/handlers"
	mw "party2026/middleware"
	"party2026/migrations"
	"party2026/models"
)

func main() {
	_ = godotenv.Load()

	if len(os.Args) > 1 && os.Args[1] == "mailtest" {
		runMailTest()
		return
	}

	mw.InitStore()

	db, err := models.Open(migrations.FS)
	if err != nil {
		log.Fatalf("db: %v", err)
	}

	guestStore := models.NewGuestStore(db)
	configStore := models.NewConfigStore(db)
	contentStore := models.NewContentStore(db)
	adminStore := models.NewAdminStore(db)

	bootstrapAdmin(adminStore)

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:" + port()
	}

	templatesSub, err := fs.Sub(embedded.Templates, "templates")
	if err != nil {
		log.Fatalf("templates sub-fs: %v", err)
	}

	mailer, err := handlers.NewMailer(templatesSub, baseURL)
	if err != nil {
		log.Fatalf("mailer: %v", err)
	}

	renderer, err := handlers.NewRenderer(templatesSub)
	if err != nil {
		log.Fatalf("renderer: %v", err)
	}

	e := echo.New()
	e.HideBanner = true
	e.Renderer = renderer

	// Trust X-Forwarded-For when behind nginx (single proxy hop).
	e.IPExtractor = echo.ExtractIPFromXFFHeader()

	e.Use(loggerMiddleware())
	e.Use(echomw.Recover())
	e.Use(echomw.Secure())
	e.Use(csrfMiddleware())

	// Static assets — exempt from middleware that does not run on GETs anyway.
	assetsSub, _ := fs.Sub(embedded.Assets, "assets")
	e.GET("/assets/*", echo.WrapHandler(
		http.StripPrefix("/assets/", http.FileServer(http.FS(assetsSub))),
	))

	// Enumerate installed themes from embedded assets.
	var themes []string
	if themesFS, err := fs.Sub(embedded.Assets, "assets/themes"); err == nil {
		if entries, err := fs.ReadDir(themesFS, "."); err == nil {
			for _, ent := range entries {
				if ent.IsDir() {
					themes = append(themes, ent.Name())
				}
			}
		}
	}

	gh := handlers.NewGuestHandler(guestStore, configStore, contentStore, mailer)
	ah := handlers.NewAdminHandler(guestStore, configStore, contentStore, adminStore, mailer, baseURL)
	ch := handlers.NewContentHandler(contentStore)
	nh := handlers.NewNewsletterHandler(guestStore, configStore, mailer)
	cfgh := handlers.NewConfigHandler(configStore, themes)
	auh := handlers.NewAdminUsersHandler(adminStore)

	loginRL := loginRateLimiter()

	// Guest routes
	e.GET("/", gh.SpellPage)
	e.POST("/login", gh.Login, loginRL)
	e.GET("/logout", gh.Logout)
	e.POST("/lang", gh.SetLang)
	e.GET("/unsubscribe", gh.Unsubscribe)
	e.GET("/datenschutz", gh.Privacy)

	me := e.Group("/me", mw.RequireGuest)
	me.GET("", gh.Me)
	me.GET("/rsvp", gh.RSVPForm)
	me.POST("/rsvp", gh.RSVPSubmit)
	me.GET("/confirmed", gh.Confirmed)
	me.GET("/calendar.ics", gh.CalendarICS)

	// Admin routes
	e.GET("/admin/login", ah.LoginPage)
	e.POST("/admin/login", ah.LoginSubmit, loginRL)
	e.GET("/admin/logout", ah.Logout)

	admin := e.Group("/admin", mw.RequireAdmin)
	admin.GET("", ah.Dashboard)
	admin.GET("/guests", ah.GuestList)
	admin.POST("/guests", ah.GuestCreate)
	admin.GET("/guests/:id/edit", ah.GuestEdit)
	admin.POST("/guests/:id/edit", ah.GuestUpdate)
	admin.POST("/guests/:id/delete", ah.GuestDelete)
	admin.GET("/guests/:id/qr", ah.GuestQR)
	admin.GET("/guests/:id/card", ah.GuestCard)
	admin.GET("/export/csv", ah.ExportCSV)
	admin.GET("/content", ch.List)
	admin.GET("/content/:key", ch.Edit)
	admin.POST("/content/:key", ch.Save)
	admin.GET("/newsletter", nh.Page)
	admin.POST("/newsletter", nh.Send)
	admin.GET("/config", cfgh.Page)
	admin.POST("/config", cfgh.Save)
	admin.GET("/admins", auh.List)
	admin.POST("/admins", auh.Create)
	admin.POST("/admins/:id/delete", auh.Delete)

	addr := ":" + port()
	log.Printf("listening on %s", addr)
	log.Fatal(e.Start(addr))
}

// csrfMiddleware applies CSRF token validation on unsafe methods. Tokens are
// stored in a cookie (_csrf) and required in the `_csrf` form field. The token
// value is exposed to templates via context key "csrf".
func csrfMiddleware() echo.MiddlewareFunc {
	cfg := echomw.DefaultCSRFConfig
	cfg.TokenLookup = "form:_csrf,header:X-CSRF-Token"
	cfg.CookiePath = "/"
	cfg.CookieHTTPOnly = true
	cfg.CookieSameSite = http.SameSiteLaxMode
	cfg.CookieSecure = isProduction()
	cfg.ContextKey = "csrf"
	return echomw.CSRFWithConfig(cfg)
}

// loginRateLimiter throttles login attempts per IP — 1 req/sec sustained,
// burst of 10. Mitigates brute force on /login and /admin/login.
func loginRateLimiter() echo.MiddlewareFunc {
	return echomw.RateLimiterWithConfig(echomw.RateLimiterConfig{
		Store: echomw.NewRateLimiterMemoryStoreWithConfig(
			echomw.RateLimiterMemoryStoreConfig{
				Rate:      rate.Limit(1),
				Burst:     10,
				ExpiresIn: 10 * time.Minute,
			},
		),
		IdentifierExtractor: func(c echo.Context) (string, error) {
			return c.RealIP(), nil
		},
		DenyHandler: func(c echo.Context, _ string, _ error) error {
			return echo.NewHTTPError(http.StatusTooManyRequests, "Too many attempts, slow down.")
		},
	})
}

// loggerMiddleware logs each request but redacts the `spell` query parameter
// (a guest's login code) to avoid leaking secrets into access logs.
func loggerMiddleware() echo.MiddlewareFunc {
	return echomw.LoggerWithConfig(echomw.LoggerConfig{
		Format: `{"time":"${time_rfc3339}","remote_ip":"${remote_ip}","method":"${method}","uri":"${custom}","status":${status},"latency":"${latency_human}"}` + "\n",
		CustomTagFunc: func(c echo.Context, buf *bytes.Buffer) (int, error) {
			u := *c.Request().URL
			if q := u.Query(); q.Get("spell") != "" {
				q.Set("spell", "REDACTED")
				u.RawQuery = q.Encode()
			}
			return buf.WriteString(u.RequestURI())
		},
	})
}

func isProduction() bool {
	return strings.EqualFold(os.Getenv("GO_ENV"), "production")
}

func port() string {
	if p := os.Getenv("PORT"); p != "" {
		return p
	}
	return "3000"
}

func bootstrapAdmin(store *models.AdminStore) {
	u := os.Getenv("ADMIN_USER")
	p := os.Getenv("ADMIN_PASSWORD")
	if u == "" || p == "" {
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	if err != nil {
		return
	}
	existing, err := store.FindByUsername(u)
	if err == nil {
		_ = store.UpdatePassword(existing.ID, string(hash))
		log.Printf("admin password updated from env for %q", u)
		return
	}
	count, err := store.Count()
	if err != nil || count > 0 {
		return
	}
	if _, err := store.Create(u, string(hash)); err != nil {
		log.Printf("admin bootstrap failed: %v", err)
		return
	}
	log.Printf("admin account created from env for %q", u)
}

func runMailTest() {
	if len(os.Args) < 3 {
		log.Fatalf("usage: %s mailtest <recipient@email.com>", os.Args[0])
	}
	to := os.Args[2]

	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	if port == "" {
		port = "587"
	}
	from := os.Getenv("SMTP_FROM")

	log.Printf("SMTP host=%s port=%s from=%s", host, port, from)
	log.Printf("Sending test email to %s…", to)

	cfg := map[string]string{}
	if name := os.Getenv("SMTP_FROM_NAME"); name != "" {
		cfg["smtp_from_name"] = name
	}

	if err := handlers.SendTestMail(to, cfg); err != nil {
		log.Fatalf("mail test failed: %v", err)
	}
	log.Println("mail test OK")
}
