-- +goose Up
INSERT INTO content_blocks (key, label, body_de, body_en, updated_at)
SELECT 'footer_note', 'Fußzeile', 'Erstellt mit Liebe von Florian', 'Made with love by Florian', datetime('now')
WHERE NOT EXISTS (SELECT 1 FROM content_blocks WHERE key = 'footer_note');

INSERT INTO content_blocks (key, label, body_de, body_en, updated_at)
SELECT 'privacy', 'Datenschutzerklärung',
       'Wir nehmen den Schutz deiner Daten ernst. Diese Seite beschreibt, welche Daten wir erheben und wie wir sie verwenden.

### Welche Daten wir speichern

- Name und persönlicher Einladungscode
- RSVP-Angaben (Zusage/Absage, Begleitung, Kinder, Songwunsch, Nachricht)
- Optional: E-Mail-Adresse und Newsletter-Einwilligung

### Wofür wir die Daten nutzen

- Organisation der Veranstaltung
- Kommunikation rund um die Party (nur bei E-Mail-Angabe)
- Newsletter (nur bei ausdrücklicher Einwilligung)

### Deine Rechte

Du kannst deine Angaben jederzeit auf der Anmeldeseite ändern. Newsletter-Empfänger können sich über den Abmeldelink in jeder E-Mail abmelden.

Bei Fragen wende dich an den Gastgeber.',
       'We take your data seriously. This page describes what we collect and how we use it.

### What we store

- Name and personal invitation code
- RSVP details (attendance, plus-one, children, song request, message)
- Optional: email address and newsletter consent

### How we use your data

- Event organisation
- Party-related communication (only if you provide an email)
- Newsletter (only with explicit consent)

### Your rights

You can update your details on the RSVP page at any time. Newsletter recipients can unsubscribe via the link in every email.

Questions? Contact the host.',
       datetime('now')
WHERE NOT EXISTS (SELECT 1 FROM content_blocks WHERE key = 'privacy');

-- +goose Down
DELETE FROM content_blocks WHERE key IN ('footer_note', 'privacy');
