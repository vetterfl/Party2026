-- +goose Up
INSERT INTO site_config (key, value, updated_at) VALUES
    ('calendar_enabled',        '0', datetime('now')),
    ('calendar_time_end',       '',  datetime('now')),
    ('calendar_location',       '',  datetime('now')),
    ('calendar_description_de', '',  datetime('now')),
    ('calendar_description_en', '',  datetime('now'))
ON CONFLICT(key) DO NOTHING;

-- +goose Down
DELETE FROM site_config WHERE key IN (
    'calendar_enabled',
    'calendar_time_end',
    'calendar_location',
    'calendar_description_de',
    'calendar_description_en'
);
