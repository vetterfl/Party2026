package handlers

import (
	"bytes"
	"html/template"
	"io"
	"io/fs"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/yuin/goldmark"
	"party2026/locales"
	"party2026/models"
)

type Renderer struct {
	templates *template.Template
}

func NewRenderer(tplFS fs.FS) (*Renderer, error) {
	funcMap := template.FuncMap{
		"t": func(lang, key string) string {
			return locales.T(lang, key)
		},
		"content": func(blocks map[string]models.ContentBlock, key, lang string) template.HTML {
			b, ok := blocks[key]
			if !ok {
				return ""
			}
			body := b.BodyDE
			if lang == "en" && b.BodyEN != "" {
				body = b.BodyEN
			}
			if body == "" {
				return ""
			}
			var buf bytes.Buffer
			if err := goldmark.Convert([]byte(body), &buf); err != nil {
				return template.HTML(template.HTMLEscapeString(body))
			}
			return template.HTML(buf.String())
		},
		"cfg": func(m map[string]string, key string) string {
			return m[key]
		},
		"md": func(s string) template.HTML {
			var buf bytes.Buffer
			if err := goldmark.Convert([]byte(s), &buf); err != nil {
				return template.HTML(template.HTMLEscapeString(s))
			}
			return template.HTML(buf.String())
		},
		"deref": func(s *string) string {
			if s == nil {
				return ""
			}
			return *s
		},
		"isNil": func(v *string) bool { return v == nil },
		"formatDate": func(t time.Time) string {
			return t.Format("02.01.2006 15:04")
		},
		"formatDatePtr": func(t *time.Time) string {
			if t == nil {
				return "—"
			}
			return t.Format("02.01.2006 15:04")
		},
		"statusLabel": func(lang, status string) string {
			labels := map[string]map[string]string{
				"de": {
					"added":     "Hinzugefügt",
					"invited":   "Eingeladen",
					"accepted":  "Zugesagt",
					"declined":  "Abgesagt",
					"tentative": "Vielleicht",
				},
				"en": {
					"added":     "Added",
					"invited":   "Invited",
					"accepted":  "Accepted",
					"declined":  "Declined",
					"tentative": "Tentative",
				},
			}
			if l, ok := labels[lang]; ok {
				if v, ok := l[status]; ok {
					return v
				}
			}
			return status
		},
		"statusClass": func(status string) string {
			switch status {
			case "added":
				return "status-added"
			case "accepted":
				return "status-accepted"
			case "declined":
				return "status-declined"
			case "tentative":
				return "status-tentative"
			default:
				return "status-invited"
			}
		},
		"add":   func(a, b int) int { return a + b },
		"lower": strings.ToLower,
		"boolCheck": func(b bool) string {
			if b {
				return "✓"
			}
			return "–"
		},
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFS(tplFS,
		"*.html",
		"me/*.html",
		"admin/*.html",
		"email/*.html",
	)
	if err != nil {
		return nil, err
	}
	return &Renderer{templates: tmpl}, nil
}

// Render implements echo.Renderer
func (r *Renderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return r.templates.ExecuteTemplate(w, name, data)
}

// BaseData is passed to all guest-facing pages
type BaseData struct {
	Lang    string
	Guest   *models.Guest
	Config  map[string]string
	Content map[string]models.ContentBlock
	Theme   string
}

func newBase(c echo.Context, cfg map[string]string, content map[string]models.ContentBlock, g *models.Guest, theme string) BaseData {
	return BaseData{
		Lang:    getLang(c),
		Guest:   g,
		Config:  cfg,
		Content: content,
		Theme:   theme,
	}
}

func getLang(c echo.Context) string {
	cookie, err := c.Cookie("lang")
	if err == nil && (cookie.Value == "en" || cookie.Value == "de") {
		return cookie.Value
	}
	return "de"
}
