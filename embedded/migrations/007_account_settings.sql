-- +goose Up
-- +goose StatementBegin
PRAGMA foreign_keys = ON;

CREATE TABLE account_settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id INTEGER NOT NULL UNIQUE REFERENCES accounts(id) ON DELETE CASCADE,
    theme TEXT NOT NULL DEFAULT 'default',
    registration BOOLEAN NOT NULL DEFAULT false, -- Use 1 for TRUE in SQLite
    heckle BOOLEAN NOT NULL DEFAULT false, -- Use 0 for FALSE in SQLite
    created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Trigger to update updated_at on row update
-- Note: SQLite trigger syntax is slightly different
CREATE TRIGGER trg_account_settings_after_update
AFTER UPDATE ON account_settings
FOR EACH ROW
BEGIN
    UPDATE account_settings SET updated = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_account_settings_after_update;
DROP TABLE IF EXISTS account_settings;
-- +goose StatementEnd 