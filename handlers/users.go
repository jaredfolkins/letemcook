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

	baseView := NewBaseView(c)
	baseView.ActiveNav = "account"
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
			baseView := NewBaseView(c)
			baseView.Title = "Not Found"
			return c.String(http.StatusNotFound, "User not found")

		}
		log.Printf("Error fetching user %d for account %d: %v", userID, loggedInUserAccountID, err)
		baseView := NewBaseView(c)
		baseView.Title = "Server Error"
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

	baseView := NewBaseView(c)
	baseView.Title = "User: " + user.Username // Set a specific title
	baseView.ActiveNav = "account"
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

func CreateUserHandler(c LemcContext) error {
	actingUser := c.UserContext().ActingAs
	if !actingUser.CanAdministerAccount() {
		log.Printf("User %d attempted to create user without permission in account %d", actingUser.ID, actingUser.Account.ID)
		setFlashHeader(c, "error", "You do not have permission to create users.")
		return c.String(http.StatusForbidden, "Forbidden")
	}

	username := util.Sanitize(c.FormValue("username"))
	email := util.Sanitize(c.FormValue("email"))
	password := c.FormValue("password")

	if username == "" || email == "" || password == "" {
		log.Printf("Attempt to create user with missing fields by user %d", actingUser.ID)
		setFlashHeader(c, "error", "Username, email, and password cannot be empty.")
		// Re-render the current user list (or an empty list) to show the error
		// We need the current pagination info to fetch the correct list
		pageStr := c.QueryParam("page") // Or default to 1 if not submitting from a paginated view
		limitStr := c.QueryParam("limit")
		page, _ := strconv.Atoi(pageStr)
		if page < 1 {
			page = 1
		}
		limit, _ := strconv.Atoi(limitStr)
		if limit < 1 {
			limit = DefaultUserLimit
		}
		users, _ := models.GetUsersByAccountID(actingUser.Account.ID, page, limit) // Ignoring error for simplicity here
		totalUsers, _ := models.CountUsersByAccountID(actingUser.Account.ID)
		totalPages := 0
		if totalUsers > 0 {
			totalPages = int(math.Ceil(float64(totalUsers) / float64(limit)))
		}
		baseView := NewBaseView(c)
		baseView.ActiveNav = "account"
		v := models.UsersView{
			BaseView:    baseView,
			Users:       users,
			CurrentPage: page,
			TotalPages:  totalPages,
			Limit:       limit,
		}
		return HTML(c, pages.UsersPartial(v)) // Return the partial component with the flash message header
	}

	// Call the model function to create the user
	_, err := models.CreateUser(username, email, password, actingUser.Account.ID)
	if err != nil {
		log.Printf("Error creating user by %d: %v", actingUser.ID, err)
		errMsg := "Failed to create user. Please try again."
		if strings.Contains(err.Error(), "UNIQUE constraint failed: users.email") {
			errMsg = "Email address is already in use."
		} else if strings.Contains(err.Error(), "UNIQUE constraint failed: users.username") {
			errMsg = "Username is already taken."
		}
		setFlashHeader(c, "error", errMsg)
		// Fetch and return the current list again to show the error
		pageStr := c.QueryParam("page")
		limitStr := c.QueryParam("limit")
		page, _ := strconv.Atoi(pageStr)
		if page < 1 {
			page = 1
		}
		limit, _ := strconv.Atoi(limitStr)
		if limit < 1 {
			limit = DefaultUserLimit
		}
		users, _ := models.GetUsersByAccountID(actingUser.Account.ID, page, limit)
		totalUsers, _ := models.CountUsersByAccountID(actingUser.Account.ID)
		totalPages := 0
		if totalUsers > 0 {
			totalPages = int(math.Ceil(float64(totalUsers) / float64(limit)))
		}
		baseView := NewBaseView(c)
		v := models.UsersView{
			BaseView:    baseView,
			Users:       users,
			CurrentPage: page,
			TotalPages:  totalPages,
			Limit:       limit,
		}
		return HTML(c, pages.UsersPartial(v))
	}

	// Success!
	setFlashHeader(c, "success", fmt.Sprintf("User '%s' created successfully.", username))

	// Fetch the updated list of users for the current/last page
	// Determine the page the new user will be on (usually the last page)
	pageStr := c.QueryParam("page") // Get current page if provided
	limitStr := c.QueryParam("limit")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1 // Default to first page if not specified
	}
	limit, _ := strconv.Atoi(limitStr)
	if limit < 1 {
		limit = DefaultUserLimit
	}

	totalUsers, err := models.CountUsersByAccountID(actingUser.Account.ID)
	if err != nil {
		log.Printf("Error counting users after creation: %v", err)
		// Handle error, maybe redirect or show a generic message
		return c.String(http.StatusInternalServerError, "Error retrieving updated user list")
	}
	totalPages := 0
	if totalUsers > 0 {
		totalPages = int(math.Ceil(float64(totalUsers) / float64(limit)))
	}

	// Optionally, redirect to the last page to see the new user
	// page = totalPages

	users, err := models.GetUsersByAccountID(actingUser.Account.ID, page, limit)
	if err != nil {
		log.Printf("Error fetching users after creation: %v", err)
		// Handle error
		return c.String(http.StatusInternalServerError, "Error retrieving updated user list")
	}

	baseView := NewBaseView(c)
	v := models.UsersView{
		BaseView:    baseView,
		Users:       users,
		CurrentPage: page,
		TotalPages:  totalPages,
		Limit:       limit,
	}

	// Return only the updated user list partial
	return HTML(c, pages.UsersPartial(v))
}
