package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/jaredfolkins/letemcook/db"
	"github.com/jaredfolkins/letemcook/models"
)

func PutUserAccountPermissionToggleHandler(c LemcContext) error {
	userIDStr := c.Param("user_id")
	accountIDStr := c.Param("account_id")
	permissionName := c.Param("permission_name")

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		slog.Error("Invalid user ID", "error", err, "userID", userIDStr)
		return c.NoContent(http.StatusBadRequest)
	}

	accountID, err := strconv.ParseInt(accountIDStr, 10, 64)
	if err != nil {
		slog.Error("Invalid account ID", "error", err, "accountID", accountIDStr)
		return c.NoContent(http.StatusBadRequest)
	}

	// Determine the new boolean state from the form value
	// HTMX sends the value of the checkbox when it's checked, and nothing when unchecked.
	// We need to check if the key exists in the form data.
	formParams, err := c.FormParams()
	if err != nil {
		slog.Error("Failed to parse form params", "error", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	// The presence of the key indicates the toggle is 'on' (true)
	newValue := formParams.Has(permissionName)

	slog.Info("Toggling account permission", "userID", userID, "accountID", accountID, "permission", permissionName, "newValue", newValue)

	// Correctly access the database connection via the db package
	err = models.UpdateUserAccountPermission(db.Db().DB, userID, accountID, permissionName, newValue)
	if err != nil {
		slog.Error("Failed to update user account permission", "error", err, "userID", userID, "accountID", accountID, "permission", permissionName)
		// Add an error flash message
		c.AddErrorFlash("perm-update", fmt.Sprintf("Failed to update permission '%s' for account %d", permissionName, accountID))
		return c.NoContent(http.StatusInternalServerError)
	}

	// Add a success flash message including the account ID
	flashMsg := fmt.Sprintf("Permission '%s' for account %d updated successfully.", permissionName, accountID)
	c.AddSuccessFlash("perm-update", flashMsg)

	// Return No Content, HTMX handles the UI update based on the toggle state
	return c.NoContent(http.StatusOK)
}
