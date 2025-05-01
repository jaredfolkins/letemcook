package handlers

import (
	"database/sql"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/jaredfolkins/letemcook/db"
	"github.com/jaredfolkins/letemcook/models"
)

func listFiles(directory string) ([]string, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	var fileNames []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".mp3") {
			fileNames = append(fileNames, entry.Name())
		}
	}

	return fileNames, nil
}

func getRandomFileName(fileNames []string) string {
	return fileNames[rand.Intn(len(fileNames))]
}

func GetHeckle(c LemcContext) error {
	var accountID int64
	var userHeckleEnabled *bool // Use pointer to differentiate not set vs set to false

	userCtx := c.UserContext()
	if userCtx != nil && userCtx.ActingAs != nil {
		if userCtx.ActingAs.Account != nil {
			accountID = userCtx.ActingAs.Account.ID
			log.Printf("GetHeckle: Using account ID %d from user context.", accountID)
		}
		// Get heckle status directly from user object
		userHeckleEnabled = &userCtx.ActingAs.Heckle
		log.Printf("GetHeckle: User heckle setting: %v", *userHeckleEnabled)

		// If user has explicitly disabled heckle, return immediately
		if !*userHeckleEnabled {
			log.Printf("GetHeckle: Heckle disabled in user profile for user %d", userCtx.ActingAs.ID)
			return c.NoContent(http.StatusNoContent)
		}
	} else {
		// If not logged in or context missing, try getting from squid parameter
		squid := c.QueryParam("squid")
		if squid != "" {
			account, err := models.AccountBySquid(squid)
			if err == nil && account != nil {
				accountID = account.ID
				log.Printf("GetHeckle: Using account ID %d from squid parameter '%s'.", accountID, squid)
			} else {
				log.Printf("GetHeckle: Error looking up account by squid '%s': %v", squid, err)
			}
		} else {
			log.Println("GetHeckle: No user context and no squid parameter provided.")
		}
	}

	// If we couldn't determine an account ID, we can't get settings.
	if accountID == 0 {
		log.Println("GetHeckle: Could not determine account ID.")
		return c.NoContent(http.StatusNoContent)
	}

	// Check user-level settings first if we have a user context
	// var userSettings *models.UserSettings // REMOVED
	// if userCtx != nil && userCtx.ActingAs != nil { // REMOVED Block
	// 	var err error
	// 	log.Printf("GetHeckle: Checking user-level settings for user %d", userCtx.ActingAs.ID)
	// 	userID = userCtx.ActingAs.ID
	// 	userSettings, err = models.GetUserSettings(userID)
	//
	// 	log.Printf("GetHeckle: User settings: %+v", userSettings)
	// 	if err != nil {
	// 		if err == sql.ErrNoRows {
	// 			// No user settings found, continue with account settings check
	// 			log.Printf("GetHeckle: No user settings found for user %d, continuing", userID)
	// 		} else {
	// 			log.Printf("GetHeckle: Error fetching user settings for user %d: %v", userID, err)
	// 			return c.NoContent(http.StatusNoContent)
	// 		}
	// 	}
	// }

	// Now we have an accountID, proceed to get accountSettings
	accountSettings, err := models.GetAccountSettingsByAccountID(db.Db().DB, accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("GetHeckle: No account settings found for account %d. Heckle disabled.", accountID)
		} else {
			log.Printf("GetHeckle: Error fetching account settings for account %d: %v", accountID, err)
		}
		// If account settings don't exist or error, heckle is implicitly disabled (unless user enabled it)
		if userHeckleEnabled == nil || !*userHeckleEnabled { // Double check user setting if account setting missing
			return c.NoContent(http.StatusNoContent)
		}
		// If user enabled it, but account settings are missing/error, proceed based on user setting
	} else if !accountSettings.Heckle {
		// Account settings exist but heckle is disabled
		log.Printf("GetHeckle: Heckle disabled in account settings for account %d.", accountID)
		// If user didn't explicitly enable it, return no content
		if userHeckleEnabled == nil || !*userHeckleEnabled {
			return c.NoContent(http.StatusNoContent)
		}
		// User has explicitly enabled it, overriding account setting
		log.Printf("GetHeckle: User setting overrides disabled account setting for account %d.", accountID)
	}

	// REMOVED combined check block
	// if userSettings != nil && userSettings.Heckle == false {
	// 	// User has explicitly disabled heckle in their profile
	// 	log.Printf("GetHeckle: Heckle disabled in user settings for user %d", userID)
	// 	return c.NoContent(http.StatusNoContent)
	// }
	//
	// if (accountSettings == nil || !accountSettings.Heckle) && (userSettings == nil || !userSettings.Heckle) {
	// 	// Heckle is disabled in settings, or settings object is unexpectedly nil
	// 	return c.NoContent(http.StatusNoContent)
	// }

	// If we reach here, heckle is enabled either by account or user override.
	log.Printf("GetHeckle: Heckle is enabled for account %d (user override: %v).", accountID, userHeckleEnabled != nil && *userHeckleEnabled)

	// Heckle is enabled, proceed to find a file
	var heckleDir string
	if os.Getenv("LEMC_ENV") == "dev" || os.Getenv("LEMC_ENV") == "development" {
		heckleDir = filepath.Join("embedded", "assets", "heckle", "public")
	} else {
		heckleDir = filepath.Join("data", "assets", "heckle", "public")
	}

	fileNames, err := listFiles(heckleDir)
	if err != nil {
		log.Printf("GetHeckle: Error listing files in %s: %v", heckleDir, err)
		return c.NoContent(http.StatusNoContent)
	}

	if len(fileNames) == 0 {
		log.Printf("GetHeckle: No .mp3 files found in %s", heckleDir)
		return c.NoContent(http.StatusNoContent)
	}

	randomFileName := getRandomFileName(fileNames)
	encodedFileName := url.PathEscape(randomFileName)
	h := Heckle{}
	h.FileName = path.Join("/", "heckle", "public", encodedFileName) // Use forward slashes for URL path
	return c.JSON(http.StatusOK, h)
}

type Heckle struct {
	FileName string `json:"file"`
}
