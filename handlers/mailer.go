package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html/template"
	"io/fs"
	"net/smtp"
	"os"
	"strings"

	"party2026/models"
)

type Mailer struct {
	tmpl    *template.Template
	baseURL string
}

func NewMailer(tplFS fs.FS, baseURL string) (*Mailer, error) {
	funcMap := template.FuncMap{
		"deref": func(s *string) string {
			if s == nil {
				return ""
			}
			return *s
		},
	}
	tmpl, err := template.New("").Funcs(funcMap).ParseFS(tplFS, "email/*.html")
	if err != nil {
		return nil, err
	}
	return &Mailer{tmpl: tmpl, baseURL: baseURL}, nil
}

func (m *Mailer) SendConfirmation(g *models.Guest, cfg map[string]string, lang string) {
	if g.Email == nil || *g.Email == "" {
		return
	}

	subject := "Deine Anmeldung – Summer Party 2026"
	if lang == "en" {
		subject = "Your RSVP – Summer Party 2026"
	}

	var buf bytes.Buffer
	_ = m.tmpl.ExecuteTemplate(&buf, "confirmation.html", map[string]interface{}{
		"Guest":       g,
		"Lang":        lang,
		"Config":      cfg,
		"UnsubURL":    m.baseURL + "/unsubscribe?token=" + m.UnsubToken(g.ID),
	})

	_ = sendMail(*g.Email, subject, buf.String(), cfg)
}

func (m *Mailer) SendNewsletter(recipients []models.Guest, subject, bodyHTML string, cfg map[string]string) error {
	var errs []string
	for _, g := range recipients {
		if g.Email == nil || *g.Email == "" {
			continue
		}
		html := bodyHTML + fmt.Sprintf(
			`<p style="font-size:11px;color:#666">
			<a href="%s/unsubscribe?token=%s">Abmelden / Unsubscribe</a></p>`,
			m.baseURL, m.UnsubToken(g.ID),
		)
		if err := sendMail(*g.Email, subject, html, cfg); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", *g.Email, err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("send errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

func (m *Mailer) UnsubToken(guestID string) string {
	secret := os.Getenv("SESSION_SECRET")
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(guestID))
	return hex.EncodeToString(mac.Sum(nil)) + "." + guestID
}

func (m *Mailer) ValidateUnsubToken(token string) string {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return ""
	}
	expected := m.UnsubToken(parts[1])
	if !hmac.Equal([]byte(expected), []byte(token)) {
		return ""
	}
	return parts[1]
}

func sendMail(to, subject, bodyHTML string, cfg map[string]string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	from := os.Getenv("SMTP_FROM")
	fromName := cfg["smtp_from_name"]
	if fromName == "" {
		fromName = "Florian"
	}
	if host == "" || from == "" {
		return fmt.Errorf("SMTP not configured")
	}
	if port == "" {
		port = "587"
	}

	msg := fmt.Sprintf("From: %s <%s>\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		fromName, from, to, subject, bodyHTML)

	addr := host + ":" + port
	var auth smtp.Auth
	if user != "" {
		auth = smtp.PlainAuth("", user, pass, host)
	}
	return smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
}
