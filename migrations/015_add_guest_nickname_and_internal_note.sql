-- +goose Up
ALTER TABLE guests ADD COLUMN nickname TEXT NOT NULL DEFAULT '';
ALTER TABLE guests ADD COLUMN internal_note TEXT;

UPDATE guests SET nickname = name WHERE trim(nickname) = '';

-- +goose Down
ALTER TABLE guests DROP COLUMN nickname;
ALTER TABLE guests DROP COLUMN internal_note;
