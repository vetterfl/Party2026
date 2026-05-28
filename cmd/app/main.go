package main

import (
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"golang.org/x/crypto/bcrypt"

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

	// Bootstrap admin credentials from env (only if set)
	if u := os.Getenv("ADMIN_USER"); u != "" {
		_ = configStore.Set("admin_user", u)
	}
	if p := os.Getenv("ADMIN_PASSWORD"); p != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
		if err == nil {
			_ = configStore.Set("admin_password_hash", string(hash))
			log.Println("admin password updated from env")
		}
	}

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
	e.Use(echomw.Logger())
	e.Use(echomw.Recover())

	// Static assets
	assetsSub, _ := fs.Sub(embedded.Assets, "assets")
	e.GET("/assets/*", echo.WrapHandler(
		http.StripPrefix("/assets/", http.FileServer(http.FS(assetsSub))),
	))

	// Enumerate installed themes from embedded assets
	var themes []string
	if themesFS, err := fs.Sub(embedded.Assets, "assets/themes"); err == nil {
		if entries, err := fs.ReadDir(themesFS, "."); err == nil {
			for _, e := range entries {
				if e.IsDir() {
					themes = append(themes, e.Name())
				}
			}
		}
	}

	gh := handlers.NewGuestHandler(guestStore, configStore, contentStore, mailer)
	ah := handlers.NewAdminHandler(guestStore, configStore, contentStore, mailer, baseURL)
	ch := handlers.NewContentHandler(contentStore)
	nh := handlers.NewNewsletterHandler(guestStore, configStore, mailer)
	cfgh := handlers.NewConfigHandler(configStore, themes)

	// Guest routes
	e.GET("/", gh.SpellPage)
	e.POST("/login", gh.Login)
	e.GET("/logout", gh.Logout)
	e.POST("/lang", gh.SetLang)
	e.GET("/unsubscribe", gh.Unsubscribe)

	me := e.Group("/me", mw.RequireGuest)
	me.GET("", gh.Me)
	me.GET("/rsvp", gh.RSVPForm)
	me.POST("/rsvp", gh.RSVPSubmit)
	me.GET("/confirmed", gh.Confirmed)
	me.GET("/calendar.ics", gh.CalendarICS)

	// Admin routes
	e.GET("/admin/login", ah.LoginPage)
	e.POST("/admin/login", ah.LoginSubmit)
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

	addr := ":" + port()
	log.Printf("listening on %s", addr)
	log.Fatal(e.Start(addr))
}

func port() string {
	if p := os.Getenv("PORT"); p != "" {
		return p
	}
	return "3000"
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
