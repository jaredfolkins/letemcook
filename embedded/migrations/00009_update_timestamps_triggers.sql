-- Migration: 009_update_timestamps_triggers
-- Created at: $(date +%Y-%m-%d %H:%M:%S)
-- Purpose: Add triggers to automatically update the 'updated' timestamp on remaining tables when a row is updated.

-- +goose Up
-- +goose StatementBegin
-- Users table trigger
CREATE TRIGGER IF NOT EXISTS trg_users_after_update
AFTER UPDATE ON users
FOR EACH ROW
BEGIN
    UPDATE users SET updated = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

-- Configurations table trigger
CREATE TRIGGER IF NOT EXISTS trg_configurations_after_update
AFTER UPDATE ON configurations
FOR EACH ROW
BEGIN
    UPDATE configurations SET updated = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

-- Accounts table trigger
CREATE TRIGGER IF NOT EXISTS trg_accounts_after_update
AFTER UPDATE ON accounts
FOR EACH ROW
BEGIN
    UPDATE accounts SET updated = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

-- Permissions_system table trigger
CREATE TRIGGER IF NOT EXISTS trg_permissions_system_after_update
AFTER UPDATE ON permissions_system
FOR EACH ROW
BEGIN
    UPDATE permissions_system SET updated = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

-- Permissions_accounts table trigger
CREATE TRIGGER IF NOT EXISTS trg_permissions_accounts_after_update
AFTER UPDATE ON permissions_accounts
FOR EACH ROW
BEGIN
    UPDATE permissions_accounts SET updated = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

-- Permissions_apps table trigger
CREATE TRIGGER IF NOT EXISTS trg_permissions_apps_after_update
AFTER UPDATE ON permissions_apps
FOR EACH ROW
BEGIN
    UPDATE permissions_apps SET updated = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

-- Permissions_cookbooks table trigger
CREATE TRIGGER IF NOT EXISTS trg_permissions_cookbooks_after_update
AFTER UPDATE ON permissions_cookbooks
FOR EACH ROW
BEGIN
    UPDATE permissions_cookbooks SET updated = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

-- Apps table trigger
CREATE TRIGGER IF NOT EXISTS trg_apps_after_update
AFTER UPDATE ON apps
FOR EACH ROW
BEGIN
    UPDATE apps SET updated = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_users_after_update;
DROP TRIGGER IF EXISTS trg_configurations_after_update;
DROP TRIGGER IF EXISTS trg_accounts_after_update;
DROP TRIGGER IF EXISTS trg_permissions_system_after_update;
DROP TRIGGER IF EXISTS trg_permissions_accounts_after_update;
DROP TRIGGER IF EXISTS trg_permissions_apps_after_update;
DROP TRIGGER IF EXISTS trg_permissions_cookbooks_after_update;
DROP TRIGGER IF EXISTS trg_apps_after_update;
-- +goose StatementEnd 