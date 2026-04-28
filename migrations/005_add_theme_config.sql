-- +goose Up
INSERT INTO site_config (key, value, updated_at) VALUES
    ('theme_login', 'midnight-pool', datetime('now')),
    ('theme_me',    'midnight-pool', datetime('now'));

-- +goose Down
DELETE FROM site_config WHERE key IN ('theme_login', 'theme_me');
