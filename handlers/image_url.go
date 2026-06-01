package handlers

import (
	"net/url"
	"strings"
)

func sanitizeImageURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return ""
	}
	return u.String()
}

func cssImageURL(raw string) string {
	u := sanitizeImageURL(raw)
	if u == "" {
		return ""
	}
	u = strings.ReplaceAll(u, `\`, `\\`)
	u = strings.ReplaceAll(u, `'`, `\'`)
	return u
}
