package handlers

import (
	"log"
	"os"

	"github.com/jaredfolkins/letemcook/models"
)

// canShowAppsNav checks if the user has permissions to see the Apps nav item.
func canShowAppsNav(userCtx *models.UserContext) bool {
	// Early return for invalid user context
	if userCtx == nil || userCtx.ActingAs == nil || userCtx.ActingAs.Account == nil || userCtx.ActingAs.Account.ID == 0 {
		return false
	}

	accountID := userCtx.ActingAs.Account.ID

	// Early return if no permissions exist
	if userCtx.ActingAs.Permissions == nil {
		return false
	}

	// Check system admin permission first
	if userCtx.ActingAs.Permissions.PermSystem != nil && userCtx.ActingAs.Permissions.PermSystem.CanAdminister {
		return true
	}

	// Check account permissions
	return hasAccountPermissionForApps(userCtx.ActingAs.Permissions.PermissionsAccounts, accountID)
}

// hasAccountPermissionForApps checks if user has app-related permissions for the account
func hasAccountPermissionForApps(permissions []*models.PermAccount, accountID int64) bool {
	for _, perm := range permissions {
		if perm.AccountID == accountID && (perm.CanAdminister || perm.CanViewapps) {
			return true
		}
	}
	return false
}

// canShowCookbooksNav checks if the user has permissions to see the Cookbooks nav item.
func canShowCookbooksNav(userCtx *models.UserContext) bool {
	// Early return for invalid user context
	if userCtx == nil || userCtx.ActingAs == nil || userCtx.ActingAs.Account == nil || userCtx.ActingAs.Account.ID == 0 {
		return false
	}

	accountID := userCtx.ActingAs.Account.ID

	// Early return if no permissions exist
	if userCtx.ActingAs.Permissions == nil {
		return false
	}

	// Check system admin permission first
	if userCtx.ActingAs.Permissions.PermSystem != nil &&
		(userCtx.ActingAs.Permissions.PermSystem.CanAdminister || userCtx.ActingAs.Permissions.PermSystem.IsOwner) {
		return true
	}

	// Check account permissions
	return hasAccountPermissionForCookbooks(userCtx.ActingAs.Permissions.PermissionsAccounts, accountID)
}

// hasAccountPermissionForCookbooks checks if user has cookbook-related permissions for the account
func hasAccountPermissionForCookbooks(permissions []*models.PermAccount, accountID int64) bool {
	for _, perm := range permissions {
		if perm.AccountID == accountID && (perm.CanAdminister || perm.CanViewCookbooks || perm.IsOwner) {
			return true
		}
	}
	return false
}

// canShowAccountNav checks if the user has permission to administer the current account.
func canShowAccountNav(userCtx *models.UserContext) bool {
	// Must have a valid user context, acting user, account, and permissions.
	if userCtx == nil || userCtx.ActingAs == nil || userCtx.ActingAs.Account == nil || userCtx.ActingAs.Permissions == nil {
		return false
	}

	accountID := userCtx.ActingAs.Account.ID

	// Check account-specific permissions.
	for _, perm := range userCtx.ActingAs.Permissions.PermissionsAccounts {
		if perm.AccountID == accountID {
			// Show if the user can administer this account or is the owner.
			return perm.CanAdminister || perm.IsOwner
		}
	}

	// If no matching account permission found.
	return false
}

// canShowSystemNav checks if the user has permission to see the System nav item.
func canShowSystemNav(userCtx *models.UserContext) bool {
	// Must have a valid user context, acting user, and permissions.
	if userCtx == nil || userCtx.ActingAs == nil || userCtx.ActingAs.Permissions == nil || userCtx.ActingAs.Permissions.PermSystem == nil {
		return false
	}
	// Show if the user can administer the system or is a system owner.
	return userCtx.ActingAs.Permissions.PermSystem.CanAdminister || userCtx.ActingAs.Permissions.PermSystem.IsOwner
}

// BaseViewOption is a functional option for configuring BaseView
type BaseViewOption func(*models.BaseView)

// WithTitle sets the title of the BaseView
func WithTitle(title string) BaseViewOption {
	return func(bv *models.BaseView) {
		bv.Title = title
	}
}

// WithActiveNav sets the active navigation item
func WithActiveNav(activeNav string) BaseViewOption {
	return func(bv *models.BaseView) {
		bv.ActiveNav = activeNav
	}
}

// WithActiveSubNav sets the active sub-navigation item
func WithActiveSubNav(activeSubNav string) BaseViewOption {
	return func(bv *models.BaseView) {
		bv.ActiveSubNav = activeSubNav
	}
}

// WithNavigation sets both active navigation and sub-navigation items
func WithNavigation(activeNav, activeSubNav string) BaseViewOption {
	return func(bv *models.BaseView) {
		bv.ActiveNav = activeNav
		bv.ActiveSubNav = activeSubNav
	}
}

func NewBaseView(c LemcContext, opts ...BaseViewOption) models.BaseView {
	lemcEnv := os.Getenv("LEMC_ENV")

	// Default registrationEnabled to false
	registrationEnabled := false
	var err error

	userCtx := c.UserContext()
	showAppsNav := canShowAppsNav(userCtx)
	showCookbooksNav := canShowCookbooksNav(userCtx)
	showAccountNav := canShowAccountNav(userCtx)
	showSystemNav := canShowSystemNav(userCtx) // Calculate using the new function

	// Check if UserContext and ActingAs are valid before fetching settings
	if userCtx != nil && userCtx.ActingAs != nil && userCtx.ActingAs.Account != nil && userCtx.ActingAs.Account.ID != 0 {
		accountID := userCtx.ActingAs.Account.ID
		// Use the model function to fetch registration settings
		registrationEnabled, err = models.AccountSettingsByAccountID(accountID)
		if err != nil {
			// Log error but continue with registration disabled
			log.Printf("Error fetching registration setting in NewBaseView: %v", err)
			registrationEnabled = false // Explicitly set to false on error
		}
	}

	baseView := models.BaseView{
		Theme:               c.Theme(),
		CacheBuster:         c.CacheBuster(),
		UserContext:         userCtx,
		Env:                 lemcEnv,
		RegistrationEnabled: registrationEnabled,
		ShowAppsNav:         showAppsNav,
		ShowCookbooksNav:    showCookbooksNav,
		ShowAccountNav:      showAccountNav,
		ShowSystemNav:       showSystemNav, // Set the calculated value
		ActiveNav:           "",
		ActiveSubNav:        "",
	}

	// Apply functional options
	for _, opt := range opts {
		opt(&baseView)
	}

	return baseView
}

func NewBaseViewWithSquidAndAccountName(c LemcContext, squid string, name string, opts ...BaseViewOption) models.BaseView {
	lemcEnv := os.Getenv("LEMC_ENV")

	// Default registrationEnabled to false
	registrationEnabled := false
	var err error
	userCtx := c.UserContext()
	showAppsNav := canShowAppsNav(userCtx)
	showCookbooksNav := canShowCookbooksNav(userCtx)
	showAccountNav := canShowAccountNav(userCtx)
	showSystemNav := canShowSystemNav(userCtx) // Calculate using the new function

	// Find the account using the model function
	account, err := models.AccountBySquid(squid) // Use the new model function
	if err != nil || account == nil {
		log.Printf("Could not find account by squid '%s' in NewBaseViewWithSquidAndAccountName: %v", squid, err)
		// Keep registrationEnabled as false
	} else {
		// Use the model function to fetch registration settings
		registrationEnabled, err = models.AccountSettingsByAccountID(account.ID)
		if err != nil {
			// Log error but continue with registration disabled
			log.Printf("Error fetching registration setting in NewBaseViewWithSquidAndAccountName: %v", err)
			registrationEnabled = false // Explicitly set to false on error
		}
	}

	baseView := models.BaseView{
		AccountSquid:        squid,
		AccountName:         name,
		Title:               name,
		Theme:               c.Theme(),
		CacheBuster:         c.CacheBuster(),
		UserContext:         userCtx,
		Env:                 lemcEnv,
		RegistrationEnabled: registrationEnabled,
		ShowAppsNav:         showAppsNav,
		ShowCookbooksNav:    showCookbooksNav,
		ShowAccountNav:      showAccountNav,
		ShowSystemNav:       showSystemNav, // Set the calculated value
		ActiveNav:           "",
		ActiveSubNav:        "",
	}

	// Apply functional options
	for _, opt := range opts {
		opt(&baseView)
	}

	return baseView
}
