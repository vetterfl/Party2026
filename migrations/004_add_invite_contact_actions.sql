-- +goose Up
ALTER TABLE guests ADD COLUMN phone_e164 TEXT;

INSERT INTO site_config (key, value, updated_at) VALUES
    ('invite_message_de', 'Hi {name}, hier ist deine Einladung: {url}', datetime('now')),
    ('invite_message_en', 'Hi {name}, here is your invitation: {url}', datetime('now'))
ON CONFLICT(key) DO NOTHING;

-- +goose Down
DELETE FROM site_config WHERE key IN ('invite_message_de', 'invite_message_en');
