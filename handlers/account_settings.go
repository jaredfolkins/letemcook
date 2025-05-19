package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/a-h/templ"
	"github.com/jaredfolkins/letemcook/db"
	"github.com/jaredfolkins/letemcook/models"
	"github.com/jaredfolkins/letemcook/paths"
	"github.com/jaredfolkins/letemcook/util"
	"github.com/jaredfolkins/letemcook/views/pages"
)

const (
	themeNameRegex = `^[a-zA-Z0-9\-]+$` // Allow letters, numbers, hyphen
)

var validThemeName = regexp.MustCompile(themeNameRegex)

func GetAccountSettingsHandler(c LemcContext) error {
	viewData, settingsComponent, err := partialAccountSettingsHandler(c)
	if err != nil {
		log.Printf("Error getting account settings: %v", err)
		return err
	}

	partial := strings.ToLower(c.QueryParam("partial"))
	if partial == "true" {
		return HTML(c, settingsComponent)
	}

	indexView := pages.AccountSettingsIndex(viewData, settingsComponent)
	return HTML(c, indexView)
}

func partialAccountSettingsHandler(c LemcContext) (models.AccountSettingsView, templ.Component, error) {
	user := c.UserContext().ActingAs
	if user == nil {
		log.Println("Error: GetAccountSettingsHandler called without authenticated user")
		return models.AccountSettingsView{}, nil, fmt.Errorf("user not authenticated")
	}

	accountID := user.Account.ID

	settings, err := models.GetAccountSettingsByAccountID(db.Db().DB, accountID)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error getting account settings for account %d: %v", accountID, err)
		return models.AccountSettingsView{}, nil, err
	}

	if settings == nil {
		settings = &models.AccountSettings{
			AccountID:    accountID,
			Theme:        util.DefaultTheme,
			Registration: false,
			Heckle:       false,
		}
		log.Printf("No settings found for account %d. Creating default settings.", accountID)
		err = models.UpsertAccountSettings(db.Db(), settings)
		if err != nil {
			log.Printf("Error saving default settings for account %d: %v", accountID, err)
			return models.AccountSettingsView{}, nil, err
		}
	}

	var themesDir string
	if os.Getenv("LEMC_ENV") == "dev" || os.Getenv("LEMC_ENV") == "development" {
		themesDir = filepath.Join("embedded", "assets", "themes")
	} else {
		themesDir = filepath.Join("data", "assets", "themes")
	}

	availableThemes := []string{}
	entries, err := os.ReadDir(themesDir)
	if err != nil {
		log.Printf("Error reading themes directory '%s': %v. Proceeding without dynamic themes.", themesDir, err)
	} else {
		for _, entry := range entries {
			if entry.IsDir() && validThemeName.MatchString(entry.Name()) {
				availableThemes = append(availableThemes, entry.Name())
			}
		}
	}

	foundCurrent := false
	for _, theme := range availableThemes {
		if theme == settings.Theme {
			foundCurrent = true
			break
		}
	}

	if !foundCurrent && settings.Theme != "" {
		availableThemes = append([]string{settings.Theme}, availableThemes...)
	}

	baseView := NewBaseViewWithSquidAndAccountName(c, user.Account.Squid, user.Account.Name)
	baseView.ActiveNav = "account"
	baseView.ActiveSubNav = paths.AccountSettings
	viewData := models.AccountSettingsView{
		BaseView:        baseView,
		Settings:        settings,
		AvailableThemes: availableThemes, // Pass the list of themes
	}

	settingsComponent := pages.AccountSettings(viewData)
	return viewData, settingsComponent, nil
}

func PostAccountSettingsHandler(c LemcContext) error {
	user := c.UserContext().ActingAs
	if user == nil {
		log.Println("Error: PostAccountSettingsHandler called without authenticated user")
		return c.Redirect(http.StatusTemporaryRedirect, "/login")
	}

	accountID := user.Account.ID

	currentSettings, err := models.GetAccountSettingsByAccountID(db.Db().DB, accountID) // Use .DB
	if err != nil && err != sql.ErrNoRows {                                             // Use sql.ErrNoRows
		log.Printf("Error fetching settings before update for account %d: %v", accountID, err)
		c.AddErrorFlash("settings-update", "Failed to load existing settings before update.")
		return c.Redirect(http.StatusSeeOther, paths.AccountSettings) // Redirect back
	}

	if currentSettings == nil {
		currentSettings = &models.AccountSettings{
			AccountID: accountID,
		}
	}

	currentSettings.Theme = c.FormValue("theme")
	currentSettings.Registration = c.FormValue("registration") == "on" // Checkbox value
	currentSettings.Heckle = c.FormValue("heckle") == "on"             // Checkbox value

	err = models.UpsertAccountSettings(db.Db(), currentSettings) // Use .DB
	if err != nil {
		log.Printf("Error upserting account settings for account %d: %v", accountID, err)
		c.AddErrorFlash("settings-update", "Failed to save settings.")
		return c.Redirect(http.StatusSeeOther, paths.AccountSettings) // Redirect back
	}

	log.Printf("Successfully updated settings for account %d", accountID)
	c.AddSuccessFlash("settings-update", "Settings updated successfully!")

	_, settingsComponent, err := partialAccountSettingsHandler(c)
	if err != nil {
		log.Printf("Error getting account settings: %v", err)
		return err
	}

	return HTML(c, settingsComponent)
}
