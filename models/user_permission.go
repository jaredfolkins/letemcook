package models

import (
	"database/sql"
	"fmt"
	"log/slog"
)

// UpdateUserAccountPermission updates a specific boolean permission for a user on an account.
func UpdateUserAccountPermission(db *sql.DB, userID, accountID int64, permissionName string, newValue bool) error {
	// Map the permission name from the URL/form to the actual database column name.
	// This prevents SQL injection by ensuring only valid column names are used.
	var columnName string
	switch permissionName {
	case "can_administer":
		columnName = "can_administer"
	case "can_create_apps":
		columnName = "can_create_apps" // Make sure this matches your schema
	case "can_view_apps":
		columnName = "can_view_apps" // Make sure this matches your schema
	case "can_create_cookbooks":
		columnName = "can_create_cookbooks"
	case "can_view_cookbooks":
		columnName = "can_view_cookbooks"
	case "is_owner":
		columnName = "is_owner"
	default:
		// Log and return an error if the permission name is not recognized.
		errMsg := fmt.Sprintf("invalid permission name: %s", permissionName)
		slog.Error(errMsg)
		return fmt.Errorf("%s", errMsg)
	}

	// Convert boolean to integer for SQLite (assuming 1 for true, 0 for false)
	var newValueInt int
	if newValue {
		newValueInt = 1
	} else {
		newValueInt = 0
	}

	// Construct the dynamic SQL query safely using the correct table name.
	query := fmt.Sprintf("UPDATE permissions_accounts SET %s = ? WHERE user_id = ? AND account_id = ?", columnName)

	// Execute the query.
	result, err := db.Exec(query, newValueInt, userID, accountID)
	if err != nil {
		slog.Error("Database error updating user account permission", "error", err, "query", query, "userID", userID, "accountID", accountID, "column", columnName, "newValue", newValueInt)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error("Error getting rows affected after update", "error", err)
		// Continue even if we can't get rows affected, the update might have succeeded.
	}

	if rowsAffected == 0 {
		slog.Warn("No rows updated. User/Account permission record might not exist.", "userID", userID, "accountID", accountID, "column", columnName)
		// Depending on requirements, you might want to return an error here or ensure the record exists.
	}

	slog.Info("Successfully updated user account permission", "userID", userID, "accountID", accountID, "permission", columnName, "newValue", newValue)
	return nil
}
