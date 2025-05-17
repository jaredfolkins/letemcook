package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jaredfolkins/letemcook/db"
	"github.com/jmoiron/sqlx"
)

type AppView struct {
	ID    int64  `json:"id"`    // Unique identifier for the App.
	UUID  string `json:"uuid"`  // Unique identifier (UUID) for the App.
	Title string `json:"title"` // The display title or name of the App.
}

type App struct {
	Created             time.Time `db:"created" json:"created"`
	Updated             time.Time `db:"updated" json:"updated"`
	ID                  int64     `db:"id" json:"id"`
	AccountID           int64     `db:"account_id" json:"account_id"`
	OwnerID             int64     `db:"owner_id" json:"owner_id"`
	CookbookID          int64     `db:"cookbook_id" json:"cookbook_id"`
	UUID                string    `db:"uuid" json:"uuid"`
	Name                string    `db:"name" json:"name"`
	Description         string    `db:"description" json:"description"`
	YAMLShared          string    `db:"yaml_shared" json:"yaml_shared"`
	YAMLIndividual      string    `db:"yaml_individual" json:"yaml_individual"`
	ApiKey              string    `db:"api_key" json:"api_key,omitempty"`
	IsActive            bool      `db:"is_active" json:"is_active"`
	IsDeleted           bool      `db:"is_deleted" json:"is_deleted"`
	IsAssignedByDefault bool      `db:"is_assigned_by_default" json:"is_assigned_by_default"`
	OnRegister          bool      `db:"on_register" json:"on_register,omitempty"`

	ThumbnailTimestamp string `db:"-" json:"-"`

	// User-specific permissions for this app, populated from permissions_apps
	// The db:"userperms" tag tells sqlx to look for columns prefixed with "userperms."
	UserPerms *PermApp `db:"userperms" json:"-"`
}

func Apps(userID, accountID int64, page, limit int) ([]App, error) {
	var its []App
	offset := (page - 1) * limit

	// Select app details (c.*) and user-specific permissions (pc.*)
	// Use aliases prefixed with "userperms." to match the struct tag and enable automatic scanning.
	query := `
	SELECT
		c.*,
		pc.id AS "userperms.id",
		pc.user_id AS "userperms.user_id",
		pc.account_id AS "userperms.account_id",
		pc.cookbook_id AS "userperms.cookbook_id",
		pc.app_id AS "userperms.app_id",
		pc.created AS "userperms.created",
		pc.updated AS "userperms.updated",
		pc.can_shared AS "userperms.can_shared",
                pc.can_individual AS "userperms.can_individual",
                pc.can_administer AS "userperms.can_administer",
                pc.is_owner AS "userperms.is_owner",
                pc.api_key AS "userperms.api_key"
        FROM
		Apps c
	LEFT JOIN
		permissions_apps pc ON c.id = pc.app_id AND pc.user_id = ?
	WHERE
		c.account_id = ?
		AND c.is_deleted = false
		AND pc.user_id IS NOT NULL -- Crucial: Ensure user has *some* specific permission record for the app
	ORDER BY
		c.updated DESC
	LIMIT ?
	OFFSET ?
	`

	// Use sqlx.Select which handles StructScan automatically for slices
	err := db.Db().Select(&its, query, userID, accountID, limit, offset)
	if err != nil {
		// sql.ErrNoRows is not an error in this context, just means no apps match.
		if errors.Is(err, sql.ErrNoRows) {
			return its, nil // Return empty slice
		}
		log.Printf("Error querying apps with permissions for user %d, account %d: %v", userID, accountID, err)
		return nil, err
	}

	// After Select, sqlx should have populated App and the embedded *PermApp (UserPerms) correctly.
	// If pc columns were NULL (due to LEFT JOIN and no match), UserPerms should remain nil (or be a pointer to a zero struct, need to verify sqlx behavior).
	// The WHERE pc.user_id IS NOT NULL should prevent rows where permissions are entirely missing.

	return its, nil
}

func Countapps(userID, accountID int64) (int, error) {
	var count int
	query := `
	SELECT
		COUNT(DISTINCT c.id) -- Use DISTINCT
	FROM
		Apps c
	LEFT JOIN -- Use LEFT JOIN for specific permissions
		permissions_apps pc ON c.id = pc.app_id AND pc.user_id = ?
	JOIN -- Use JOIN for account permissions (must exist)
		permissions_accounts pa ON c.account_id = pa.account_id AND pa.user_id = ?
	WHERE
		c.account_id = ?         -- Filter by requested Account ID
		AND c.is_deleted = false   -- Exclude deleted Apps
		AND (
			-- Check if specific permissions exist and grant access
			(pc.user_id IS NOT NULL AND (pc.can_shared = true OR pc.can_individual = true OR pc.can_administer = true OR pc.is_owner = true))
		)
	`
	err := db.Db().Get(&count, query, userID, userID, accountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil // No Apps found is not an error
		}
		log.Printf("Error counting Apps for user %d, account %d: %v", userID, accountID, err)
		return 0, err
	}
	return count, nil
}

func (c *App) Create(tx *sqlx.Tx) error {
	uuidWithTime, err := uuid.NewV7()
	if err != nil {
		return err
	}

	apiKeyWithTime, err := uuid.NewV7()
	if err != nil {
		return err
	}

	c.UUID = uuidWithTime.String()
	c.ApiKey = apiKeyWithTime.String()

	query := `
	INSERT INTO Apps 
		(account_id, owner_id, cookbook_id, uuid, name, description, yaml_shared, yaml_individual, api_key, is_active, is_deleted, is_assigned_by_default)
	VALUES
		(:account_id, :owner_id, :cookbook_id, :uuid, :name, :description, :yaml_shared, :yaml_individual, :api_key, :is_active, :is_deleted, :is_assigned_by_default)
	`
	sqlResponse, err := tx.NamedExec(query, c)
	if err != nil {
		return err
	}

	id, err := sqlResponse.LastInsertId()
	if err != nil {
		return err
	}

	query = `
        INSERT INTO permissions_apps
                (user_id, account_id, app_id, cookbook_id, can_shared, can_individual, can_administer, is_owner, api_key)
        VALUES
                (?, ?, ?, ?, true, true, true, true, ?)
        `

	userApiKey := uuid.New().String()

	log.Println(c.OwnerID, c.AccountID, id, c.CookbookID)

	_, err = tx.Exec(query, c.OwnerID, c.AccountID, id, c.CookbookID, userApiKey)
	if err != nil {
		log.Println(c)
		return err
	}
	return nil
}

// AppRegistrationInfo holds the minimal information needed for assigning
// permissions during user registration.
type AppRegistrationInfo struct {
	ID         int64 `db:"id"`
	CookbookID int64 `db:"cookbook_id"`
	AccountID  int64 `db:"account_id"`
}

// GetAppsForRegistrationByAccountID fetches apps within an account that are marked
// to be assigned to new users upon registration. It executes within the provided transaction.
func GetAppsForRegistrationByAccountID(tx *sqlx.Tx, accountID int64) ([]AppRegistrationInfo, error) {
	appsToAssign := []AppRegistrationInfo{}
	query := `SELECT id, cookbook_id, account_id 
	            FROM apps 
	           WHERE account_id = $1 
	             AND on_register = true 
	             AND is_deleted = false`

	err := tx.Select(&appsToAssign, query, accountID)
	if err != nil {
		// Don't treat ErrNoRows as a fatal error, just means no apps to assign.
		if errors.Is(err, sql.ErrNoRows) {
			return appsToAssign, nil // Return empty slice
		}
		// Wrap other errors for context
		return nil, fmt.Errorf("failed to query apps for registration for account %d: %w", accountID, err)
	}
	return appsToAssign, nil
}

func AppByUUIDAndAccountID(uuid string, accountID int64) (*App, error) {
	app := &App{}
	query := `
		SELECT 
			id, created, updated, account_id, owner_id, cookbook_id, uuid, name, description, yaml_shared, yaml_individual, api_key, is_active, is_deleted, is_assigned_by_default, on_register
		FROM 
			apps 
		WHERE 
			uuid = $1 AND account_id = $2
	`
	err := db.Db().Get(app, query, uuid, accountID)
	if err != nil {
		return nil, err
	}
	return app, nil
}

func (c *App) Update(tx *sqlx.Tx) (err error) {
	c.Updated = time.Now()

	prior := struct {
		YAMLShared     string `db:"yaml_shared"`
		YAMLIndividual string `db:"yaml_individual"`
	}{}

	if err = tx.Get(&prior, `SELECT yaml_shared, yaml_individual FROM apps WHERE id = ? AND account_id = ?`, c.ID, c.AccountID); err != nil {
		return err
	}

	query := `
        UPDATE Apps SET
                name = :name,
                description = :description,
                yaml_shared = :yaml_shared,
                yaml_individual = :yaml_individual,
                updated = :updated,
                -- You might want to allow updating other fields like is_active, etc.
                is_active = :is_active,
                is_deleted = :is_deleted,
                is_assigned_by_default = :is_assigned_by_default
        WHERE id = :id AND account_id = :account_id -- Ensure we only update the correct App in the correct account
        `
	result, err := tx.NamedExec(query, c)
	if err != nil {
		log.Printf("Error updating App ID %d: %v", c.ID, err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected for App update ID %d: %v", c.ID, err)
		return err // Return error, but the update might have succeeded
	}

	if rowsAffected == 0 {
		log.Printf("No rows affected when updating App ID %d for account ID %d. App might not exist or belong to this account.", c.ID, c.AccountID)
		return sql.ErrNoRows // Indicate that no record was updated
	}

	if prior.YAMLShared != c.YAMLShared || prior.YAMLIndividual != c.YAMLIndividual {
		if _, err = tx.Exec(`INSERT INTO app_history (app_id, yaml_shared, yaml_individual) VALUES (?, ?, ?)`, c.ID, prior.YAMLShared, prior.YAMLIndividual); err != nil {
			return err
		}
	}

	log.Printf("Successfully updated App ID %d. Rows affected: %d", c.ID, rowsAffected)
	return nil
}

// AppByUUID retrieves an App record by its UUID regardless of account.
func AppByUUID(uuid string) (*App, error) {
	app := &App{}
	query := `
                SELECT
                        id, created, updated, account_id, owner_id, cookbook_id, uuid, name, description, yaml_shared, yaml_individual, api_key, is_active, is_deleted, is_assigned_by_default, on_register
                FROM
                        apps
                WHERE
                        uuid = $1
        `
	err := db.Db().Get(app, query, uuid)
	if err != nil {
		return nil, err
	}
	return app, nil
}

// AppByAPIKey retrieves an App using its API key.
func AppByAPIKey(apiKey string) (*App, error) {
	app := &App{}
	query := `
                SELECT
                        id, created, updated, account_id, owner_id, cookbook_id, uuid, name, description, yaml_shared, yaml_individual, api_key, is_active, is_deleted, is_assigned_by_default, on_register
                FROM
                        apps
                WHERE
                        api_key = $1
        `
	err := db.Db().Get(app, query, apiKey)
	if err != nil {
		return nil, err
	}
	return app, nil
}

// AppByUUIDAndUserAPIKey fetches an App and user permission by app UUID and the user's API key.
func AppByUUIDAndUserAPIKey(uuidStr, apiKey string) (*App, *PermApp, error) {
	app, err := AppByUUID(uuidStr)
	if err != nil {
		return nil, nil, err
	}

	perm := &PermApp{}
	query := `SELECT id, user_id, account_id, cookbook_id, app_id, created, updated, can_shared, can_individual, can_administer, is_owner, api_key
                  FROM permissions_apps
                  WHERE app_id = $1 AND api_key = $2`
	err = db.Db().Get(perm, query, app.ID, apiKey)
	if err != nil {
		return nil, nil, err
	}

	return app, perm, nil
}

// AppByID retrieves an App record by its ID.
func AppByID(id int64) (*App, error) {
	app := &App{}
	query := `
                SELECT
                        id, created, updated, account_id, owner_id, cookbook_id, uuid, name, description,
                        yaml_shared, yaml_individual, api_key, is_active, is_deleted, is_assigned_by_default, on_register
                FROM
                        apps
                WHERE
                        id = $1
        `
	err := db.Db().Get(app, query, id)
	if err != nil {
		return nil, err
	}
	return app, nil
}
