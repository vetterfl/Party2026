package handlers

import "strings"

// featureEnabled reports whether an optional feature toggle is on. Toggles
// default to ON when unset — the features shipped before the config switch
// existed, so a missing key must not hide them. Only an explicit "0"/"false"
// disables.
func featureEnabled(cfg map[string]string, key string) bool {
	v, ok := cfg[key]
	if !ok || v == "" {
		return true
	}
	return !(v == "0" || strings.EqualFold(v, "false"))
}

func galleryEnabled(cfg map[string]string) bool { return featureEnabled(cfg, "gallery_enabled") }
func carpoolEnabled(cfg map[string]string) bool { return featureEnabled(cfg, "carpool_enabled") }
