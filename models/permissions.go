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

type Permission string

const (
	CanAccessCookbooksView Permission = "view_cookbook"
	CanAccessAppsView      Permission = "view_app"

	CanEditCookbook   Permission = "edit_cookbook"
	CanCreateCookbook Permission = "create_cookbook"

	CanEditApp   Permission = "edit_app" // Placeholder, adjust as needed
	CanCreateApp Permission = "create_app"

	CanAdministerAccount Permission = "admin_account"
	CanAdministerSystem  Permission = "admin_system"
	CanSharedApp         Permission = "shared_app"
	CanIndividualApp     Permission = "individual_app"
	CanAclApp            Permission = "acl_app"

	// Permissions used for toggling flags
	ToggleAppIndividual Permission = "toggle_app_individual"
	ToggleAppShared     Permission = "toggle_app_shared"
	ToggleAppAdminister Permission = "toggle_app_administer"
)

func HasCookbookPermission(userID, accountID int64, cookbookUUID string, perm Permission) (bool, error) {
	var cookbookID int64
	err := db.Db().QueryRow("SELECT id FROM cookbooks WHERE uuid = ?", cookbookUUID).Scan(&cookbookID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("Cookbook not found for UUID: %s", cookbookUUID)
			return false, nil
		}
		log.Printf("Error getting cookbook ID for UUID %s: %v", cookbookUUID, err)
		return false, err // Internal error
	}

	var canView, canEdit, isOwner bool
	query := `SELECT can_view, can_edit, is_owner
	           FROM permissions_cookbooks
	           WHERE user_id = ? AND account_id = ? AND cookbook_id = ?`
	err = db.Db().QueryRow(query, userID, accountID, cookbookID).Scan(&canView, &canEdit, &isOwner)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("No specific cookbook permission found for user %d, account %d, cookbook %d (UUID %s)", userID, accountID, cookbookID, cookbookUUID)
			return false, nil
		}
		log.Printf("Error querying cookbook permissions for user %d, cookbook %d: %v", userID, cookbookID, err)
		return false, err // Internal error
	}

	switch perm {
	case CanAccessCookbooksView:
		return canView || canEdit || isOwner, nil
	case CanEditCookbook:
		return canEdit || isOwner, nil
	default:
		log.Printf("Unhandled permission type in HasCookbookPermission: %s", perm)
		return false, errors.New("internal configuration error: unhandled permission type")
	}
}

func HasAppPermission(userID, accountID int64, appUUID string, perm Permission) (bool, error) {
	var appID int64
	err := db.Db().QueryRow("SELECT id FROM Apps WHERE uuid = ?", appUUID).Scan(&appID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("App not found for UUID: %s", appUUID)
			return false, nil // App doesn't exist, no permission.
		}
		log.Printf("Error getting App ID for UUID %s: %v", appUUID, err)
		return false, err // Internal error
	}

	var canShared, canIndividual, canAdmin, isOwner bool
	query := `SELECT can_shared, can_individual, can_administer, is_owner
	           FROM permissions_apps
	           WHERE user_id = ? AND account_id = ? AND app_id = ?`
	err = db.Db().QueryRow(query, userID, accountID, appID).Scan(&canShared, &canIndividual, &canAdmin, &isOwner)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("No specific App permission found for user %d, account %d, App %d (UUID %s)", userID, accountID, appID, appUUID)
			return false, nil
		}
		log.Printf("Error querying App permissions for user %d, App %d: %v", userID, appID, err)
		return false, err // Internal error
	}

	log.Printf("Permission check for user %d, App %d (UUID %s), perm '%s'. DB values: shared=%t, indiv=%t, admin=%t, owner=%t",
		userID, appID, appUUID, perm, canShared, canIndividual, canAdmin, isOwner)

	switch perm {
	case CanAccessAppsView:
		return true, nil
	case CanSharedApp:
		return canShared, nil
	case CanIndividualApp:
		log.Printf("Returning result for CanIndividualApp: %t", canIndividual) // ADD LOGGING
		return canIndividual, nil
	case CanAclApp: // Renamed from CanEditApp? Let's assume this means managing permissions/admin
		return canAdmin || isOwner, nil
	case CanEditApp: // Define what "Edit App" means - separate from ACL? e.g., editing content/details
		return canAdmin || isOwner, nil // Keeping original logic for now
	default:
		log.Printf("Unhandled permission type in HasAppPermission: %s", perm)
		return false, errors.New("internal configuration error: unhandled permission type")
	}
}

func HasAccountPermission(userID, accountID int64, perm Permission) (bool, error) {
	var canAdminister, canCreateapps, canViewapps, canCreateCookbooks, canViewCookbooks, isOwner bool
	query := `SELECT can_administer, can_create_apps, can_view_apps,
	                 can_create_cookbooks, can_view_cookbooks, is_owner
	           FROM permissions_accounts
	           WHERE user_id = ? AND account_id = ?`
	err := db.Db().QueryRow(query, userID, accountID).Scan(
		&canAdminister, &canCreateapps, &canViewapps,
		&canCreateCookbooks, &canViewCookbooks, &isOwner,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("No specific account permission found for user %d, account %d", userID, accountID)
			return false, nil
		}
		log.Printf("Error querying account permissions for user %d, account %d: %v", userID, accountID, err)
		return false, err // Internal error
	}

	switch perm {
	case CanAccessCookbooksView:
		return canViewCookbooks || canAdminister || isOwner, nil
	case CanAccessAppsView:
		return canViewapps || canAdminister || isOwner, nil
	case CanCreateCookbook:
		return canCreateCookbooks || canAdminister || isOwner, nil
	case CanCreateApp:
		return canCreateapps || canAdminister || isOwner, nil
	case CanAdministerAccount:
		return canAdminister || isOwner, nil
	default:
		log.Printf("Unhandled permission type in HasAccountPermission: %s", perm)
		return false, errors.New("internal configuration error: unhandled permission type")
	}
}

func HasSystemPermission(userID int64, perm Permission) (bool, error) {
	var canAdminister, isOwner bool
	query := `SELECT can_administer, is_owner
	           FROM permissions_system
	           WHERE user_id = ?`
	err := db.Db().QueryRow(query, userID).Scan(&canAdminister, &isOwner)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("No specific App permission found for user %d", userID)
			return false, nil
		}
		log.Printf("Error querying App permissions for user %d: %v", userID, err)
		return false, err // Internal error
	}

	switch perm {
	case CanAdministerSystem:
		return canAdminister || isOwner, nil
	default:
		log.Printf("Unhandled permission type in HasSystemPermission: %s", perm)
		return false, errors.New("internal configuration error: unhandled permission type")
	}
}

type Permissions struct {
	PermSystem           *PermSystem
	PermissionsAccounts  []*PermAccount
	PermissionsCookbooks []*PermCookbook
	PermissionsApps      []*PermApp
}

type PermSystem struct {
	Created       time.Time `db:"created"`
	Updated       time.Time `db:"updated"`
	ID            int64     `db:"id"`
	UserID        int64     `db:"user_id"`
	AccountID     int64     `db:"account_id"`
	CanAdminister bool      `db:"can_administer"`
	IsOwner       bool      `db:"is_owner"`
}

type PermAccount struct {
	Created            time.Time `db:"created"`
	Updated            time.Time `db:"updated"`
	ID                 int64     `db:"id"`
	UserID             int64     `db:"user_id"`
	AccountID          int64     `db:"account_id"`
	AccountName        string    `db:"account_name"`
	CanAdminister      bool      `db:"can_administer"`
	CanCreateapps      bool      `db:"can_create_apps"`
	CanViewapps        bool      `db:"can_view_apps"`
	CanCreateCookbooks bool      `db:"can_create_cookbooks"`
	CanViewCookbooks   bool      `db:"can_view_cookbooks"`
	IsOwner            bool      `db:"is_owner"`
}
type PermApp struct {
	Created       time.Time `db:"created"`
	Updated       time.Time `db:"updated"`
	ID            int64     `db:"id"`
	UserID        int64     `db:"user_id"`
	AccountID     int64     `db:"account_id"`
	CookbookID    int64     `db:"cookbook_id"`
	AppID         int64     `db:"app_id"`
	CanShared     bool      `db:"can_shared"`
	CanIndividual bool      `db:"can_individual"`
	CanAdminister bool      `db:"can_administer"`
	IsOwner       bool      `db:"is_owner"`
	ApiKey        string    `db:"api_key"`
}

func (pc *PermApp) UpsertappPermissions(tx *sqlx.Tx) error {
	if pc.ApiKey == "" {
		pc.ApiKey = uuid.New().String()
	}
	query := `
        INSERT INTO permissions_apps
                (user_id, account_id, app_id, cookbook_id, can_shared, can_individual, can_administer, is_owner, api_key)
        VALUES
                ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        ON CONFLICT (user_id, account_id, app_id)
        DO UPDATE SET
                can_shared = excluded.can_shared,
                can_individual = excluded.can_individual,
                can_administer = excluded.can_administer,
                is_owner = excluded.is_owner,
                -- Note: cookbook_id might change if the App is re-linked, but typically shouldn't be updated here.
                -- If it needs updating, add: cookbook_id = excluded.cookbook_id
                updated = CURRENT_TIMESTAMP -- Update the timestamp on modification
        `
	_, err := tx.Exec(query,
		pc.UserID, pc.AccountID, pc.AppID, pc.CookbookID,
		pc.CanShared, pc.CanIndividual, pc.CanAdminister, pc.IsOwner,
		pc.ApiKey,
	)
	if err != nil {
		log.Printf("🔥 Failed to upsert App permission for user %d, App %d: %v", pc.UserID, pc.AppID, err)
		return err
	}
	return nil
}

func (pc *PermApp) DeleteappPermissions(tx *sqlx.Tx) error {
	query := `
		DELETE 
		FROM 
			permissions_apps 
		WHERE
			user_id = :user_id 
		AND 
			account_id = :account_id 
		AND 
			app_id = :app_id`

	_, err := tx.NamedExec(query, pc)
	if err != nil {
		log.Printf("🔥 Failed to delete App permission for user %d, App %d: %v", pc.UserID, pc.AppID, err)
		return err
	}
	return nil
}

type PermCookbook struct {
	Created    time.Time `db:"created"`
	Updated    time.Time `db:"updated"`
	ID         int64     `db:"id"`
	UserID     int64     `db:"user_id"`
	AccountID  int64     `db:"account_id"`
	CookbookID int64     `db:"cookbook_id"`
	CanView    bool      `db:"can_view"`
	CanEdit    bool      `db:"can_edit"`
	IsOwner    bool      `db:"is_owner"`
}

func (u *User) CanCreateCookbook() bool {
	b, err := HasAccountPermission(u.ID, u.Account.ID, CanCreateCookbook)
	if err != nil {
		log.Printf("Error checking CanCreateCookbook for user %d: %v", u.ID, err)
	}
	return b
}

func (u *User) CanCreateapp() bool {
	b, err := HasAccountPermission(u.ID, u.Account.ID, CanCreateApp)
	if err != nil {
		log.Printf("Error checking CanCreateApp for user %d: %v", u.ID, err)
	}
	return b
}

// CanAdministerAccount checks if the user has permission to administer the current account.
func (u *User) CanAdministerAccount() bool {
	b, err := HasAccountPermission(u.ID, u.Account.ID, CanAdministerAccount)
	if err != nil {
		log.Printf("Error checking CanAdministerAccount for user %d: %v", u.ID, err)
	}
	return b
}

// PermissionsBundle holds all permissions for a user across different scopes.
type PermissionsBundle struct {
	System    *PermSystem                `json:"system,omitempty"`
	Accounts  []*PermAccount             `json:"accounts,omitempty"`
	Cookbooks []*PermCookbookWithDetails `json:"cookbooks,omitempty"`
	Apps      []*PermAppWithDetails      `json:"apps,omitempty"`
}

// PermCookbookWithDetails adds cookbook details to the permission info.
type PermCookbookWithDetails struct {
	PermCookbook
	CookbookName string `db:"cookbook_name"`
	CookbookUUID string `db:"cookbook_uuid"`
}

// PermAppWithDetails adds app and cookbook details to the permission info.
type PermAppWithDetails struct {
	PermApp
	AppName      string `db:"app_name"`
	AppUUID      string `db:"app_uuid"`
	CookbookName string `db:"cookbook_name"` // Name of the cookbook the app belongs to
	CookbookUUID string `db:"cookbook_uuid"` // UUID of the cookbook the app belongs to
}

// GetAllPermissionsForUser retrieves all known permissions for a given user and account.
func GetAllPermissionsForUser(userID int64, accountID int64) (PermissionsBundle, error) {
	bundle := PermissionsBundle{}
	var err error

	// Get System Permissions
	bundle.System = &PermSystem{} // Initialize to avoid nil pointer if no row found initially
	querySystem := `SELECT * FROM permissions_system WHERE user_id = $1`
	err = db.Db().Get(bundle.System, querySystem, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("No system permissions found for user %d", userID)
			bundle.System = nil // Set back to nil if truly no rows
		} else {
			log.Printf("Error getting system permissions for user %d: %v", userID, err)
			return bundle, fmt.Errorf("failed to get system permissions: %w", err)
		}
	}

	// Get Account Permissions (for ALL accounts the user has permissions in)
	queryAccounts := `
		SELECT pa.*, a.name as account_name
		FROM permissions_accounts pa
		JOIN accounts a ON pa.account_id = a.id
		WHERE pa.user_id = $1
		ORDER BY a.name
	`
	err = db.Db().Select(&bundle.Accounts, queryAccounts, userID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("Error getting account permissions for user %d: %v", userID, err) // Updated log message
		return bundle, fmt.Errorf("failed to get account permissions: %w", err)
	}

	// Get Cookbook Permissions with details
	queryCookbooks := `
		SELECT 
			p.*, 
			c.name as cookbook_name,
			c.uuid as cookbook_uuid
		FROM permissions_cookbooks p
		JOIN cookbooks c ON p.cookbook_id = c.id
		WHERE p.user_id = $1 AND p.account_id = $2
		ORDER BY c.name
	`
	err = db.Db().Select(&bundle.Cookbooks, queryCookbooks, userID, accountID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("Error getting cookbook permissions for user %d, account %d: %v", userID, accountID, err)
		return bundle, fmt.Errorf("failed to get cookbook permissions: %w", err)
	}

	// Get App Permissions with details
	queryApps := `
		SELECT 
			p.*, 
			a.name as app_name,
			a.uuid as app_uuid,
			cb.name as cookbook_name,
			cb.uuid as cookbook_uuid 
		FROM permissions_apps p
		JOIN apps a ON p.app_id = a.id
		JOIN cookbooks cb ON a.cookbook_id = cb.id -- Assuming apps have a cookbook_id
		WHERE p.user_id = $1 AND p.account_id = $2
		ORDER BY a.name
	`
	err = db.Db().Select(&bundle.Apps, queryApps, userID, accountID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("Error getting app permissions for user %d, account %d: %v", userID, accountID, err)
		return bundle, fmt.Errorf("failed to get app permissions: %w", err)
	}

	log.Printf("Successfully fetched permissions bundle for user %d, account %d. System: %v, Accounts: %d, Cookbooks: %d, Apps: %d",
		userID, accountID, bundle.System != nil, len(bundle.Accounts), len(bundle.Cookbooks), len(bundle.Apps))

	return bundle, nil
}

// AssignAppPermissionOnRegister inserts a permission record for a user and an app,
// specifically for the initial registration process where default permissions are granted.
// It uses ON CONFLICT DO NOTHING to avoid errors if a permission already exists (e.g., re-registration attempt).
func AssignAppPermissionOnRegister(tx *sqlx.Tx, userID, accountID, appID, cookbookID int64) error {
	permQuery := `
                INSERT INTO permissions_apps
                        (user_id, account_id, app_id, cookbook_id, can_individual, can_shared, can_administer, is_owner, api_key)
                VALUES
                        ($1, $2, $3, $4, $5, $6, $7, $8, $9)
                ON CONFLICT (user_id, account_id, app_id)
                DO NOTHING
        `
	apiKey := uuid.New().String()
	// Assign default permissions: can_individual=true, others=false
	_, err := tx.Exec(permQuery, userID, accountID, appID, cookbookID, true, false, false, false, apiKey)
	if err != nil {
		// Wrap error for context
		return fmt.Errorf("failed to assign initial app permission for user %d, app %d, account %d: %w", userID, appID, accountID, err)
	}
	log.Printf("Assigned app %d to new user %d in account %d via on_register", appID, userID, accountID)
	return nil
}

// ToggleAppPermission fetches an app permission, toggles a specific boolean field,
// and updates it in the database within a transaction.
func ToggleAppPermission(userID, accountID, appID int64, permToToggle Permission) error {
	tx, err := db.Db().Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction for toggling permission: %w", err)
	}
	// Defer rollback in case of errors after this point
	defer func() {
		if err != nil {
			rbErr := tx.Rollback()
			if rbErr != nil {
				log.Printf("Error rolling back transaction after failing to toggle permission: %v", rbErr)
			}
		}
	}()

	// 1. Fetch the current permission record within the transaction
	pc := PermApp{}
	// Select only necessary fields + id for update, and use FOR UPDATE for locking
	getQuery := `SELECT id, can_shared, can_individual, can_administer
	               FROM permissions_apps
	              WHERE user_id = $1 AND account_id = $2 AND app_id = $3 FOR UPDATE`
	err = tx.Get(&pc, getQuery, userID, accountID, appID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("permission record not found for user %d, account %d, app %d", userID, accountID, appID)
		}
		return fmt.Errorf("failed to fetch permission record for toggle (user %d, app %d): %w", userID, appID, err)
	}

	// 2. Toggle the appropriate field based on permToToggle
	var fieldToUpdate string
	switch permToToggle {
	case ToggleAppIndividual:
		pc.CanIndividual = !pc.CanIndividual
		fieldToUpdate = "can_individual"
	case ToggleAppShared:
		pc.CanShared = !pc.CanShared
		fieldToUpdate = "can_shared"
	case ToggleAppAdminister:
		pc.CanAdminister = !pc.CanAdminister
		fieldToUpdate = "can_administer"
	default:
		// Rollback explicitly here as the defer won't trigger if we return early
		tx.Rollback() // Ignoring rollback error here
		return fmt.Errorf("invalid permission type provided for toggling: %s", permToToggle)
	}

	// 3. Update the record using NamedExec
	// Note: We dynamically build the SET clause to only update the toggled field and 'updated'.
	updateQuery := fmt.Sprintf(`UPDATE permissions_apps SET
	                                %s = :%s,
	                                updated = CURRENT_TIMESTAMP
	                            WHERE id = :id`, fieldToUpdate, fieldToUpdate)

	_, err = tx.NamedExec(updateQuery, pc)
	if err != nil {
		// Error is already set, defer will handle rollback
		return fmt.Errorf("failed to update permission record after toggle (user %d, app %d, field %s): %w", userID, appID, fieldToUpdate, err)
	}

	// 4. Commit the transaction
	err = tx.Commit()
	if err != nil {
		// Error is set, defer *should not* rollback after a failed commit, but we still return the error
		return fmt.Errorf("failed to commit transaction after toggling permission (user %d, app %d): %w", userID, appID, err)
	}

	log.Printf("Successfully toggled permission '%s' for user %d, app %d", fieldToUpdate, userID, appID)
	return nil
}

// AppPermissionsByUserAccountAndApp fetches specific app permissions for a given user, account, and app ID.
// It returns the permissions object or an error. If no specific permissions are found,
// it returns a zero-value PermApp struct and no error.
func AppPermissionsByUserAccountAndApp(userID, accountID, appID int64) (*PermApp, error) {
	appPerms := &PermApp{}
	permQuery := `SELECT id, user_id, account_id, cookbook_id, app_id, created, updated, can_shared, can_individual, can_administer, is_owner, api_key
                       FROM permissions_apps
                       WHERE user_id = $1 AND account_id = $2 AND app_id = $3`
	err := db.Db().Get(appPerms, permQuery, userID, accountID, appID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// No specific permissions found. Return an empty struct to indicate this.
			log.Printf("No specific app permissions found for user %d, account %d, app %d. Returning default (zero) permissions.", userID, accountID, appID)
			return &PermApp{}, nil // Return empty struct, not an error
		}
		// Other database error fetching permissions.
		log.Printf("Error fetching app permissions for user %d, account %d, app %d: %v", userID, accountID, appID, err)
		return nil, fmt.Errorf("database error fetching app permissions: %w", err)
	}
	// Permissions found
	return appPerms, nil
}
