-- +goose Up
-- +goose StatementBegin
ALTER TABLE apps ADD COLUMN is_mcp_enabled BOOLEAN NOT NULL DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE apps DROP COLUMN is_mcp_enabled;
-- +goose StatementEnd
