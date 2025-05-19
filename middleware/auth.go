package middleware

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"

	"time"

	"github.com/jaredfolkins/letemcook/db"
	"github.com/jaredfolkins/letemcook/models"
	"github.com/jaredfolkins/letemcook/paths"
	"github.com/jaredfolkins/letemcook/util"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

// Before middleware initializes the custom context with user information
func Before(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cc := NewCustomContext(c)
		userCtx := &models.UserContext{} // Initialize an empty context
		var loggedInUser *models.User
		var actingAsUser *models.User
		var loggedInOK bool = false

		sess, err := session.Get("session", c) // Check for session errors
		if err != nil {
			// This likely means the secret changed or the cookie is corrupted.
			log.Printf("Error getting session, potentially invalid secret or cookie: %v. Clearing cookie and redirecting to login.", err)
			// Get a new, empty session to overwrite/clear the existing one
			newSess, _ := session.Get("session", c) // Ignore error on getting a *new* session
			newSess.Options.MaxAge = -1             // Expire the cookie
			newSess.Save(c.Request(), c.Response())

			// Redirect to login page for the default account (ID 1)
			squid, name, err := util.SquidAndNameByAccountID(1)
			if err != nil {
				log.Printf("Error getting squid/name for default account ID 1: %v. Redirecting to setup.", err)
				return c.Redirect(http.StatusFound, "/setup") // Redirect to setup if default account lookup fails
			}
			loginURL := fmt.Sprintf("%s?squid=%s&account=%s", paths.Login, squid, name)
			log.Printf("Redirecting to default login: %s", loginURL)
			return c.Redirect(http.StatusFound, loginURL)
		}

		loggedInUserIDVal := sess.Values["logged_in_user_id"]
		loggedInAccountIDVal := sess.Values["logged_in_account_id"]
		actingAsUserIDVal := sess.Values["acting_as_user_id"]
		actingAsAccountIDVal := sess.Values["acting_as_account_id"]

		if loggedInUserIDVal != nil && loggedInAccountIDVal != nil {
			loggedInUserID, okUser := loggedInUserIDVal.(int64)
			loggedInAccountID, okAccount := loggedInAccountIDVal.(int64)
			if okUser && okAccount {
				var err error
				loggedInUser, err = models.UserByIDAndAccountID(loggedInUserID, loggedInAccountID)
				if err != nil {
					log.Printf("Error fetching logged-in user %d for account %d from session: %v. Clearing session.", loggedInUserID, loggedInAccountID, err)
					delete(sess.Values, "logged_in_user_id")
					delete(sess.Values, "logged_in_account_id")
					delete(sess.Values, "acting_as_user_id")
					delete(sess.Values, "acting_as_account_id")
					sess.Options.MaxAge = -1             // Expire cookie
					sess.Save(c.Request(), c.Response()) // Save the cleared session
				} else if loggedInUser != nil {
					loggedInOK = true // Mark logged in user as successfully fetched
				}
			} else {
				log.Printf("Session data type assertion failed for logged-in user: userID type %T, accountID type %T", loggedInUserIDVal, loggedInAccountIDVal)
			}
		}

		if loggedInOK && actingAsUserIDVal != nil && actingAsAccountIDVal != nil {
			actingAsUserID, okUser := actingAsUserIDVal.(int64)
			actingAsAccountID, okAccount := actingAsAccountIDVal.(int64)
			if okUser && okAccount {
				if actingAsUserID == loggedInUser.ID && actingAsAccountID == loggedInUser.Account.ID {
					actingAsUser = loggedInUser
				} else {
					var err error
					actingAsUser, err = models.UserByIDAndAccountID(actingAsUserID, actingAsAccountID)
					if err != nil {
						log.Printf("Error fetching acting-as user %d for account %d from session: %v. Defaulting ActingAs to LoggedInAs.", actingAsUserID, actingAsAccountID, err)
						actingAsUser = loggedInUser // Default to logged-in user on error
					} else if actingAsUser == nil {
						log.Printf("Warning: Fetched nil acting-as user %d for account %d without error. Defaulting ActingAs to LoggedInAs.", actingAsUserID, actingAsAccountID)
						actingAsUser = loggedInUser
					}
				}
			} else {
				log.Printf("Session data type assertion failed for acting-as user: userID type %T, accountID type %T. Defaulting ActingAs to LoggedInAs.", actingAsUserIDVal, actingAsAccountIDVal)
				actingAsUser = loggedInUser // Default to logged-in user if types are wrong
			}
		} else if loggedInOK {
			actingAsUser = loggedInUser
		}

		if loggedInUser != nil && actingAsUser != nil {
			userCtx.LoggedInAs = loggedInUser
			userCtx.ActingAs = actingAsUser
		}
		cc.userContext = userCtx

		return next(cc)
	}
}

// BeforeNav middleware to refresh the top navigation
func BeforeNav(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cc, ok := c.(*lemcContext)
		if !ok {
			cc = NewCustomContext(c)
		}
		cc.Response().Header().Set("HX-Trigger", "refreshNavtop")
		return next(cc)
	}
}

// After middleware to save the session after the request
func After(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cc, ok := c.(*lemcContext)
		if !ok {
			cc = NewCustomContext(c)
		}

		sess, err := session.Get("session", cc)
		if err != nil {
			return err
		}

		err = sess.Save(cc.Request(), cc.Response())
		if err != nil {
			return err
		}

		return next(cc)
	}
}

// RedirIfNotSetup redirects to setup if the system is not configured
func RedirIfNotSetup(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		account, err := models.AccountByID(1)
		if err != nil || errors.Is(err, sql.ErrNoRows) || account == nil {
			return c.Redirect(http.StatusFound, "/setup")
		}
		return next(c)
	}
}

// RedirIfAuthd redirects authenticated users
func RedirIfAuthd(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		log.Println("redirIfAuthd")
		cc, ok := c.(LemcContext)
		if !ok {
			var ctxImpl *lemcContext
			ctxImpl, ok = c.(*lemcContext)
			if !ok {
				ctxImpl = NewCustomContext(c)
			}
			cc = ctxImpl
		}
		if cc.UserContext().LoggedInAs != nil {
			return cc.Redirect(http.StatusTemporaryRedirect, paths.Apps)
		}
		return next(c)
	}
}

// RedirIfNotAuthd redirects unauthenticated users to login
func RedirIfNotAuthd(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cc, ok := c.(LemcContext)
		if !ok {
			var ctxImpl *lemcContext
			ctxImpl, ok = c.(*lemcContext)
			if !ok {
				ctxImpl = NewCustomContext(c)
			}
			cc = ctxImpl
		}
		if cc.UserContext().LoggedInAs == nil {
			return cc.Redirect(http.StatusTemporaryRedirect, "/login")
		}
		return next(c)
	}
}

// ThemeMiddleware sets theme information based on user settings
func ThemeMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		theme := util.DefaultTheme
		cacheBuster := "ts"

		cc, ok := c.(LemcContext)
		if !ok {
			var ctxImpl *lemcContext
			ctxImpl, ok = c.(*lemcContext)
			if !ok {
				ctxImpl = NewCustomContext(c)
			}
			cc = ctxImpl
		}

		// Need to access the actual implementation to set theme/cacheBuster
		impl, ok := cc.(*lemcContext)
		if !ok {
			// This shouldn't happen given the code above, but just in case
			impl = NewCustomContext(c)
			cc = impl
		}
		impl.theme = theme
		impl.cacheBuster = cacheBuster

		userCtx := cc.UserContext()
		if userCtx != nil && userCtx.ActingAs != nil && userCtx.ActingAs.Account != nil {
			accountID := userCtx.ActingAs.Account.ID
			settings, err := models.GetAccountSettingsByAccountID(db.Db().DB, accountID)
			if err != nil {
				log.Printf("Warning: Error getting account settings for account %d in themeMiddleware: %v. Using default theme.", accountID, err)
			} else if settings != nil && settings.Theme != "" {
				log.Printf("Using account-specific theme: %s", settings.Theme)
				impl.theme = settings.Theme
				impl.cacheBuster = settings.Updated.Truncate(time.Second).Format("20060102150405")
				return next(cc)
			}
		}

		accountSquid := c.QueryParam("squid")
		if accountSquid != "" {
			account, err := models.AccountBySquid(accountSquid)
			if err != nil {
				log.Printf("Warning: Error getting account by squid %s in themeMiddleware: %v. Using default theme.", accountSquid, err)
				impl.theme = theme
				impl.cacheBuster = cacheBuster
				return next(cc)
			}

			accountSettings, err := models.GetAccountSettingsByAccountID(db.Db().DB, account.ID)
			if err != nil {
				log.Printf("Warning: Error getting account settings for account %d in themeMiddleware: %v. Using default theme.", account.ID, err)
			} else if accountSettings != nil && accountSettings.Theme != "" {
				theme = accountSettings.Theme
				cacheBuster = accountSettings.Updated.Truncate(time.Second).Format("20060102150405")
				impl.theme = theme
				impl.cacheBuster = cacheBuster
				return next(cc)
			}
		}

		impl.theme = theme
		impl.cacheBuster = cacheBuster
		return next(cc)
	}
}

// SetUserContext sets the user context in a lemcContext
// This should only be used for testing, normally the user context is set from the session
func SetUserContext(c echo.Context, userCtx *models.UserContext) LemcContext {
	cc, ok := c.(*lemcContext)
	if !ok {
		cc = NewCustomContext(c)
	}
	cc.userContext = userCtx
	return cc
}
