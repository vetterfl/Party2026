-- +goose Up
UPDATE guests SET status = 'invited', updated_at = datetime('now')
WHERE status = 'no_response';

-- +goose Down
-- Cannot restore which rows were no_response vs invited.
