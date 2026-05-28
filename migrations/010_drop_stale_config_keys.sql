-- +goose Up
DELETE FROM site_config WHERE key IN (
    'charity_name',
    'charity_url',
    'admin_user',
    'admin_password_hash'
);

-- +goose Down
-- No-op: dropped rows are obsolete; cannot reconstruct meaningful values.
