-- +goose Up
INSERT INTO site_config (key, value, updated_at)
SELECT 'hero_image_url',
       'https://images.unsplash.com/photo-1530541930197-ff16ac917b0e?auto=format&fit=crop&w=1920&q=80',
       datetime('now')
WHERE NOT EXISTS (SELECT 1 FROM site_config WHERE key = 'hero_image_url');

-- +goose Down
DELETE FROM site_config WHERE key = 'hero_image_url';
