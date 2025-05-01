-- Migration: 011_optimize_indexes
-- Purpose: Add missing indexes and remove redundant ones to optimize database performance

-- +goose Up
-- +goose StatementBegin
PRAGMA foreign_keys = ON;

-- Add missing indexes for foreign keys
CREATE INDEX IF NOT EXISTS idx_cookbooks_account_id ON cookbooks(account_id);
CREATE INDEX IF NOT EXISTS idx_cookbooks_owner_id ON cookbooks(owner_id);

CREATE INDEX IF NOT EXISTS idx_apps_account_id ON apps(account_id);
CREATE INDEX IF NOT EXISTS idx_apps_owner_id ON apps(owner_id);
CREATE INDEX IF NOT EXISTS idx_apps_cookbook_id ON apps(cookbook_id);

CREATE INDEX IF NOT EXISTS idx_permissions_apps_app_id ON permissions_apps(app_id);
CREATE INDEX IF NOT EXISTS idx_permissions_apps_cookbook_id ON permissions_apps(cookbook_id);

CREATE INDEX IF NOT EXISTS idx_permissions_cookbooks_cookbook_id ON permissions_cookbooks(cookbook_id);

-- Add composite indexes for unique constraints
-- Note: SQLite may already use the unique constraint as an index, but explicit indexes can help with query planning
CREATE INDEX IF NOT EXISTS idx_cookbooks_account_name ON cookbooks(account_id, name);
CREATE INDEX IF NOT EXISTS idx_apps_account_name ON apps(account_id, name);

-- Remove redundant index (primary key is already indexed)
DROP INDEX IF EXISTS configurations_id_idx;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Restore removed index
CREATE INDEX configurations_id_idx ON configurations(id);

-- Remove added indexes
DROP INDEX IF EXISTS idx_cookbooks_account_id;
DROP INDEX IF EXISTS idx_cookbooks_owner_id;
DROP INDEX IF EXISTS idx_apps_account_id;
DROP INDEX IF EXISTS idx_apps_owner_id;
DROP INDEX IF EXISTS idx_apps_cookbook_id;
DROP INDEX IF EXISTS idx_permissions_apps_app_id;
DROP INDEX IF EXISTS idx_permissions_apps_cookbook_id;
DROP INDEX IF EXISTS idx_permissions_cookbooks_cookbook_id;
DROP INDEX IF EXISTS idx_cookbooks_account_name;
DROP INDEX IF EXISTS idx_apps_account_name;
-- +goose StatementEnd 