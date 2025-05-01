-- +goose Up
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_user_settings_after_update;
DROP INDEX IF EXISTS idx_user_settings_user_id;
DROP TABLE IF EXISTS user_settings;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Recreate the table, index, and trigger if rolling back
PRAGMA foreign_keys = ON;

CREATE TABLE user_settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    heckle BOOLEAN NOT NULL DEFAULT false,
    created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_user_settings_after_update
AFTER UPDATE ON user_settings
FOR EACH ROW
BEGIN
    UPDATE user_settings SET updated = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

CREATE INDEX idx_user_settings_user_id ON user_settings(user_id);
-- +goose StatementEnd 