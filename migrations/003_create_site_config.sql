-- +goose Up
CREATE TABLE site_config (
    key        TEXT PRIMARY KEY,
    value      TEXT NOT NULL DEFAULT '',
    updated_at DATETIME NOT NULL
);

INSERT INTO site_config (key, value, updated_at) VALUES
    ('party_date',        '2026-08-01',   datetime('now')),
    ('party_time_start',  '16:00',        datetime('now')),
    ('party_name_de',     'Summer Party', datetime('now')),
    ('party_name_en',     'Summer Party', datetime('now')),
    ('rsvp_deadline',     '2026-07-15',   datetime('now')),
    ('charity_name',      '',             datetime('now')),
    ('charity_url',       '',             datetime('now')),
    ('smtp_from_name',    'Florian',      datetime('now')),
    ('admin_user',        'florian',      datetime('now')),
    ('admin_password_hash', '',           datetime('now'));

-- +goose Down
DROP TABLE site_config;
