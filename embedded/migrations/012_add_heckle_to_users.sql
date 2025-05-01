-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN heckle BOOLEAN NOT NULL DEFAULT true;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- SQLite does not support dropping columns directly in older versions.
-- A common workaround is to create a new table, copy data, drop old, rename new.
-- However, for a simple boolean added, removing it might not be strictly necessary for rollback.
-- We'll provide the standard ALTER TABLE DROP COLUMN for compatibility, though it might fail on older SQLite.
ALTER TABLE users DROP COLUMN heckle;
-- +goose StatementEnd 