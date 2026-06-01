-- +goose Up
ALTER TABLE guests ADD COLUMN login_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE guests ADD COLUMN view_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE guests ADD COLUMN interaction_count INTEGER NOT NULL DEFAULT 0;

INSERT OR IGNORE INTO site_config (key, value, updated_at)
VALUES ('rsvp_notify_email', '', datetime('now'));

-- +goose Down
ALTER TABLE guests DROP COLUMN login_count;
ALTER TABLE guests DROP COLUMN view_count;
ALTER TABLE guests DROP COLUMN interaction_count;

DELETE FROM site_config WHERE key = 'rsvp_notify_email';
