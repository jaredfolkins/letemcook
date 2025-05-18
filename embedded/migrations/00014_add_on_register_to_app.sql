-- +goose Up
-- +goose StatementBegin
ALTER TABLE apps ADD COLUMN on_register BOOLEAN DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Attempt to drop the column. This might fail on older SQLite versions.
ALTER TABLE apps DROP COLUMN on_register;
-- +goose StatementEnd 