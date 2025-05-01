package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/jaredfolkins/letemcook/db"
	"github.com/jaredfolkins/letemcook/models"
	"github.com/jaredfolkins/letemcook/views/partials"
)

func highlightMatch(text, search string) string {
	lowerText := strings.ToLower(text)
	lowerSearch := strings.ToLower(search)
	highlighted := ""

	for i := 0; i < len(text); i++ {
		if i+len(search) <= len(text) && lowerText[i:i+len(search)] == lowerSearch {
			highlighted += `<span class="bg-warning">` + text[i:i+len(search)] + `</span>`
			i += len(search) - 1
		} else {
			highlighted += string(text[i])
		}
	}

	return highlighted
}

func GetAclCookbookSearchHandler(c LemcContext) error {
	cb := &models.Cookbook{AccountID: c.UserContext().ActingAs.Account.ID}
	err := cb.ByUUID(c.Param("uuid"))
	if err != nil {
		return err
	}

	v := models.CoreView{
		Cookbook: cb,
		BaseView: NewBaseView(c),
	}

	search := c.FormValue("acl-search")
	if len(search) < 1 {
		dsr := partials.DisplayAclSearchResults(v)
		return HTML(c, dsr)
	}

	cba, err := models.SearchForCookbookAclUsersNotAssigned(search, c.UserContext().ActingAs.Account.ID, cb.ID, 100)
	if err != nil {
		log.Println("SearchForCookbookAclUsersNotAssigned: ", err)
	}

	for i, v := range cba {
		cba[i].Email = highlightMatch(v.Email, search)
		cba[i].Username = highlightMatch(v.Username, search)
	}

	v.CookbookAclSearchResults = cba
	dsr := partials.DisplayAclSearchResults(v)
	return HTML(c, dsr)
}

func GetAclAppSearchHandler(c LemcContext) error {
	appUUID := c.Param("uuid")
	accountID := c.UserContext().ActingAs.Account.ID
	app, err := models.AppByUUIDAndAccountID(appUUID, accountID)
	if err != nil {
		log.Printf("GetAclAppSearchHandler: Error finding app %s: %v", appUUID, err)
		return c.NoContent(http.StatusNotFound) // Or return HTML with empty results
	}

	v := models.CoreView{
		App:      app, // Use app field
		BaseView: NewBaseView(c),
	}

	search := c.FormValue("acl-search")
	if len(search) < 1 {
		csr := partials.DisplayAppAclSearchResults(v) // Need a app-specific partial
		return HTML(c, csr)
	}

	searchResults, err := models.SearchForappAclUsersNotAssigned(search, c.UserContext().ActingAs.Account.ID, app.ID, 100)
	if err != nil {
		log.Printf("GetAclAppSearchHandler: Error searching users for app %d: %v", app.ID, err)
	}

	for i := range searchResults {
		searchResults[i].Email = highlightMatch(searchResults[i].Email, search)
		searchResults[i].Username = highlightMatch(searchResults[i].Username, search)
	}

	v.AppAclSearchResults = searchResults         // Store results in CoreView
	csr := partials.DisplayAppAclSearchResults(v) // Use the app-specific partial
	return HTML(c, csr)
}

func PostAclUserToappHandler(c LemcContext) error {
	accountID := c.UserContext().ActingAs.Account.ID
	appUUID := c.Param("uuid")

	app, err := models.AppByUUIDAndAccountID(appUUID, accountID)
	if err != nil {
		log.Printf("PostAclUserToappHandler: Error finding app %s: %v", appUUID, err)
		c.AddErrorFlash("acl", "app not found")
		return c.NoContent(http.StatusNotFound)
	}

	userIDStr := c.Param("uid")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		log.Printf("PostAclUserToappHandler: Invalid user ID format '%s': %v", userIDStr, err)
		c.AddErrorFlash("acl", "Invalid user identifier")
		return c.NoContent(http.StatusBadRequest)
	}

	user, err := models.UserByIDAndAccountID(userID, accountID)
	if err != nil {
		log.Printf("PostAclUserToappHandler: Error finding user %d: %v", userID, err)
		c.AddErrorFlash("acl", "User to add not found")
		return c.NoContent(http.StatusNotFound)
	}

	pc := models.PermApp{
		UserID:        userID,
		AccountID:     accountID,
		AppID:         app.ID,
		CookbookID:    app.CookbookID, // Store the associated cookbook ID
		CanIndividual: true,
		CanShared:     false,
		CanAdminister: false,
		IsOwner:       false, // Only the creator should be owner
	}

	tx, err := db.Db().Beginx()
	if err != nil {
		log.Printf("PostAclUserToappHandler: Error starting transaction: %v", err)
		c.AddErrorFlash("acl", "Database error")
		return c.NoContent(http.StatusInternalServerError)
	}

	err = pc.UpsertappPermissions(tx)
	if err != nil {
		log.Printf("PostAclUserToappHandler: Error upserting app permission for user %d on app %d: %v", userID, app.ID, err)
		tx.Rollback()
		c.AddErrorFlash("acl", "Failed to add user permission")
		return c.NoContent(http.StatusInternalServerError)
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("PostAclUserToappHandler: Error committing transaction: %v", err)
		c.AddErrorFlash("acl", "Database commit error")
		return c.NoContent(http.StatusInternalServerError)
	}

	AppAcls, err := models.AppAclsUsers(accountID, app.ID)
	if err != nil {
		log.Printf("PostAclUserToappHandler: Error re-fetching ACLs after add: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	v := models.CoreView{
		BaseView: NewBaseView(c),
		App:      app,
		ViewType: "acls",
		AppAcls:  AppAcls,
	}

	c.AddSuccessFlash("acl", fmt.Sprintf("Added user [%s] to app ACL", user.Username))
	aclView := partials.AppAcls(v)
	return HTML(c, aclView)
}

func PostAclUserToCookbookHandler(c LemcContext) error {
	cb := &models.Cookbook{AccountID: c.UserContext().ActingAs.Account.ID}
	err := cb.ByUUID(c.Param("uuid"))
	if err != nil {
		return err
	}

	v := models.CoreView{Cookbook: cb, BaseView: NewBaseView(c)}

	user_id, err := strconv.ParseInt(c.Param("uid"), 10, 64)
	if err != nil {
		c.AddErrorFlash("acl", "failed to parse uid")
		return c.NoContent(http.StatusNotFound)
	}

	user, err := models.UserByIDAndAccountID(user_id, c.UserContext().ActingAs.Account.ID)
	if err != nil {
		c.AddErrorFlash("acl", "user not found")
		return c.NoContent(http.StatusNotFound)
	}

	pc := models.PermCookbook{
		UserID:     user.ID,
		AccountID:  cb.AccountID,
		CookbookID: cb.ID,
		CanView:    true,
		CanEdit:    false,
		IsOwner:    false,
	}

	tx, err := db.Db().Beginx()
	if err != nil {
		c.AddErrorFlash("acl", "tx error")
		return c.NoContent(http.StatusNotFound)
	}

	err = pc.UpsertCookbookPermissions(tx)
	if err != nil {
		log.Println("Error:", err)
		err = tx.Rollback()
		if err != nil {
			log.Println("🔥 Failed to rollback transaction: ", err)
		}
		c.AddErrorFlash("acl", "failed to update acl")
		return c.NoContent(http.StatusNotFound)
	}

	err = tx.Commit()
	if err != nil {
		log.Println("Error:", err)
		c.AddErrorFlash("acl", "tx commit failed")
		return c.NoContent(http.StatusNotFound)
	}

	cba, err := models.CookbookAclsUsers(c.UserContext().ActingAs.Account.ID, cb.ID)
	if err != nil {
		c.AddErrorFlash("acl", "failed to find acls")
		return c.NoContent(http.StatusNotFound)
	}

	v.YamlDefault.UUID = cb.UUID
	v.ViewType = "acls"
	v.CookbookAcls = cba

	c.AddSuccessFlash("acl", fmt.Sprintf("added [%s] to acl", user.Username))
	aclsView := partials.Acls(v)
	return HTML(c, aclsView)
}

func PutAclToogleEditCookbookHandler(c LemcContext) error {
	cb := &models.Cookbook{AccountID: c.UserContext().ActingAs.Account.ID}
	err := cb.ByUUID(c.Param("uuid"))
	if err != nil {
		return err
	}

	v := models.CoreView{Cookbook: cb, BaseView: NewBaseView(c)}

	user_id, err := strconv.ParseInt(c.Param("uid"), 10, 64)
	if err != nil {
		log.Println("Error:", err)
		c.AddErrorFlash("acl", "parsing parameters error")
		return c.NoContent(http.StatusConflict)
	}

	user, err := models.UserByIDAndAccountID(user_id, cb.AccountID)
	if err != nil {
		log.Println("Error:", err)
		c.AddErrorFlash("acl", "user not found")
		return c.NoContent(http.StatusNotFound)
	}

	pc := models.PermCookbook{}
	err = pc.CookbookPermissions(user.ID, cb.AccountID, cb.ID)
	if err != nil {
		log.Println("Error:", err)
		c.AddErrorFlash("acl", "cookbook not found")
		return c.NoContent(http.StatusNotFound)
	}

	tx, err := db.Db().Beginx()
	if err != nil {
		c.AddErrorFlash("acl", "tx error")
		return c.NoContent(http.StatusNotFound)
	}

	if pc.CanEdit {
		pc.CanEdit = false
	} else {
		pc.CanEdit = true
	}

	err = pc.UpdateCookbookPermissions(tx)
	if err != nil {
		log.Println("Error:", err)
		err = tx.Rollback()
		if err != nil {
			log.Println("🔥 Failed to rollback transaction: ", err)
		}
		c.AddErrorFlash("acl", "unable to update cookbook acls")
		return c.NoContent(http.StatusNotFound)
	}

	err = tx.Commit()
	if err != nil {
		log.Println("Error:", err)
		c.AddErrorFlash("acl", "tx commit failed")
		return c.NoContent(http.StatusNotFound)
	}

	cba, err := models.CookbookAclsUsers(c.UserContext().ActingAs.Account.ID, cb.ID)
	if err != nil {
		return err
	}

	v.YamlDefault.UUID = cb.UUID
	v.ViewType = "acls"
	v.CookbookAcls = cba

	aclsView := partials.Acls(v)
	c.AddSuccessFlash("acl", fmt.Sprintf("updated [%s] acl", user.Username))
	return HTML(c, aclsView)
}

func DeleteUserFromCookbookHandler(c LemcContext) error {
	cb := &models.Cookbook{AccountID: c.UserContext().ActingAs.Account.ID}
	err := cb.ByUUID(c.Param("uuid"))
	if err != nil {
		return err
	}
	v := models.CoreView{Cookbook: cb, BaseView: NewBaseView(c)}

	user_id, err := strconv.ParseInt(c.Param("uid"), 10, 64)
	if err != nil {
		c.AddErrorFlash("acl", "unable to parse uid ")
		return c.NoContent(http.StatusNotFound)
	}

	user, err := models.UserByIDAndAccountID(user_id, c.UserContext().ActingAs.Account.ID)
	if err != nil {
		c.AddErrorFlash("acl", "unable to find user")
		return c.NoContent(http.StatusNotFound)
	}

	pc := models.PermCookbook{
		UserID:     user.ID,
		AccountID:  cb.AccountID,
		CookbookID: cb.ID,
	}

	tx, err := db.Db().Beginx()
	if err != nil {
		return err
	}

	err = pc.DeleteCookbookPermissions(tx)
	if err != nil {
		log.Println("Error:", err)
		err = tx.Rollback()
		if err != nil {
			log.Println("🔥 Failed to rollback transaction: ", err)
		}
		c.AddErrorFlash("acl", "unable delete permission")
		return c.NoContent(http.StatusNotFound)
	}

	err = tx.Commit()
	if err != nil {
		log.Println("Error:", err)
		return err
	}

	cba, err := models.CookbookAclsUsers(c.UserContext().ActingAs.Account.ID, cb.ID)
	if err != nil {
		c.AddErrorFlash("acl", "unable to find acls")
		return c.NoContent(http.StatusNotFound)
	}

	v.YamlDefault.UUID = cb.UUID
	v.ViewType = "acls"
	v.CookbookAcls = cba

	c.AddSuccessFlash("acl", fmt.Sprintf("removed [%s] from acls", user.Username))
	aclsView := partials.Acls(v)
	return HTML(c, aclsView)
}

func PutappAclToggleIndividualHandler(c LemcContext) error {
	appUUID := c.Param("uuid")
	accountID := c.UserContext().ActingAs.Account.ID
	app, err := models.AppByUUIDAndAccountID(appUUID, accountID)
	if err != nil {
		log.Printf("PutappAclToggleIndividualHandler: Error finding app %s: %v", appUUID, err)
		return c.NoContent(http.StatusNotFound)
	}

	userID, err := strconv.ParseInt(c.Param("uid"), 10, 64)
	if err != nil {
		log.Printf("PutappAclToggleIndividualHandler: Error parsing user ID '%s': %v", c.Param("uid"), err)
		c.AddErrorFlash("acl", "Invalid user ID format")
		return c.NoContent(http.StatusBadRequest)
	}

	err = models.ToggleAppPermission(userID, accountID, app.ID, models.ToggleAppIndividual)
	if err != nil {
		log.Printf("PutappAclToggleIndividualHandler: Error toggling permission: %v", err)
		if strings.Contains(err.Error(), "permission record not found") {
			c.AddErrorFlash("acl", "Permissions not found for user")
			return c.NoContent(http.StatusNotFound)
		}
		c.AddErrorFlash("acl", "Failed to update permission")
		return c.NoContent(http.StatusInternalServerError)
	}

	appAcls, err := models.AppAclsUsers(accountID, app.ID)
	if err != nil {
		log.Printf("PutappAclToggleIndividualHandler: Error re-fetching ACLs: %v", err)
		c.AddErrorFlash("acl", "Error refreshing permissions view")
		return c.NoContent(http.StatusInternalServerError)
	}

	cookbook, err := models.CookbookByIDAndAccountID(app.CookbookID, accountID)
	if err != nil {
		log.Printf("PutappAclToggleIndividualHandler: Error fetching cookbook %d: %v", app.CookbookID, err)
		cookbook = nil
	}

	v := models.CoreView{
		App:         app,
		Cookbook:    cookbook,
		AppAcls:     appAcls,
		BaseView:    NewBaseView(c),
		ViewType:    "acls",
		YamlDefault: models.YamlDefault{UUID: app.UUID},
	}

	c.AddSuccessFlash("acl", "Individual permission updated")
	aclView := partials.AppAcls(v)
	return HTML(c, aclView)
}

func PutappAclToggleSharedHandler(c LemcContext) error {
	appUUID := c.Param("uuid")
	accountID := c.UserContext().ActingAs.Account.ID
	app, err := models.AppByUUIDAndAccountID(appUUID, accountID)
	if err != nil {
		log.Printf("PutappAclToggleSharedHandler: Error finding app %s: %v", appUUID, err)
		return c.NoContent(http.StatusNotFound)
	}

	userID, err := strconv.ParseInt(c.Param("uid"), 10, 64)
	if err != nil {
		log.Printf("PutappAclToggleSharedHandler: Error parsing user ID '%s': %v", c.Param("uid"), err)
		c.AddErrorFlash("acl", "Invalid user ID format")
		return c.NoContent(http.StatusBadRequest)
	}

	err = models.ToggleAppPermission(userID, accountID, app.ID, models.ToggleAppShared)
	if err != nil {
		log.Printf("PutappAclToggleSharedHandler: Error toggling permission: %v", err)
		if strings.Contains(err.Error(), "permission record not found") {
			c.AddErrorFlash("acl", "Permissions not found for user")
			return c.NoContent(http.StatusNotFound)
		}
		c.AddErrorFlash("acl", "Failed to update permission")
		return c.NoContent(http.StatusInternalServerError)
	}

	appAcls, err := models.AppAclsUsers(accountID, app.ID)
	if err != nil {
		log.Printf("PutappAclToggleSharedHandler: Error re-fetching ACLs: %v", err)
		c.AddErrorFlash("acl", "Error refreshing permissions view")
		return c.NoContent(http.StatusInternalServerError)
	}

	cookbook, err := models.CookbookByIDAndAccountID(app.CookbookID, accountID)
	if err != nil {
		log.Printf("PutappAclToggleSharedHandler: Error fetching cookbook %d: %v", app.CookbookID, err)
		cookbook = nil
	}

	v := models.CoreView{
		App:         app,
		Cookbook:    cookbook,
		AppAcls:     appAcls,
		BaseView:    NewBaseView(c),
		ViewType:    "acls",
		YamlDefault: models.YamlDefault{UUID: app.UUID},
	}

	c.AddSuccessFlash("acl", "Shared permission updated")
	aclView := partials.AppAcls(v)
	return HTML(c, aclView)
}

func PutappAclToggleAdminHandler(c LemcContext) error {
	appUUID := c.Param("uuid")
	accountID := c.UserContext().ActingAs.Account.ID
	app, err := models.AppByUUIDAndAccountID(appUUID, accountID)
	if err != nil {
		log.Printf("PutappAclToggleAdminHandler: Error finding app %s: %v", appUUID, err)
		return c.NoContent(http.StatusNotFound)
	}

	userID, err := strconv.ParseInt(c.Param("uid"), 10, 64)
	if err != nil {
		log.Printf("PutappAclToggleAdminHandler: Error parsing user ID '%s': %v", c.Param("uid"), err)
		c.AddErrorFlash("acl", "Invalid user ID format")
		return c.NoContent(http.StatusBadRequest)
	}

	err = models.ToggleAppPermission(userID, accountID, app.ID, models.ToggleAppAdminister)
	if err != nil {
		log.Printf("PutappAclToggleAdminHandler: Error toggling permission: %v", err)
		if strings.Contains(err.Error(), "permission record not found") {
			c.AddErrorFlash("acl", "Permissions not found for user")
			return c.NoContent(http.StatusNotFound)
		}
		c.AddErrorFlash("acl", "Failed to update permission")
		return c.NoContent(http.StatusInternalServerError)
	}

	appAcls, err := models.AppAclsUsers(accountID, app.ID)
	if err != nil {
		log.Printf("PutappAclToggleAdminHandler: Error re-fetching ACLs: %v", err)
		c.AddErrorFlash("acl", "Error refreshing permissions view")
		return c.NoContent(http.StatusInternalServerError)
	}

	cookbook, err := models.CookbookByIDAndAccountID(app.CookbookID, accountID)
	if err != nil {
		log.Printf("PutappAclToggleAdminHandler: Error fetching cookbook %d: %v", app.CookbookID, err)
		cookbook = nil
	}

	v := models.CoreView{
		App:         app,
		Cookbook:    cookbook,
		AppAcls:     appAcls,
		BaseView:    NewBaseView(c),
		ViewType:    "acls",
		YamlDefault: models.YamlDefault{UUID: app.UUID},
	}

	c.AddSuccessFlash("acl", "Admin permission updated")
	aclView := partials.AppAcls(v)
	return HTML(c, aclView)
}

func DeleteUserFromappHandler(c LemcContext) error {
	accountID := c.UserContext().ActingAs.Account.ID
	appUUID := c.Param("uuid")

	app, err := models.AppByUUIDAndAccountID(appUUID, accountID)
	if err != nil {
		log.Printf("DeleteUserFromappHandler: Error finding app %s: %v", appUUID, err)
		c.AddErrorFlash("acl", "app not found")
		return c.NoContent(http.StatusNotFound)
	}

	userIDStr := c.Param("uid")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		log.Printf("DeleteUserFromappHandler: Invalid user ID format '%s': %v", userIDStr, err)
		c.AddErrorFlash("acl", "Invalid user identifier")
		return c.NoContent(http.StatusBadRequest)
	}

	user, err := models.UserByIDAndAccountID(userID, accountID)
	if err != nil {
		log.Printf("DeleteUserFromappHandler: Error finding user %d: %v", userID, err)
		c.AddErrorFlash("acl", "User to remove not found")
		return c.NoContent(http.StatusNotFound)
	}

	if app.OwnerID == userID {
		log.Printf("DeleteUserFromappHandler: Attempt to delete app owner (User ID: %d, app ID: %d)", userID, app.ID)
		c.AddErrorFlash("acl", "Cannot remove the app owner.")

		AppAcls, err := models.AppAclsUsers(accountID, app.ID)
		if err != nil {
			log.Printf("DeleteUserFromappHandler: Error re-fetching ACLs after owner delete attempt: %v", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		v := models.CoreView{
			BaseView: NewBaseView(c),
			App:      app,
			ViewType: "acls",
			AppAcls:  AppAcls,
		}
		aclsView := partials.AppAcls(v)
		return HTML(c, aclsView) // Option 2: Rely on flash message
	}

	pc := models.PermApp{
		UserID:    userID,
		AccountID: accountID,
		AppID:     app.ID, // Use the app's internal ID
	}

	tx, err := db.Db().Beginx()
	if err != nil {
		log.Printf("DeleteUserFromappHandler: Error starting transaction: %v", err)
		c.AddErrorFlash("acl", "Database error")
		return c.NoContent(http.StatusInternalServerError)
	}

	err = pc.DeleteappPermissions(tx) // Assuming this model function exists
	if err != nil {
		log.Printf("DeleteUserFromappHandler: Error deleting app permission for user %d on app %d: %v", userID, app.ID, err)
		tx.Rollback()
		c.AddErrorFlash("acl", "Failed to remove user permission")
		return c.NoContent(http.StatusInternalServerError)
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("DeleteUserFromappHandler: Error committing transaction: %v", err)
		c.AddErrorFlash("acl", "Database commit error")
		return c.NoContent(http.StatusInternalServerError)
	}

	AppAcls, err := models.AppAclsUsers(accountID, app.ID)
	if err != nil {
		log.Printf("DeleteUserFromappHandler: Error re-fetching ACLs after delete: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	v := models.CoreView{
		BaseView: NewBaseView(c),
		App:      app, // Pass the fetched app
		ViewType: "acls",
		AppAcls:  AppAcls, // Pass the updated list
	}

	c.AddSuccessFlash("acl", fmt.Sprintf("Removed user [%s] from app ACL", user.Username))
	aclsView := partials.AppAcls(v) // Assuming a partials.AppAcls exists
	return HTML(c, aclsView)
}

// PatchAppOnRegisterToggleHandler toggles the OnRegister flag for a given app.
func PatchAppOnRegisterToggleHandler(c LemcContext) error {
	appUUID := c.Param("uuid")
	accountID := c.UserContext().ActingAs.Account.ID
	app, err := models.AppByUUIDAndAccountID(appUUID, accountID)
	if err != nil {
		log.Printf("PatchAppOnRegisterToggleHandler: Error finding app %s: %v", appUUID, err)
		c.AddErrorFlash("app", "App not found.")
		return c.NoContent(http.StatusNotFound)
	}

	// Toggle the OnRegister status
	newStatus := !app.OnRegister

	// Update the database
	query := `UPDATE apps SET on_register = ?, updated = CURRENT_TIMESTAMP WHERE id = ?`
	_, err = db.Db().Exec(query, newStatus, app.ID)
	if err != nil {
		log.Printf("PatchAppOnRegisterToggleHandler: Error updating app %d OnRegister status: %v", app.ID, err)
		c.AddErrorFlash("app", "Failed to update On Register setting.")
		return c.NoContent(http.StatusInternalServerError)
	}

	log.Printf("PatchAppOnRegisterToggleHandler: Toggled OnRegister for app %s (ID: %d) to %t", appUUID, app.ID, newStatus)
	c.AddSuccessFlash("app", "On Register setting updated successfully.")
	// Success, no content to return as per hx-swap="none"
	return c.NoContent(http.StatusNoContent)
}
