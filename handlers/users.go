package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/jaredfolkins/letemcook/models"
	"github.com/jaredfolkins/letemcook/paths"
	"github.com/jaredfolkins/letemcook/util"
	"github.com/jaredfolkins/letemcook/views/pages"
)

// Function to set flash message header (Helper)
func setFlashHeader(c LemcContext, flashType string, message string) error {
	// Store message as a map[string]string for consistency with JS parsing
	flashMap := map[string]string{"message": message}
	jsonMsg, err := json.Marshal(flashMap)
	if err != nil {
		log.Printf("Error marshaling flash message: %v", err)
		// Don't fail the request, just log the error
		return nil // Or return the error if you want to handle it upstream
	}

	headerName := ""
	if flashType == "success" {
		headerName = "X-Lemc-Flash-Success"
	} else if flashType == "error" {
		headerName = "X-Lemc-Flash-Error"
	} else {
		log.Printf("Unknown flash type: %s", flashType)
		return nil // Or return an error
	}

	c.Response().Header().Set(headerName, string(jsonMsg))
	return nil
}

const DefaultUserLimit = 10 // Define a default limit for users

func GetAllUsers(c LemcContext) error {
	partial := strings.ToLower(c.QueryParam("partial")) // Added for potential partial updates later
	pageStr := c.QueryParam("page")
	limitStr := c.QueryParam("limit")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = DefaultUserLimit
	}

	accountID := c.UserContext().ActingAs.Account.ID

	totalUsers, err := models.CountUsersByAccountID(accountID)
	if err != nil {
		log.Printf("Error counting users for account %d: %v", accountID, err)
		return c.String(http.StatusInternalServerError, "Failed to retrieve user count")
	}
	totalPages := 0
	if totalUsers > 0 {
		totalPages = int(math.Ceil(float64(totalUsers) / float64(limit)))
	}

	users, err := models.GetUsersByAccountID(accountID, page, limit)
	if err != nil {
		log.Printf("Error fetching users for account %d (page %d, limit %d): %v", accountID, page, limit, err)
		return c.String(http.StatusInternalServerError, "Failed to retrieve users")
	}

	baseView := NewBaseView(c, WithNavigation("account", paths.AccountUsers))
	v := models.UsersView{
		BaseView:    baseView,
		Users:       users,
		CurrentPage: page,
		TotalPages:  totalPages,
		Limit:       limit,
	}

	up := pages.UsersPage(v)
	if partial == "true" {
		return HTML(c, up) // Render only the user list partial
	}

	uv := pages.UsersIndex(v, up) // Adjust if your full page template differs
	return HTML(c, uv)
}

func GetUserHandler(c LemcContext) error {
	idStr := c.Param("id")
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.Printf("Invalid user ID: %s, Error: %v", idStr, err)
		return c.String(http.StatusBadRequest, "Invalid user ID")
	}

	loggedInUserAccountID := c.UserContext().ActingAs.Account.ID

	user, err := models.UserByIDAndAccountID(userID, loggedInUserAccountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("User %d not found in account %d", userID, loggedInUserAccountID)
			return c.String(http.StatusNotFound, "User not found")

		}
		log.Printf("Error fetching user %d for account %d: %v", userID, loggedInUserAccountID, err)
		return c.String(http.StatusInternalServerError, "Failed to retrieve user information")
	}

	cookbooks, err := models.Cookbooks(userID, 1, 20) // Using userID from URL param
	if err != nil {
		log.Printf("Error fetching cookbooks for user %d: %v", userID, err)
		cookbooks = []models.Cookbook{} // Default to empty list on error
	}

	rawapps, err := models.Apps(userID, loggedInUserAccountID, 1, 20)
	var apps []models.AppView
	if err != nil {
		log.Printf("Error fetching apps for user %d in account %d: %v", userID, loggedInUserAccountID, err)
		apps = []models.AppView{} // Default to empty list on error
	} else {
		apps = make([]models.AppView, len(rawapps))
		for i, rc := range rawapps {
			apps[i] = models.AppView{
				ID:    rc.ID,
				UUID:  rc.UUID,
				Title: rc.Name,
			}
		}
	}

	// Fetch all permissions for the user
	permissions, err := models.GetAllPermissionsForUser(user.ID, user.Account.ID)
	if err != nil {
		log.Printf("Error fetching permissions for user %d: %v", user.ID, err)
		// Decide if this is a fatal error or if the page can still be shown
		// For now, let's log and continue, the template will need to handle missing permissions
		// Alternatively, return an internal server error:
		// return c.String(http.StatusInternalServerError, "Failed to retrieve user permissions")
	}

	baseView := NewBaseView(c,
		WithTitle("User: "+user.Username),
		WithNavigation("account", paths.AccountUsers))
	v := models.UserDetailView{
		BaseView:    baseView,
		User:        *user,
		Cookbooks:   cookbooks,
		Apps:        apps,
		Permissions: permissions,
	}

	userDetailComponent := pages.UserDetailPage(v) // This component needs to be created
	return HTML(c, userDetailComponent)
}

// Helper function to parse pagination parameters
func parsePaginationParams(c LemcContext) (page int, limit int) {
	pageStr := c.QueryParam("page")
	limitStr := c.QueryParam("limit")

	page, _ = strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	limit, _ = strconv.Atoi(limitStr)
	if limit < 1 {
		limit = DefaultUserLimit
	}

	return page, limit
}

// Helper function to build users view for pagination
func buildUsersView(c LemcContext, users []models.User, totalUsers int, page, limit int) models.UsersView {
	totalPages := 0
	if totalUsers > 0 {
		totalPages = int(math.Ceil(float64(totalUsers) / float64(limit)))
	}

	baseView := NewBaseView(c, WithNavigation("account", paths.AccountUsers))
	return models.UsersView{
		BaseView:    baseView,
		Users:       users,
		CurrentPage: page,
		TotalPages:  totalPages,
		Limit:       limit,
	}
}

// Helper function to handle validation error response
func handleValidationError(c LemcContext, message string, actingUser *models.User) error {
	setFlashHeader(c, "error", message)

	page, limit := parsePaginationParams(c)
	users, _ := models.GetUsersByAccountID(actingUser.Account.ID, page, limit)
	totalUsers, _ := models.CountUsersByAccountID(actingUser.Account.ID)

	v := buildUsersView(c, users, totalUsers, page, limit)
	return HTML(c, pages.UsersPartial(v))
}

// Helper function to handle creation error response
func handleCreationError(c LemcContext, err error, actingUser *models.User) error {
	log.Printf("Error creating user by %d: %v", actingUser.ID, err)

	errMsg := "Failed to create user. Please try again."
	if strings.Contains(err.Error(), "UNIQUE constraint failed: users.email") {
		errMsg = "Email address is already in use."
	} else if strings.Contains(err.Error(), "UNIQUE constraint failed: users.username") {
		errMsg = "Username is already taken."
	}

	return handleValidationError(c, errMsg, actingUser)
}

// Helper function to build success response
func buildSuccessResponse(c LemcContext, username string, actingUser *models.User) error {
	setFlashHeader(c, "success", fmt.Sprintf("User '%s' created successfully.", username))

	page, limit := parsePaginationParams(c)

	totalUsers, err := models.CountUsersByAccountID(actingUser.Account.ID)
	if err != nil {
		log.Printf("Error counting users after creation: %v", err)
		return c.String(http.StatusInternalServerError, "Error retrieving updated user list")
	}

	users, err := models.GetUsersByAccountID(actingUser.Account.ID, page, limit)
	if err != nil {
		log.Printf("Error fetching users after creation: %v", err)
		return c.String(http.StatusInternalServerError, "Error retrieving updated user list")
	}

	v := buildUsersView(c, users, totalUsers, page, limit)
	return HTML(c, pages.UsersPartial(v))
}

func CreateUserHandler(c LemcContext) error {
	actingUser := c.UserContext().ActingAs

	// Check permission first
	if !actingUser.CanAdministerAccount() {
		log.Printf("User %d attempted to create user without permission in account %d", actingUser.ID, actingUser.Account.ID)
		setFlashHeader(c, "error", "You do not have permission to create users.")
		return c.String(http.StatusForbidden, "Forbidden")
	}

	// Extract and validate form values
	username := util.Sanitize(c.FormValue("username"))
	email := util.Sanitize(c.FormValue("email"))
	password := c.FormValue("password")

	// Validate required fields
	if username == "" || email == "" || password == "" {
		log.Printf("Attempt to create user with missing fields by user %d", actingUser.ID)
		return handleValidationError(c, "Username, email, and password cannot be empty.", actingUser)
	}

	// Attempt to create user
	_, err := models.CreateUser(username, email, password, actingUser.Account.ID)
	if err != nil {
		return handleCreationError(c, err, actingUser)
	}

	// Handle success
	return buildSuccessResponse(c, username, actingUser)
}
