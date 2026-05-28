-- +goose Up
CREATE TABLE admins (
    id            TEXT PRIMARY KEY,
    username      TEXT NOT NULL UNIQUE COLLATE NOCASE,
    password_hash TEXT NOT NULL,
    created_at    DATETIME NOT NULL,
    updated_at    DATETIME NOT NULL
);

INSERT INTO admins (id, username, password_hash, created_at, updated_at)
SELECT
    lower(hex(randomblob(16))),
    (SELECT value FROM site_config WHERE key = 'admin_user'),
    (SELECT value FROM site_config WHERE key = 'admin_password_hash'),
    datetime('now'),
    datetime('now')
WHERE EXISTS (
    SELECT 1 FROM site_config
    WHERE key = 'admin_user' AND trim(value) != ''
)
AND EXISTS (
    SELECT 1 FROM site_config
    WHERE key = 'admin_password_hash' AND trim(value) != ''
);

-- +goose Down
DROP TABLE admins;
