package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"html/template"
	"io/fs"
	"net"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"party2026/models"
)

type Mailer struct {
	tmpl    *template.Template
	baseURL string
}

func NewMailer(tplFS fs.FS, baseURL string) (*Mailer, error) {
	funcMap := template.FuncMap{"deref": derefStringPtr}
	tmpl, err := template.New("").Funcs(funcMap).ParseFS(tplFS, "email/*.html")
	if err != nil {
		return nil, err
	}
	return &Mailer{tmpl: tmpl, baseURL: baseURL}, nil
}

func derefStringPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
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
		"Guest":    g,
		"Lang":     lang,
		"Config":   cfg,
		"BaseURL":  m.baseURL,
		"UnsubURL": m.baseURL + "/unsubscribe?token=" + m.UnsubToken(g.ID),
	})

	_ = sendMail(*g.Email, subject, buf.String(), cfg)
}

func (m *Mailer) SendNewsletter(recipients []models.Guest, subject, bodyMD string, cfg map[string]string) error {
	var errs []string
	for _, g := range recipients {
		if g.Email == nil || *g.Email == "" {
			continue
		}
		personalSubject := personalizeNewsletterText(subject, g)
		bodyHTML, err := renderNewsletterMarkdown(personalizeNewsletterText(bodyMD, g))
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", *g.Email, err))
			continue
		}
		html := bodyHTML + fmt.Sprintf(
			`<p style="font-size:11px;color:#666">
			<a href="%s/unsubscribe?token=%s">Abmelden / Unsubscribe</a></p>`,
			m.baseURL, m.UnsubToken(g.ID),
		)
		if err := sendMail(*g.Email, personalSubject, html, cfg); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", *g.Email, err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("send errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

func personalizeNewsletterText(text string, g models.Guest) string {
	text = strings.ReplaceAll(text, "{name}", g.Name)
	return text
}

func renderNewsletterMarkdown(bodyMD string) (string, error) {
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(bodyMD), &buf); err != nil {
		return bodyMD, err
	}
	return buf.String(), nil
}

// SendTestMail sends a simple HTML message to verify SMTP settings.
func SendTestMail(to string, cfg map[string]string) error {
	subject := "Party2026 SMTP test"
	body := fmt.Sprintf(
		`<p>This is a test email from Party2026.</p><p>Sent at %s.</p>`,
		time.Now().Format(time.RFC3339),
	)
	return sendMail(to, subject, body, cfg)
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

type smtpConfig struct {
	host     string
	port     string
	user     string
	pass     string
	from     string
	fromName string
	addr     string
	useSSL   bool
}

func loadSMTPConfig(cfg map[string]string) (smtpConfig, error) {
	c := smtpConfig{
		host: os.Getenv("SMTP_HOST"),
		port: os.Getenv("SMTP_PORT"),
		user: os.Getenv("SMTP_USER"),
		pass: os.Getenv("SMTP_PASS"),
		from: os.Getenv("SMTP_FROM"),
	}
	if c.port == "" {
		c.port = "587"
	}
	if c.from == "" || c.host == "" {
		return c, fmt.Errorf("SMTP not configured (need SMTP_HOST and SMTP_FROM)")
	}

	c.fromName = cfg["smtp_from_name"]
	if c.fromName == "" {
		c.fromName = os.Getenv("SMTP_FROM_NAME")
	}
	if c.fromName == "" {
		c.fromName = "Florian"
	}

	c.addr = net.JoinHostPort(c.host, c.port)
	c.useSSL = c.port == "465" || strings.EqualFold(os.Getenv("SMTP_TLS"), "ssl")
	return c, nil
}

func sendMail(to, subject, bodyHTML string, cfg map[string]string) error {
	smtpCfg, err := loadSMTPConfig(cfg)
	if err != nil {
		return err
	}

	msg := buildMailMessage(smtpCfg.fromName, smtpCfg.from, to, subject, bodyHTML)

	var auth smtp.Auth
	if smtpCfg.user != "" {
		auth = smtp.PlainAuth("", smtpCfg.user, smtpCfg.pass, smtpCfg.host)
	}

	if smtpCfg.useSSL {
		return sendMailImplicitTLS(smtpCfg.addr, smtpCfg.host, auth, smtpCfg.from, to, msg)
	}
	return sendMailSTARTTLS(smtpCfg.addr, smtpCfg.host, auth, smtpCfg.from, to, msg)
}

func buildMailMessage(fromName, from, to, subject, bodyHTML string) []byte {
	return []byte(fmt.Sprintf(
		"From: %s <%s>\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		fromName, from, to, subject, bodyHTML,
	))
}

func sendMailImplicitTLS(addr, host string, auth smtp.Auth, from, to string, msg []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host})
	if err != nil {
		return fmt.Errorf("tls dial: %w", err)
	}
	defer conn.Close()

	return deliverMail(conn, host, auth, from, to, msg)
}

func sendMailSTARTTLS(addr, host string, auth smtp.Auth, from, to string, msg []byte) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer conn.Close()

	return deliverMail(conn, host, auth, from, to, msg)
}

func deliverMail(conn net.Conn, host string, auth smtp.Auth, from, to string, msg []byte) error {
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Close()

	// STARTTLS when not already on TLS (port 465 connects via tls.Dial).
	if _, ok := conn.(*tls.Conn); !ok {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(&tls.Config{ServerName: host}); err != nil {
				return fmt.Errorf("starttls: %w", err)
			}
		}
	}

	if auth != nil {
		if ok, _ := client.Extension("AUTH"); ok {
			if err := client.Auth(auth); err != nil {
				return fmt.Errorf("auth: %w", err)
			}
		}
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("mail from: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("rcpt to: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("data: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("close data: %w", err)
	}
	return client.Quit()
}
