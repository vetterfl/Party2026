-- +goose Up
INSERT INTO content_blocks (key, label, body_de, body_en, updated_at)
SELECT 'gallery_intro', 'Foto-Galerie Untertitel',
       'Erinnerungen an einen schönen Tag. Tippe auf ein Bild, um es groß zu sehen.',
       'Memories from a lovely day. Tap a picture to view it large.',
       datetime('now')
WHERE NOT EXISTS (SELECT 1 FROM content_blocks WHERE key = 'gallery_intro');

-- +goose Down
DELETE FROM content_blocks WHERE key = 'gallery_intro';
