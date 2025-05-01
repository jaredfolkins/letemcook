package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/jaredfolkins/letemcook/models"
	"github.com/labstack/echo/v4"
)

// ApplyMiddlewares applies multiple middleware functions in reverse order
func ApplyMiddlewares(h echo.HandlerFunc, m ...echo.MiddlewareFunc) echo.HandlerFunc {
	for i := len(m) - 1; i >= 0; i-- {
		if m[i] != nil {
			h = m[i](h)
		}
	}
	return h
}

// CheckPermission returns an echo.MiddlewareFunc that checks if the user has
// AT LEAST ONE of the required permissions to access the route.
func CheckPermission(requiredPermissions ...models.Permission) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var err error
			objectUUID := c.Param("uuid") // UUID of the object being accessed (cookbook, app) if applicable
			hasPermission := false        // Assume no permission initially

			cc, ok := c.(LemcContext)
			if !ok {
				var ctxImpl *lemcContext
				ctxImpl, ok = c.(*lemcContext)
				if !ok {
					ctxImpl = NewCustomContext(c)
				}
				cc = ctxImpl
			}

			user := cc.UserContext().ActingAs // Use ActingAs as per recent change
			if user == nil {
				log.Printf("Warning: Unauthenticated user reached permission-protected route: %s", c.Path())
				return cc.Redirect(http.StatusTemporaryRedirect, "/login")
			}

			accountID := user.Account.ID // Assumes User has an Account field/struct with ID

			// Convert []models.Permission to []string for logging
			permStrings := make([]string, len(requiredPermissions))
			for i, p := range requiredPermissions {
				permStrings[i] = string(p)
			}
			log.Printf("User %d (Account %d) attempting action requiring ONE OF permissions: [%s] on path %s",
				user.ID, accountID, strings.Join(permStrings, ", "), c.Path())

			var checkResult bool
			// Check each required permission
			for _, requiredPermission := range requiredPermissions {
				switch requiredPermission {
				case models.CanEditCookbook:
					if objectUUID == "" {
						log.Printf("Error: Missing UUID parameter for permission check (%s) on path %s", requiredPermission, c.Path())
					}

					checkResult, err = models.HasCookbookPermission(user.ID, accountID, objectUUID, requiredPermission)
					if err != nil {
						log.Printf("Error checking cookbook permission %s for user %d on object %s: %v", requiredPermission, user.ID, objectUUID, err)
					}

				case models.CanSharedApp, models.CanIndividualApp, models.CanAclApp:
					if objectUUID == "" {
						log.Printf("Error: Missing UUID parameter for permission check (%s) on path %s", requiredPermission, c.Path())
					}
					checkResult, err = models.HasAppPermission(user.ID, accountID, objectUUID, requiredPermission)
					if err != nil {
						log.Printf("Error checking app permission %s for user %d on object %s: %v", requiredPermission, user.ID, objectUUID, err)
					}

				case models.CanAccessCookbooksView, models.CanEditApp, models.CanAccessAppsView, models.CanCreateCookbook, models.CanCreateApp, models.CanAdministerAccount:
					checkResult, err = models.HasAccountPermission(user.ID, accountID, requiredPermission)
					if err != nil {
						log.Printf("Error checking account permission %s for user %d: %v", requiredPermission, user.ID, err)
					}

				case models.CanAdministerSystem:
					checkResult, err = models.HasSystemPermission(user.ID, requiredPermission)
					if err != nil {
						log.Printf("Error checking system permission %s for user %d: %v", requiredPermission, user.ID, err)
					}

				default:
					log.Printf("Error: Unhandled permission type '%s' in checkPermission middleware for path %s", requiredPermission, c.Path())
				}

				// If this permission check is successful, grant access and stop checking others
				if checkResult {
					hasPermission = true
					log.Printf("Permission GRANTED for user %d (Account %d) via permission '%s' (required one of: [%s]) on path %s",
						user.ID, accountID, requiredPermission, strings.Join(permStrings, ", "), c.Path())
					break // Exit the loop as soon as one permission matches
				}
			} // End of permission loop

			if !hasPermission {
				log.Printf("Permission DENIED for user %d (Account %d): required ONE OF [%s] on path %s",
					user.ID, accountID, strings.Join(permStrings, ", "), c.Path())
				// Consider adding a more specific error message if needed, e.g., which object they tried to access.
				cc.AddErrorFlash("permission_denied", "Forbidden: You do not have the necessary permissions to perform this action.")
				return echo.NewHTTPError(http.StatusForbidden, "Forbidden: You do not have the necessary permissions to perform this action.")
			}

			// If hasPermission is true, proceed to the next handler
			return next(c)
		}
	}
}
