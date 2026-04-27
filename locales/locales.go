package locales

var DE = map[string]string{
	"spell.placeholder":    "Dein Zauberspruch…",
	"spell.submit":         "Eintreten",
	"spell.error":          "Das war kein Zauber.",
	"spell.hint":           "Du hast eine persönliche Einladung erhalten.",
	"nav.logout":           "Abmelden",
	"nav.lang_en":          "EN",
	"nav.lang_de":          "DE",
	"rsvp.title":           "Deine Anmeldung",
	"rsvp.attending_yes":   "Ja, ich komme!",
	"rsvp.attending_no":    "Leider nicht",
	"rsvp.plus_one":        "Bring ich jemanden mit?",
	"rsvp.plus_one_yes":    "Ja",
	"rsvp.plus_one_no":     "Nein",
	"rsvp.plus_one_name":   "Name der Begleitung (optional)",
	"rsvp.children":        "Anzahl Kinder",
	"rsvp.song":            "Mein Lieblingslied (optional)",
	"rsvp.song_hint":       "Für die Playlist 🎵",
	"rsvp.comment":         "Nachricht an Florian (optional)",
	"rsvp.email":           "E-Mail (optional)",
	"rsvp.email_hint":      "Für Updates und Infos",
	"rsvp.newsletter":      "Ja, ich möchte Updates erhalten",
	"rsvp.submit":          "Absenden",
	"rsvp.update":          "Anmeldung ändern",
	"confirmed.title":      "Super!",
	"confirmed.accepted":   "Wir sehen uns am 1. August!",
	"confirmed.declined":   "Schade, vielleicht beim nächsten Mal.",
	"confirmed.change":     "Anmeldung ändern",
	"confirmed.song":       "Dein Lied:",
	"confirmed.plus_one":   "Mit Begleitung:",
	"confirmed.children":   "Kinder:",
	"charity.cta":          "Spenden statt Geschenke",
	"date.label":           "1. August 2026 · ab 16 Uhr",
}

var EN = map[string]string{
	"spell.placeholder":    "Your spell…",
	"spell.submit":         "Enter",
	"spell.error":          "That was no spell.",
	"spell.hint":           "You received a personal invitation.",
	"nav.logout":           "Logout",
	"nav.lang_en":          "EN",
	"nav.lang_de":          "DE",
	"rsvp.title":           "Your RSVP",
	"rsvp.attending_yes":   "Yes, I'm coming!",
	"rsvp.attending_no":    "Can't make it",
	"rsvp.plus_one":        "Bringing someone?",
	"rsvp.plus_one_yes":    "Yes",
	"rsvp.plus_one_no":     "No",
	"rsvp.plus_one_name":   "Guest name (optional)",
	"rsvp.children":        "Number of kids",
	"rsvp.song":            "Favourite song (optional)",
	"rsvp.song_hint":       "For the playlist 🎵",
	"rsvp.comment":         "Message for Florian (optional)",
	"rsvp.email":           "Email (optional)",
	"rsvp.email_hint":      "For updates and info",
	"rsvp.newsletter":      "Yes, I'd like to receive updates",
	"rsvp.submit":          "Send",
	"rsvp.update":          "Change RSVP",
	"confirmed.title":      "You're in!",
	"confirmed.accepted":   "See you on August 1st!",
	"confirmed.declined":   "Sorry to miss you — maybe next time.",
	"confirmed.change":     "Change RSVP",
	"confirmed.song":       "Your song:",
	"confirmed.plus_one":   "Plus one:",
	"confirmed.children":   "Kids:",
	"charity.cta":          "Donate instead of gifts",
	"date.label":           "August 1, 2026 · from 4 pm",
}

func T(lang string, key string) string {
	m := DE
	if lang == "en" {
		m = EN
	}
	if v, ok := m[key]; ok {
		return v
	}
	return key
}
