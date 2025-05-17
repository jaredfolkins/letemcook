package handlers

import (
	"log"
	"net/http"

	"github.com/jaredfolkins/letemcook/models"
	"github.com/jaredfolkins/letemcook/views/pages"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// GetProfileHandler renders the profile page for the currently logged in user
func GetProfileHandler(c LemcContext) error {
	userCtx := c.UserContext()
	if !userCtx.IsAuthenticated() {
		return echo.NewHTTPError(http.StatusUnauthorized, "Please log in to view your profile")
	}

	// Create the view model
	baseView := NewBaseView(c)
	baseView.Title = "My Profile"
	baseView.ActiveNav = "profile"

	view := models.UserDetailView{
		BaseView: baseView,
		User:     *userCtx.ActingAs,
	}

	// Check if this is a partial request
	partial := c.QueryParam("partial")

	// Generate the profile page content
	profileContent := pages.ProfilePage(view)

	if partial == "true" {
		// For HTMX requests, return just the profile content
		return HTML(c, profileContent)
	}

	// For full page loads, wrap the profile content in the base layout
	return HTML(c, pages.ProfileIndex(view, profileContent))
}

// FormChangePassword represents the password change form data
type FormChangePassword struct {
	CurrentPassword string
	NewPassword     string
	ConfirmPassword string
}

// PostChangePasswordHandler handles password change requests
func PostChangePasswordHandler(c LemcContext) error {
	// Check authentication
	userCtx := c.UserContext()
	if !userCtx.IsAuthenticated() {
		return echo.NewHTTPError(http.StatusUnauthorized, "Please log in to change your password")
	}

	// Get form values
	form := &FormChangePassword{
		CurrentPassword: c.FormValue("current_password"),
		NewPassword:     c.FormValue("new_password"),
		ConfirmPassword: c.FormValue("confirm_password"),
	}

	// Check if new passwords match
	if form.NewPassword != form.ConfirmPassword {
		c.AddErrorFlash("password", "New passwords do not match")
		return c.NoContent(http.StatusNoContent)
	}

	// Validate password length
	if len(form.NewPassword) < 12 {
		c.AddErrorFlash("password", "New password must be at least 12 characters long")
		return c.NoContent(http.StatusNoContent)
	}

	// Get current user
	userID := userCtx.ActingAs.ID
	accountID := userCtx.ActingAs.Account.ID

	user, err := models.UserByIDAndAccountID(userID, accountID)
	if err != nil {
		log.Printf("Error fetching user %d for account %d: %v", userID, accountID, err)
		c.AddErrorFlash("password", "Failed to verify user credentials")
		return c.NoContent(http.StatusNoContent)
	}

	// Verify current password
	err = bcrypt.CompareHashAndPassword([]byte(user.Hash), []byte(form.CurrentPassword))
	if err != nil {
		log.Printf("Password verification failed for user %d: %v", userID, err)
		c.AddErrorFlash("password", "Current password is incorrect")
		return c.NoContent(http.StatusNoContent)
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(form.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password for user %d: %v", userID, err)
		c.AddErrorFlash("password", "Failed to process new password")
		return c.NoContent(http.StatusNoContent)
	}

	// Update the password in the database
	err = models.UpdateUserPassword(userID, string(hashedPassword))
	if err != nil {
		log.Printf("Error updating password for user %d: %v", userID, err)
		c.AddErrorFlash("password", "Failed to update password")
		return c.NoContent(http.StatusNoContent)
	}

	// Add success flash
	c.AddSuccessFlash("password", "Password updated successfully")

	// Get the updated user data
	user, err = models.UserByIDAndAccountID(userID, accountID)
	if err != nil {
		log.Printf("Error fetching updated user data %d for account %d: %v", userID, accountID, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve updated user data")
	}

	permissions, err := models.GetAllPermissionsForUser(userID, accountID)
	if err != nil {
		log.Printf("Error fetching permissions for user %d: %v", userID, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve user permissions")
	}

	// Create view model
	baseView := NewBaseView(c)
	baseView.Title = "My Profile"

	view := models.UserDetailView{
		BaseView:    baseView,
		User:        *user,
		Permissions: permissions,
	}

	// Return just the profile content
	profileContent := pages.ProfilePage(view)
	return HTML(c, profileContent)
}

// PostToggleHeckleHandler handles toggling the heckle feature
func PostToggleHeckleHandler(c LemcContext) error {
	// Check authentication
	userCtx := c.UserContext()
	if !userCtx.IsAuthenticated() {
		return echo.NewHTTPError(http.StatusUnauthorized, "Please log in to update settings")
	}

	// Get the current user's ID
	userID := userCtx.ActingAs.ID

	// The presence of the key indicates the toggle is 'on' (true)
	heckleEnabled := c.FormValue("heckle_enabled")

	var heckleValue bool
	if heckleEnabled == "on" {
		heckleValue = true
	}

	log.Printf("heckleValue: %v", heckleValue)

	// Update the setting using the new model function
	err := models.UpdateUserHeckle(userID, heckleValue)
	if err != nil {
		// Generic error handling for the update
		log.Println()
		log.Println()
		log.Printf("Error updating heckle setting for user %d: %v", userID, err)
		c.AddErrorFlash("settings", "Failed to update heckle setting")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update setting")
	}

	// Add success flash
	statusMsg := "Heckle feature disabled"
	if heckleValue {
		statusMsg = "Heckle feature enabled"
	}
	c.AddSuccessFlash("settings", statusMsg)

	// Fetch updated settings to potentially return updated partials if needed later
	// Or simply return OK status if HTMX handles everything client-side based on toggle state
	// Based on the comment "HTMX handles the UI update", returning OK seems correct.
	return c.NoContent(http.StatusOK)
}
