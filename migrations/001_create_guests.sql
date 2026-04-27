-- +goose Up
CREATE TABLE guests (
    id            TEXT PRIMARY KEY,
    code          TEXT NOT NULL UNIQUE,
    name          TEXT NOT NULL,
    status        TEXT NOT NULL DEFAULT 'invited',
    email         TEXT,
    plus_one      INTEGER NOT NULL DEFAULT 0,
    plus_one_name TEXT,
    children      INTEGER NOT NULL DEFAULT 0,
    song          TEXT,
    comment       TEXT,
    newsletter    INTEGER NOT NULL DEFAULT 0,
    rsvp_at       DATETIME,
    created_at    DATETIME NOT NULL,
    updated_at    DATETIME NOT NULL
);

-- +goose Down
DROP TABLE guests;
