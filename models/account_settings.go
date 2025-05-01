package models

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/jaredfolkins/letemcook/db"
	"github.com/jmoiron/sqlx"
)

type AccountSettings struct {
	ID           int64
	AccountID    int64
	Theme        string
	Registration bool
	Heckle       bool
	Created      time.Time
	Updated      time.Time
}

func GetAccountSettingsByAccountID(db *sql.DB, accountID int64) (*AccountSettings, error) {
	query := `SELECT id, account_id, theme, registration, heckle, created, updated
              FROM account_settings WHERE account_id = ?`
	row := db.QueryRow(query, accountID)

	var settings AccountSettings

	err := row.Scan(
		&settings.ID,
		&settings.AccountID,
		&settings.Theme,
		&settings.Registration,
		&settings.Heckle,
		&settings.Created,
		&settings.Updated,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No settings found for account %d, defaulting to disabled", accountID)
			settings.Registration = false
		}
		return nil, err
	}

	return &settings, nil
}

func UpsertAccountSettings(db *sqlx.DB, settings *AccountSettings) error {

	query := `
        INSERT INTO account_settings (account_id, theme, registration, heckle)
        VALUES (?, ?, ?, ?)
        ON CONFLICT(account_id) DO UPDATE SET
        theme = excluded.theme,
        registration = excluded.registration,
        heckle = excluded.heckle,
        updated = CURRENT_TIMESTAMP;`

	_, err := db.Exec(query, settings.AccountID, settings.Theme, settings.Registration, settings.Heckle)
	if err != nil {
		log.Printf("Failed to upsert account settings for account %d: %v", settings.AccountID, err)
		return err
	}
	return nil
}

// AccountSettingsByAccountID fetches the registration setting for a given account ID.
// It returns true if registration is enabled, false otherwise (including if no settings row exists).
// An error is returned only for actual database query issues.
func AccountSettingsByAccountID(accountID int64) (bool, error) {
	var registration bool
	query := `SELECT registration FROM account_settings WHERE account_id = ?`
	err := db.Db().Get(&registration, query, accountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// If no settings row exists for the account, assume registration is disabled
			log.Printf("No account settings found for account_id %d, defaulting registration to disabled.", accountID)
			return false, nil // Not an error, just default behavior
		}
		log.Printf("Error fetching account settings for account_id %d: %v", accountID, err)
		return false, err // Return the actual database error
	}
	return registration, nil
}

// Save inserts or updates the account settings in the database.
func (s *AccountSettings) Save() error {
	return UpsertAccountSettings(db.Db(), s)
}
