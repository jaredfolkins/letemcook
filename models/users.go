package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/jaredfolkins/letemcook/db"
)

// UserDetailView contains data for the user detail page
type UserDetailView struct {
	BaseView
	User        User
	Cookbooks   []Cookbook        // Existing field for cookbooks
	Apps        []AppView         // Existing field for apps
	Permissions PermissionsBundle // Added field for all user permissions
}

// UsersView contains data for the users list page
// ... existing code ...

func (u UserDetailView) Title() string {
	return fmt.Sprintf("User Details: %s", u.User.Username)
}

// GetUserIDsForSharedCookbook retrieves a list of user IDs who have permission
// to view/interact with a shared cookbook, based on direct permissions,
// account admin rights, system admin rights, or ownership.
func GetUserIDsForSharedCookbook(cookbookUUID string) ([]int64, error) {
	db := db.Db()

	// 1. Find the cookbook details
	var cookbookID, accountID, ownerID int64
	err := db.QueryRow(`
        SELECT id, account_id, owner_id
        FROM cookbooks
        WHERE uuid = ? AND is_deleted = FALSE
    `, cookbookUUID).Scan(&cookbookID, &accountID, &ownerID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("GetUserIDsForSharedCookbook: Cookbook with UUID %s not found", cookbookUUID)
			return []int64{}, nil // Return empty list, not an error
		}
		log.Printf("GetUserIDsForSharedCookbook: Error querying cookbook %s: %v", cookbookUUID, err)
		return nil, fmt.Errorf("failed to query cookbook details: %w", err)
	}

	// 2. Query for all relevant user IDs using DISTINCT
	query := `
        SELECT DISTINCT u.id
        FROM users u
        WHERE u.is_deleted = FALSE AND u.is_disabled = FALSE AND (
            -- Direct cookbook permission (view OR edit)
            EXISTS (
                SELECT 1 FROM permissions_cookbooks pcb
                WHERE pcb.user_id = u.id AND pcb.cookbook_id = ? AND pcb.account_id = ? AND (pcb.can_view = TRUE OR pcb.can_edit = TRUE)
            )
            -- Is the owner
            OR u.id = ?
        )
    `

	rows, err := db.Query(query, cookbookID, accountID, ownerID)
	if err != nil {
		log.Printf("GetUserIDsForSharedCookbook: Error querying permissions for cookbook %d (UUID: %s): %v", cookbookID, cookbookUUID, err)
		return nil, fmt.Errorf("failed to query user permissions: %w", err)
	}
	defer rows.Close()

	var userIDs []int64
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			log.Printf("GetUserIDsForSharedCookbook: Error scanning user ID: %v", err)
			return nil, fmt.Errorf("failed to scan user ID: %w", err)
		}
		userIDs = append(userIDs, userID)
	}

	if err = rows.Err(); err != nil {
		log.Printf("GetUserIDsForSharedCookbook: Error iterating rows: %v", err)
		return nil, fmt.Errorf("failed to iterate user ID rows: %w", err)
	}

	log.Printf("GetUserIDsForSharedCookbook: Found %d users for cookbook UUID %s", len(userIDs), cookbookUUID)
	return userIDs, nil
}

// GetUserIDsForSharedApp retrieves a list of user IDs who have permission
// to view/interact with a shared app, based on direct permissions ('can_shared' or 'can_administer'),
// account admin rights, system admin rights, or ownership.
func GetUserIDsForSharedApp(appUUID string) ([]int64, error) {
	db := db.Db()

	// 1. Find the app details
	var appID, accountID, ownerID int64
	err := db.QueryRow(`
        SELECT id, account_id, owner_id
        FROM apps
        WHERE uuid = ? AND is_deleted = FALSE
    `, appUUID).Scan(&appID, &accountID, &ownerID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("GetUserIDsForSharedApp: App with UUID %s not found", appUUID)
			return []int64{}, nil // Return empty list, not an error
		}
		log.Printf("GetUserIDsForSharedApp: Error querying app %s: %v", appUUID, err)
		return nil, fmt.Errorf("failed to query app details: %w", err)
	}

	// 2. Query for all relevant user IDs using DISTINCT
	query := `
        SELECT DISTINCT u.id
        FROM users u
        WHERE u.is_deleted = FALSE AND u.is_disabled = FALSE AND (
            -- Direct app permission (shared or admin)
            EXISTS (
                SELECT 1 FROM permissions_apps pa
                WHERE pa.user_id = u.id AND pa.app_id = ? AND pa.account_id = ? AND (pa.can_shared = TRUE OR pa.can_administer = TRUE)
            )
            -- Is the owner
            OR u.id = ?
        )
    `

	rows, err := db.Query(query, appID, accountID, ownerID)
	if err != nil {
		log.Printf("GetUserIDsForSharedApp: Error querying permissions for app %d (UUID: %s): %v", appID, appUUID, err)
		return nil, fmt.Errorf("failed to query user permissions: %w", err)
	}
	defer rows.Close()

	var userIDs []int64
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			log.Printf("GetUserIDsForSharedApp: Error scanning user ID: %v", err)
			return nil, fmt.Errorf("failed to scan user ID: %w", err)
		}
		userIDs = append(userIDs, userID)
	}

	if err = rows.Err(); err != nil {
		log.Printf("GetUserIDsForSharedApp: Error iterating rows: %v", err)
		return nil, fmt.Errorf("failed to iterate user ID rows: %w", err)
	}

	log.Printf("GetUserIDsForSharedApp: Found %d users for app UUID %s", len(userIDs), appUUID)
	return userIDs, nil
}
