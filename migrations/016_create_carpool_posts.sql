-- +goose Up
CREATE TABLE carpool_posts (
    id          TEXT PRIMARY KEY,
    guest_id    TEXT NOT NULL REFERENCES guests(id) ON DELETE CASCADE,
    kind        TEXT NOT NULL DEFAULT 'offer',
    origin      TEXT NOT NULL DEFAULT '',
    travel_time TEXT NOT NULL DEFAULT '',
    seats       INTEGER NOT NULL DEFAULT 0,
    note        TEXT NOT NULL DEFAULT '',
    contact     TEXT NOT NULL DEFAULT '',
    created_at  DATETIME NOT NULL,
    updated_at  DATETIME NOT NULL
);

CREATE INDEX idx_carpool_posts_guest ON carpool_posts(guest_id);

-- +goose Down
DROP TABLE carpool_posts;
