package handlers

import (
	"fmt"
	"strings"
	"time"
)

const calendarTZ = "Europe/Berlin"

func calendarEnabled(cfg map[string]string) bool {
	return cfg["calendar_enabled"] == "1" || strings.EqualFold(cfg["calendar_enabled"], "true")
}

func buildCalendarICS(cfg map[string]string, lang, guestID string) ([]byte, error) {
	if !calendarEnabled(cfg) {
		return nil, fmt.Errorf("calendar download disabled")
	}

	loc, err := time.LoadLocation(calendarTZ)
	if err != nil {
		return nil, fmt.Errorf("timezone: %w", err)
	}

	start, err := parsePartyStart(cfg, loc)
	if err != nil {
		return nil, err
	}

	end, err := parsePartyEnd(cfg, start, loc)
	if err != nil {
		return nil, err
	}

	summary := cfg["party_name_de"]
	if lang == "en" && cfg["party_name_en"] != "" {
		summary = cfg["party_name_en"]
	}
	if summary == "" {
		summary = "Summer Party"
	}

	description := cfg["calendar_description_de"]
	if lang == "en" && cfg["calendar_description_en"] != "" {
		description = cfg["calendar_description_en"]
	} else if lang == "en" && cfg["calendar_description_en"] == "" && cfg["calendar_description_de"] != "" {
		description = cfg["calendar_description_de"]
	}

	now := time.Now().UTC()
	lines := []string{
		"BEGIN:VCALENDAR",
		"VERSION:2.0",
		"PRODID:-//Party2026//EN",
		"CALSCALE:GREGORIAN",
		"METHOD:PUBLISH",
		"BEGIN:VEVENT",
		"UID:" + icsEscape(guestID+"@party2026"),
		"DTSTAMP:" + now.Format("20060102T150405Z"),
		"DTSTART;TZID=" + calendarTZ + ":" + start.Format("20060102T150405"),
		"DTEND;TZID=" + calendarTZ + ":" + end.Format("20060102T150405"),
		"SUMMARY:" + icsEscape(summary),
	}
	if description != "" {
		lines = append(lines, "DESCRIPTION:"+icsEscape(description))
	}
	if locName := strings.TrimSpace(cfg["calendar_location"]); locName != "" {
		lines = append(lines, "LOCATION:"+icsEscape(locName))
	}
	lines = append(lines, "END:VEVENT", "END:VCALENDAR")

	return []byte(strings.Join(lines, "\r\n") + "\r\n"), nil
}

func parsePartyStart(cfg map[string]string, loc *time.Location) (time.Time, error) {
	date := strings.TrimSpace(cfg["party_date"])
	timeStart := strings.TrimSpace(cfg["party_time_start"])
	if date == "" || timeStart == "" {
		return time.Time{}, fmt.Errorf("party date and start time must be configured")
	}
	start, err := time.ParseInLocation("2006-01-02 15:04", date+" "+timeStart, loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid party date/time: %w", err)
	}
	return start, nil
}

func parsePartyEnd(cfg map[string]string, start time.Time, loc *time.Location) (time.Time, error) {
	endTime := strings.TrimSpace(cfg["calendar_time_end"])
	if endTime == "" {
		return start.Add(4 * time.Hour), nil
	}
	end, err := time.ParseInLocation("2006-01-02 15:04", start.Format("2006-01-02")+" "+endTime, loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid calendar end time: %w", err)
	}
	if !end.After(start) {
		return time.Time{}, fmt.Errorf("calendar end time must be after start time")
	}
	return end, nil
}

func icsEscape(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, "\r\n", `\n`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, ";", `\;`)
	s = strings.ReplaceAll(s, ",", `\,`)
	return s
}
