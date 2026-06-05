package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/skip2/go-qrcode"
	"golang.org/x/crypto/bcrypt"
	"party2026/middleware"
	"party2026/models"
)

type AdminHandler struct {
	guests  *models.GuestStore
	config  *models.ConfigStore
	content *models.ContentStore
	admins  *models.AdminStore
	mailer  *Mailer
	baseURL string
}

type guestInviteRow struct {
	Guest         models.Guest
	InviteURL     string
	InviteMessage string
	WhatsAppURL   string
}

func NewAdminHandler(g *models.GuestStore, cfg *models.ConfigStore, cnt *models.ContentStore, admins *models.AdminStore, m *Mailer, baseURL string) *AdminHandler {
	return &AdminHandler{guests: g, config: cfg, content: cnt, admins: admins, mailer: m, baseURL: baseURL}
}

// GET /admin/login
func (h *AdminHandler) LoginPage(c echo.Context) error {
	if middleware.IsAdminAuthed(c) {
		return c.Redirect(http.StatusSeeOther, "/admin")
	}
	return c.Render(http.StatusOK, "login.html", map[string]interface{}{
		"Error": false,
	})
}

// dummyBcryptHash is a precomputed bcrypt hash of an unguessable password.
// Used so failed lookups still incur a bcrypt round, preventing username
// enumeration via response-time differences.
const dummyBcryptHash = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"

// POST /admin/login
func (h *AdminHandler) LoginSubmit(c echo.Context) error {
	user := c.FormValue("username")
	pass := c.FormValue("password")

	admin, err := h.admins.FindByUsername(user)
	hash := dummyBcryptHash
	if err == nil {
		hash = admin.PasswordHash
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass)) != nil || err != nil {
		return c.Render(http.StatusOK, "login.html", map[string]interface{}{
			"Error": true,
		})
	}

	_ = middleware.SetAdminAuthed(c, admin.ID)
	return c.Redirect(http.StatusSeeOther, "/admin")
}

// GET /admin/logout
func (h *AdminHandler) Logout(c echo.Context) error {
	_ = middleware.ClearAdminSession(c)
	return c.Redirect(http.StatusSeeOther, "/admin/login")
}

// GET /admin — dashboard
func (h *AdminHandler) Dashboard(c echo.Context) error {
	stats, err := h.guests.Stats()
	if err != nil {
		return err
	}
	cfg, err := h.config.All()
	if err != nil {
		return err
	}

	partyDate, _ := time.Parse("2006-01-02", cfg["party_date"])
	daysUntil := int(time.Until(partyDate).Hours() / 24)

	return c.Render(http.StatusOK, "dashboard.html", map[string]interface{}{
		"Stats":     stats,
		"DaysUntil": daysUntil,
		"Config":    cfg,
	})
}

// GET /admin/messages — RSVP comment message board
func (h *AdminHandler) MessageBoard(c echo.Context) error {
	guests, err := h.guests.WithComments()
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "messages.html", map[string]interface{}{
		"Guests": guests,
	})
}

// GET /admin/guests
func (h *AdminHandler) GuestList(c echo.Context) error {
	filter := c.QueryParam("status")
	var guests []models.Guest
	var err error
	switch filter {
	case "":
		guests, err = h.guests.All()
	case "no_response":
		guests, err = h.guests.ByNoResponse()
	default:
		guests, err = h.guests.ByStatus(filter)
	}
	if err != nil {
		return err
	}
	cfg, _ := h.config.All()
	rows := h.guestInviteRows(guests, cfg)
	return c.Render(http.StatusOK, "guests.html", map[string]interface{}{
		"Guests":          guests,
		"GuestInviteRows": rows,
		"Filter":          filter,
		"BaseURL":         h.baseURL,
		"Config":          cfg,
	})
}

// POST /admin/guests — create
func (h *AdminHandler) GuestCreate(c echo.Context) error {
	name := strings.TrimSpace(c.FormValue("name"))
	if name == "" {
		return c.Redirect(http.StatusSeeOther, "/admin/guests")
	}
	_, err := h.guests.Create(name)
	if err != nil {
		return err
	}
	return c.Redirect(http.StatusSeeOther, "/admin/guests")
}

// GET /admin/guests/:id/edit
func (h *AdminHandler) GuestEdit(c echo.Context) error {
	g, err := h.guests.FindByID(c.Param("id"))
	if err != nil {
		return echo.ErrNotFound
	}
	return c.Render(http.StatusOK, "guest_edit.html", map[string]interface{}{
		"Guest":   g,
		"BaseURL": h.baseURL,
	})
}

// POST /admin/guests/:id/edit
func (h *AdminHandler) GuestUpdate(c echo.Context) error {
	g, err := h.guests.FindByID(c.Param("id"))
	if err != nil {
		return echo.ErrNotFound
	}

	g.Name = strings.TrimSpace(c.FormValue("name"))
	g.Nickname = strings.TrimSpace(c.FormValue("nickname"))
	if g.Nickname == "" {
		g.Nickname = g.Name
	}
	if note := strings.TrimSpace(c.FormValue("internal_note")); note != "" {
		g.InternalNote = &note
	} else {
		g.InternalNote = nil
	}

	status := c.FormValue("status")
	if !models.ValidGuestStatus(status) {
		return c.Render(http.StatusBadRequest, "guest_edit.html", map[string]interface{}{
			"Guest":   g,
			"BaseURL": h.baseURL,
			"Error":   "Invalid status.",
		})
	}
	g.Status = status

	if email := strings.TrimSpace(c.FormValue("email")); email != "" {
		g.Email = &email
	} else {
		g.Email = nil
	}

	phone, err := normalizePhoneE164(c.FormValue("phone_e164"))
	if err != nil {
		return c.Render(http.StatusBadRequest, "guest_edit.html", map[string]interface{}{
			"Guest":   g,
			"BaseURL": h.baseURL,
			"Error":   err.Error(),
		})
	}
	g.PhoneE164 = phone

	g.PlusOne = c.FormValue("plus_one") == "1"
	if pname := strings.TrimSpace(c.FormValue("plus_one_name")); pname != "" {
		g.PlusOneName = &pname
	} else {
		g.PlusOneName = nil
	}

	children := 0
	fmt.Sscanf(c.FormValue("children"), "%d", &children)
	g.Children = children

	if song := strings.TrimSpace(c.FormValue("song")); song != "" {
		g.Song = &song
	} else {
		g.Song = nil
	}

	if err := h.guests.Update(g); err != nil {
		return err
	}
	return c.Redirect(http.StatusSeeOther, "/admin/guests")
}

// POST /admin/guests/:id/delete
func (h *AdminHandler) GuestDelete(c echo.Context) error {
	if err := h.guests.Delete(c.Param("id")); err != nil {
		return err
	}
	return c.Redirect(http.StatusSeeOther, "/admin/guests")
}

// GET /admin/guests/:id/qr — download QR PNG
func (h *AdminHandler) GuestQR(c echo.Context) error {
	g, err := h.guests.FindByID(c.Param("id"))
	if err != nil {
		return echo.ErrNotFound
	}
	url := h.baseURL + "/?spell=" + g.Code
	png, err := qrcode.Encode(url, qrcode.Medium, 256)
	if err != nil {
		return err
	}
	c.Response().Header().Set("Content-Type", "image/png")
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="invite-%s.png"`, g.Code))
	_, err = c.Response().Write(png)
	return err
}

// GET /admin/guests/:id/card — printable concert ticket HTML
func (h *AdminHandler) GuestCard(c echo.Context) error {
	g, err := h.guests.FindByID(c.Param("id"))
	if err != nil {
		return echo.ErrNotFound
	}
	url := h.baseURL + "/?spell=" + g.Code
	png, err := qrcode.Encode(url, qrcode.Medium, 200)
	if err != nil {
		return err
	}
	qrB64 := "data:image/png;base64," + base64.StdEncoding.EncodeToString(png)
	cfg, _ := h.config.All()
	return c.Render(http.StatusOK, "card.html", map[string]interface{}{
		"Guest":  g,
		"QRPNG":  qrB64,
		"Config": cfg,
		"URL":    h.baseURL,
	})
}

// GET /admin/export/csv
func (h *AdminHandler) ExportCSV(c echo.Context) error {
	guests, err := h.guests.All()
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"Name", "Nickname", "Code", "Status", "Email", "Phone", "Plus One", "Plus One Name", "Children", "Song", "Comment", "Internal Note", "Newsletter", "RSVP At", "Logins", "Views", "Interactions"})
	for _, g := range guests {
		rsvpAt := ""
		if g.RSVPAt != nil {
			rsvpAt = g.RSVPAt.Format(time.RFC3339)
		}
		_ = w.Write(csvSafeRow(
			g.Name, g.Nickname, g.Code, g.Status,
			derefStr(g.Email), derefStr(g.PhoneE164), boolStr(g.PlusOne), derefStr(g.PlusOneName),
			fmt.Sprint(g.Children), derefStr(g.Song), derefStr(g.Comment), derefStr(g.InternalNote),
			boolStr(g.Newsletter), rsvpAt,
			fmt.Sprint(g.LoginCount), fmt.Sprint(g.ViewCount), fmt.Sprint(g.InteractionCount),
		))
	}
	w.Flush()

	c.Response().Header().Set("Content-Type", "text/csv")
	c.Response().Header().Set("Content-Disposition", "attachment; filename=guestlist.csv")
	return c.String(http.StatusOK, buf.String())
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func boolStr(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

// csvSafeRow prefixes a single quote to cells starting with =, +, -, @, tab
// or carriage return. Mitigates spreadsheet formula injection when CSV is
// opened in Excel/Sheets/LibreOffice.
func csvSafeRow(cells ...string) []string {
	out := make([]string, len(cells))
	for i, c := range cells {
		out[i] = csvSafeCell(c)
	}
	return out
}

func csvSafeCell(s string) string {
	if s == "" {
		return s
	}
	switch s[0] {
	case '=', '+', '-', '@', '\t', '\r':
		return "'" + s
	}
	return s
}

func (h *AdminHandler) guestInviteRows(guests []models.Guest, cfg map[string]string) []guestInviteRow {
	rows := make([]guestInviteRow, 0, len(guests))
	for _, guest := range guests {
		inviteURL := h.baseURL + "/?spell=" + guest.Code
		message := inviteMessage(cfg, guest.DisplayName(), guest.Code, inviteURL)
		row := guestInviteRow{
			Guest:         guest,
			InviteURL:     inviteURL,
			InviteMessage: message,
		}
		if guest.PhoneE164 != nil && *guest.PhoneE164 != "" {
			phone := strings.TrimPrefix(*guest.PhoneE164, "+")
			row.WhatsAppURL = "https://wa.me/" + phone + "?text=" + url.QueryEscape(message)
		}
		rows = append(rows, row)
	}
	return rows
}

func inviteMessage(cfg map[string]string, name, spell, inviteURL string) string {
	tpl := cfg["invite_message_de"]
	if strings.TrimSpace(tpl) == "" {
		tpl = "Hi {name}, hier ist deine Einladung: {url} (Code: {spell})"
	}
	msg := strings.ReplaceAll(tpl, "{name}", name)
	msg = strings.ReplaceAll(msg, "{spell}", spell)
	msg = strings.ReplaceAll(msg, "{url}", inviteURL)
	return msg
}

func normalizePhoneE164(raw string) (*string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	var b strings.Builder
	for i := 0; i < len(raw); i++ {
		ch := raw[i]
		switch {
		case ch >= '0' && ch <= '9':
			b.WriteByte(ch)
		case ch == '+' && b.Len() == 0:
			b.WriteByte(ch)
		case ch == ' ' || ch == '-' || ch == '(' || ch == ')' || ch == '.':
			continue
		default:
			return nil, fmt.Errorf("phone must be E.164, e.g. +491701234567")
		}
	}

	phone := b.String()
	if strings.HasPrefix(phone, "00") {
		phone = "+" + strings.TrimPrefix(phone, "00")
	}
	if !strings.HasPrefix(phone, "+") {
		return nil, fmt.Errorf("phone must include country code, e.g. +491701234567")
	}

	digits := phone[1:]
	if len(digits) < 8 || len(digits) > 15 || digits[0] == '0' {
		return nil, fmt.Errorf("phone must be E.164, e.g. +491701234567")
	}
	for i := 0; i < len(digits); i++ {
		if digits[i] < '0' || digits[i] > '9' {
			return nil, fmt.Errorf("phone must be E.164, e.g. +491701234567")
		}
	}
	return &phone, nil
}
