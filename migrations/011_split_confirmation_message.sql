-- +goose Up
INSERT INTO content_blocks (key, label, body_de, body_en, updated_at)
SELECT 'confirmation_message_accepted',
       'Bestätigung: Zusage',
       COALESCE((SELECT body_de FROM content_blocks WHERE key = 'confirmation_message'), ''),
       COALESCE((SELECT body_en FROM content_blocks WHERE key = 'confirmation_message'), ''),
       datetime('now')
WHERE NOT EXISTS (SELECT 1 FROM content_blocks WHERE key = 'confirmation_message_accepted');

INSERT INTO content_blocks (key, label, body_de, body_en, updated_at)
SELECT 'confirmation_message_declined',
       'Bestätigung: Absage',
       '',
       '',
       datetime('now')
WHERE NOT EXISTS (SELECT 1 FROM content_blocks WHERE key = 'confirmation_message_declined');

DELETE FROM content_blocks WHERE key = 'confirmation_message';

-- +goose Down
INSERT INTO content_blocks (key, label, body_de, body_en, updated_at)
SELECT 'confirmation_message',
       'Bestätigungsnachricht',
       COALESCE((SELECT body_de FROM content_blocks WHERE key = 'confirmation_message_accepted'), ''),
       COALESCE((SELECT body_en FROM content_blocks WHERE key = 'confirmation_message_accepted'), ''),
       datetime('now')
WHERE NOT EXISTS (SELECT 1 FROM content_blocks WHERE key = 'confirmation_message');

DELETE FROM content_blocks WHERE key IN (
    'confirmation_message_accepted',
    'confirmation_message_declined'
);
