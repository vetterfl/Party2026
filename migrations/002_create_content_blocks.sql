-- +goose Up
CREATE TABLE content_blocks (
    key        TEXT PRIMARY KEY,
    label      TEXT NOT NULL,
    body_de    TEXT NOT NULL DEFAULT '',
    body_en    TEXT NOT NULL DEFAULT '',
    updated_at DATETIME NOT NULL
);

INSERT INTO content_blocks (key, label, body_de, body_en, updated_at) VALUES
    ('hero_tagline',          'Tagline (Hero)',               '', '', datetime('now')),
    ('event_description',     'Event-Beschreibung',           '', '', datetime('now')),
    ('what_to_expect',        'Was erwartet euch',            '', '', datetime('now')),
    ('kids_section',          'Aktivitäten für Kinder',       'Maltisch, Plantschbecken und jede Menge Bälle!', 'Craft table, paddling pool and lots of balls!', datetime('now')),
    ('charity_name',          'Charity-Name',                 '', '', datetime('now')),
    ('charity_description',   'Charity-Beschreibung',         '', '', datetime('now')),
    ('special_thing',         'Das Besondere dieses Jahr',    '', '', datetime('now')),
    ('rsvp_note',             'Hinweis bei RSVP',             '', '', datetime('now')),
    ('confirmation_message',  'Bestätigungsnachricht',        '', '', datetime('now')),
    ('footer_note',           'Fußzeile',                     '', '', datetime('now'));

-- +goose Down
DROP TABLE content_blocks;
