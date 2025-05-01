-- Migration: 008_update_cookbook_timestamp
-- Created at: $(date +%Y-%m-%d %H:%M:%S)
-- Purpose: Add a trigger to automatically update the 'updated' timestamp on the 'cookbooks' table when a row is updated.

-- +goose Up
-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS trg_cookbooks_after_update
AFTER UPDATE ON cookbooks
FOR EACH ROW
BEGIN
    UPDATE cookbooks SET updated = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_cookbooks_after_update;
-- +goose StatementEnd 