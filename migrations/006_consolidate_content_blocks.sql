-- +goose Up
INSERT INTO content_blocks (key, label, body_de, body_en, updated_at)
SELECT 'general', 'Seiteninhalt', '', '', datetime('now')
WHERE NOT EXISTS (SELECT 1 FROM content_blocks WHERE key = 'general');

UPDATE content_blocks
SET
    body_de = (
        SELECT group_concat(part, char(10) || char(10))
        FROM (
            SELECT
                '## ' || cb.label || char(10) || char(10) || cb.body_de AS part,
                s.ord
            FROM (
                SELECT 'hero_tagline' AS key, 1 AS ord UNION ALL
                SELECT 'event_description', 2 UNION ALL
                SELECT 'what_to_expect', 3 UNION ALL
                SELECT 'kids_section', 4 UNION ALL
                SELECT 'special_thing', 5 UNION ALL
                SELECT 'charity_name', 6 UNION ALL
                SELECT 'charity_description', 7 UNION ALL
                SELECT 'footer_note', 8
            ) s
            JOIN content_blocks cb ON cb.key = s.key
            WHERE cb.body_de != ''
            ORDER BY s.ord
        )
    ),
    body_en = (
        SELECT group_concat(part, char(10) || char(10))
        FROM (
            SELECT
                '## ' || cb.label || char(10) || char(10) ||
                CASE
                    WHEN cb.body_en != '' THEN cb.body_en
                    ELSE cb.body_de
                END AS part,
                s.ord
            FROM (
                SELECT 'hero_tagline' AS key, 1 AS ord UNION ALL
                SELECT 'event_description', 2 UNION ALL
                SELECT 'what_to_expect', 3 UNION ALL
                SELECT 'kids_section', 4 UNION ALL
                SELECT 'special_thing', 5 UNION ALL
                SELECT 'charity_name', 6 UNION ALL
                SELECT 'charity_description', 7 UNION ALL
                SELECT 'footer_note', 8
            ) s
            JOIN content_blocks cb ON cb.key = s.key
            WHERE cb.body_de != '' OR cb.body_en != ''
            ORDER BY s.ord
        )
    ),
    updated_at = datetime('now')
WHERE key = 'general';

UPDATE content_blocks SET label = 'RSVP-Hinweis' WHERE key = 'rsvp_note';
UPDATE content_blocks SET label = 'Bestätigungsnachricht' WHERE key = 'confirmation_message';

DELETE FROM content_blocks WHERE key IN (
    'hero_tagline',
    'event_description',
    'what_to_expect',
    'kids_section',
    'charity_name',
    'charity_description',
    'special_thing',
    'footer_note'
);

-- +goose Down
INSERT INTO content_blocks (key, label, body_de, body_en, updated_at) VALUES
    ('hero_tagline',        'Tagline (Hero)',             '', '', datetime('now')),
    ('event_description',   'Event-Beschreibung',         '', '', datetime('now')),
    ('what_to_expect',      'Was erwartet euch',          '', '', datetime('now')),
    ('kids_section',        'Aktivitäten für Kinder',     '', '', datetime('now')),
    ('charity_name',        'Charity-Name',               '', '', datetime('now')),
    ('charity_description', 'Charity-Beschreibung',       '', '', datetime('now')),
    ('special_thing',       'Das Besondere dieses Jahr',  '', '', datetime('now')),
    ('footer_note',         'Fußzeile',                   '', '', datetime('now'));

DELETE FROM content_blocks WHERE key = 'general';
