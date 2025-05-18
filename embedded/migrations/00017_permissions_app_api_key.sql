-- +goose Up
-- +goose StatementBegin
ALTER TABLE permissions_apps ADD COLUMN api_key TEXT;
UPDATE permissions_apps SET api_key = lower(hex(randomblob(16))) WHERE api_key IS NULL OR api_key = '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_permissions_apps_api_key ON permissions_apps(api_key);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_permissions_apps_api_key;
ALTER TABLE permissions_apps DROP COLUMN api_key;
-- +goose StatementEnd
