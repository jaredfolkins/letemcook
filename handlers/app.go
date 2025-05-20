package handlers

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/jaredfolkins/letemcook/db"
	"github.com/jaredfolkins/letemcook/models"
	"github.com/jaredfolkins/letemcook/paths"
	"github.com/jaredfolkins/letemcook/util"
	"github.com/jaredfolkins/letemcook/views/pages"
	"gopkg.in/yaml.v3"
)

func GetAppIndexIndividualHandler(c LemcContext) error {
	appUUID := c.Param("uuid")
	partial := strings.ToLower(c.QueryParam("partial"))
	accountID := c.UserContext().ActingAs.Account.ID

	app, err := models.AppByUUIDAndAccountID(appUUID, accountID)
	if err != nil {
		log.Printf("Error fetching app by UUID %s for account %d: %v", appUUID, accountID, err)
		if errors.Is(err, sql.ErrNoRows) {
			return c.String(http.StatusNotFound, "app not found")
		}
		return c.String(http.StatusInternalServerError, "Error retrieving app")
	}

	// Fetch specific user permissions for this app using the model function
	userID := c.UserContext().ActingAs.ID
	appPerms, err := models.AppPermissionsByUserAccountAndApp(userID, accountID, app.ID)
	if err != nil {
		// Handle potential database errors returned by the model function
		log.Printf("Error fetching app permissions via model for user %d, app %d: %v", userID, app.ID, err)
		return c.String(http.StatusInternalServerError, "Error retrieving app permissions")
	}
	// Assign the permissions (even if they are the default zero-value ones)
	app.UserPerms = appPerms

	cb, err := models.CookbookByIDAndAccountID(app.CookbookID, accountID)
	if err != nil {
		log.Printf("Error fetching cookbook by ID %d for account %d (from app %s): %v", app.CookbookID, accountID, appUUID, err)
		if errors.Is(err, sql.ErrNoRows) {
			return c.String(http.StatusNotFound, "Associated cookbook not found")
		}
		return c.String(http.StatusInternalServerError, "Error retrieving associated cookbook")
	}

	cb.UUID = app.UUID
	cb.Description = app.Description
	cb.YamlShared = app.YAMLShared
	cb.YamlIndividual = app.YAMLIndividual

	v := models.CoreView{
		App:      app,
		Cookbook: cb,
		BaseView: NewBaseView(c),
	}

	var isAdmin bool
	yaml.Unmarshal([]byte(cb.YamlIndividual), &v.YamlDefault)
	v.ViewType = "individual"
	v.YamlDefault.UUID = cb.UUID // Ensure UUID is set even if YAML unmarshal didn't happen

	for _, page := range v.YamlDefault.Cookbook.Pages {
		jm := &util.JobMeta{
			UUID:     v.YamlDefault.UUID,
			PageID:   strconv.Itoa(page.PageID),
			UserID:   strconv.FormatInt(c.UserContext().ActingAs.ID, 10),
			Username: c.UserContext().ActingAs.Username,
		}

		cf, err := util.NewContainerFiles(jm, isAdmin)
		if err != nil {
			return err
		}

		err = cf.OpenFiles()
		if err != nil {
			return err
		}
		defer cf.CloseFiles()

		css, err := cf.Read(cf.Css)
		if err != nil {
			return err
		}
		page.CssCache = fmt.Sprintf("<style id='uuid-%s-pageid-%d-scope-%s-style'>%s</style>", cb.UUID, page.PageID, v.ViewType, css)

		html, err := cf.Read(cf.Html)
		if err != nil {
			return err
		}
		page.HtmlCache = fmt.Sprintf("<div id='uuid-%s-pageid-%d-scope-%s-html' class='page-inner'>%s</div>", cb.UUID, page.PageID, v.ViewType, html)

		js, err := cf.Read(cf.Js)
		if err != nil {
			return err
		}
		page.JsCache = fmt.Sprintf("<script id='uuid-%s-pageid-%d-scope-%s-script'>%s</script>", cb.UUID, page.PageID, v.ViewType, js)

		v.YamlDefault.Cookbook.Pages[page.PageID-1] = page
	}

	for _, wiki := range v.YamlDefault.Cookbook.Storage.Wikis {
		b64, err := base64.StdEncoding.DecodeString(wiki)
		if err != nil {
			return err
		}
		log.Println("wiki: ", string(b64))
	}

	appView := pages.AppGo(v)
	if partial == "true" {
		return HTML(c, appView)
	}
	indexView := pages.AppIndex(v, appView)
	return HTML(c, indexView)
}

func GetAppIndexAclsHandler(c LemcContext) error {
	appUUID := c.Param("uuid")
	partial := strings.ToLower(c.QueryParam("partial"))
	accountID := c.UserContext().ActingAs.Account.ID

	app, err := models.AppByUUIDAndAccountID(appUUID, accountID)
	if err != nil {
		log.Printf("Error fetching app by UUID %s for account %d: %v", appUUID, accountID, err)
		if errors.Is(err, sql.ErrNoRows) {
			return c.String(http.StatusNotFound, "app not found")
		}
		return c.String(http.StatusInternalServerError, "Error retrieving app")
	}

	// Fetch specific user permissions for this app using the model function
	userID := c.UserContext().ActingAs.ID
	appPerms, err := models.AppPermissionsByUserAccountAndApp(userID, accountID, app.ID)
	if err != nil {
		// Handle potential database errors returned by the model function
		log.Printf("Error fetching app permissions via model for user %d, app %d: %v", userID, app.ID, err)
		return c.String(http.StatusInternalServerError, "Error retrieving app permissions")
	}
	// Assign the permissions (even if they are the default zero-value ones)
	app.UserPerms = appPerms

	cb, err := models.CookbookByIDAndAccountID(app.CookbookID, accountID)
	if err != nil {
		log.Printf("Error fetching cookbook by ID %d for account %d (from app %s): %v", app.CookbookID, accountID, appUUID, err)
		if errors.Is(err, sql.ErrNoRows) {
			return c.String(http.StatusNotFound, "Associated cookbook not found")
		}
		return c.String(http.StatusInternalServerError, "Error retrieving associated cookbook")
	}

	cb.UUID = app.UUID
	cb.Name = app.Name
	cb.Description = app.Description
	cb.YamlShared = app.YAMLShared
	cb.YamlIndividual = app.YAMLIndividual

	v := models.CoreView{
		App:      app,
		Cookbook: cb,
		BaseView: NewBaseView(c),
	}

	v.ViewType = "acls"
	acls, err := models.AppAclsUsers(accountID, app.ID) // Use app.ID here
	if err != nil {
		log.Printf("Error fetching app ACLs for app ID %d: %v", app.ID, err)
		return c.String(http.StatusInternalServerError, "Error fetching permissions")
	}
	v.AppAcls = acls                                  // Store in the new field
	v.App = app                                       // Assign the fetched app to the view model
	v.YamlDefault = models.YamlDefault{UUID: cb.UUID} // Ensure UUID is available

	v.YamlDefault.UUID = cb.UUID // Ensure UUID is set even if YAML unmarshal didn't happen

	appView := pages.AppAcls(v)
	if partial == "true" {
		return HTML(c, appView)
	}
	indexView := pages.AppIndex(v, appView)
	return HTML(c, indexView)
}

func GetAppIndexSharedHandler(c LemcContext) error {
	appUUID := c.Param("uuid")
	partial := strings.ToLower(c.QueryParam("partial"))
	accountID := c.UserContext().ActingAs.Account.ID

	app, err := models.AppByUUIDAndAccountID(appUUID, accountID)
	if err != nil {
		log.Printf("Error fetching app by UUID %s for account %d: %v", appUUID, accountID, err)
		if errors.Is(err, sql.ErrNoRows) {
			return c.String(http.StatusNotFound, "app not found")
		}
		return c.String(http.StatusInternalServerError, "Error retrieving app")
	}

	// Fetch specific user permissions for this app using the model function
	userID := c.UserContext().ActingAs.ID
	appPerms, err := models.AppPermissionsByUserAccountAndApp(userID, accountID, app.ID)
	if err != nil {
		// Handle potential database errors returned by the model function
		log.Printf("Error fetching app permissions via model for user %d, app %d: %v", userID, app.ID, err)
		return c.String(http.StatusInternalServerError, "Error retrieving app permissions")
	}
	// Assign the permissions (even if they are the default zero-value ones)
	app.UserPerms = appPerms

	cb, err := models.CookbookByIDAndAccountID(app.CookbookID, accountID)
	if err != nil {
		log.Printf("Error fetching cookbook by ID %d for account %d (from app %s): %v", app.CookbookID, accountID, appUUID, err)
		if errors.Is(err, sql.ErrNoRows) {
			return c.String(http.StatusNotFound, "Associated cookbook not found")
		}
		return c.String(http.StatusInternalServerError, "Error retrieving associated cookbook")
	}

	cb.UUID = app.UUID
	cb.Name = app.Name
	cb.Description = app.Description
	cb.YamlShared = app.YAMLShared
	cb.YamlIndividual = app.YAMLIndividual

	v := models.CoreView{
		App:      app,
		Cookbook: cb,
		BaseView: NewBaseView(c),
	}

	var isAdmin bool
	yaml.Unmarshal([]byte(cb.YamlShared), &v.YamlDefault)
	v.ViewType = "shared"
	isAdmin = true
	v.YamlDefault.UUID = cb.UUID // Ensure UUID is set even if YAML unmarshal didn't happen

	for _, page := range v.YamlDefault.Cookbook.Pages {
		jm := &util.JobMeta{
			UUID:     v.YamlDefault.UUID,
			PageID:   strconv.Itoa(page.PageID),
			UserID:   strconv.FormatInt(c.UserContext().ActingAs.ID, 10),
			Username: c.UserContext().ActingAs.Username,
		}

		cf, err := util.NewContainerFiles(jm, isAdmin)
		if err != nil {
			return err
		}

		err = cf.OpenFiles()
		if err != nil {
			return err
		}
		defer cf.CloseFiles()

		css, err := cf.Read(cf.Css)
		if err != nil {
			return err
		}
		page.CssCache = fmt.Sprintf("<style id='uuid-%s-pageid-%d-scope-%s-style'>%s</style>", cb.UUID, page.PageID, v.ViewType, css)

		html, err := cf.Read(cf.Html)
		if err != nil {
			return err
		}
		page.HtmlCache = fmt.Sprintf("<div id='uuid-%s-pageid-%d-scope-%s-html' class='page-inner'>%s</div>", cb.UUID, page.PageID, v.ViewType, html)

		js, err := cf.Read(cf.Js)
		if err != nil {
			return err
		}
		page.JsCache = fmt.Sprintf("<script id='uuid-%s-pageid-%d-scope-%s-script'>%s</script>", cb.UUID, page.PageID, v.ViewType, js)

		v.YamlDefault.Cookbook.Pages[page.PageID-1] = page
	}

	appView := pages.AppGo(v)
	if partial == "true" {
		return HTML(c, appView)
	}
	indexView := pages.AppIndex(v, appView)
	return HTML(c, indexView)
}

func PostAppCreate(c LemcContext) error {
	cb := &models.Cookbook{
		AccountID: c.UserContext().ActingAs.Account.ID,
		OwnerID:   c.UserContext().ActingAs.ID,
	}

	uuid := c.FormValue("cookbook-uuid")
	log.Println("uuid: ", uuid)

	err := cb.ByUUID(uuid)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Println("error creating app: ", err)
		c.AddErrorFlash("app-create", "sql ByUUID() error")
		return c.NoContent(http.StatusConflict)
	}

	if cb.IsDeleted || !cb.IsPublished {
		log.Printf("Attempted to create app with deleted or unpublished cookbook UUID: %s (Deleted: %t, Published: %t)", uuid, cb.IsDeleted, cb.IsPublished)
		c.AddErrorFlash("app-create", "Cannot create app from a deleted or unpublished cookbook.")
		return c.NoContent(http.StatusConflict)
	}

	log.Println("cb.YamlShared: ", cb.YamlShared)
	log.Println("cb.YamlIndividual: ", cb.YamlIndividual)

	app := &models.App{
		AccountID:      c.UserContext().ActingAs.Account.ID,
		OwnerID:        c.UserContext().ActingAs.ID,
		Name:           util.Sanitize(c.FormValue("name")),
		Description:    util.Sanitize(c.FormValue("description")),
		YAMLShared:     cb.YamlShared,
		YAMLIndividual: cb.YamlIndividual,
		CookbookID:     cb.ID,
	}

	tx, err := db.Db().Beginx()
	if err != nil {
		return err
	}

	err = app.Create(tx)
	if err != nil {
		log.Println("error creating app: ", err)
		errR := tx.Rollback()
		if errR != nil {
			log.Println("🔥 Failed to rollback transaction: ", errR)
		}
		c.AddErrorFlash("app-create", "server error creating app")
		return c.NoContent(http.StatusConflict)
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	userID := c.UserContext().ActingAs.ID
	accountID := c.UserContext().ActingAs.Account.ID
	cbs, err := models.Apps(userID, accountID, 1, 10)
	if err != nil {
		return err
	}

	totalapps, err := models.Countapps(userID, accountID)
	if err != nil {
		log.Printf("Error counting apps after creation: %v", err)
		return err
	}
	limit := 10 // Use the same hardcoded limit
	totalPages := 0
	if totalapps > 0 {
		totalPages = int(math.Ceil(float64(totalapps) / float64(limit)))
	}

	newSquid, newName, err := models.SquidAndNameByAccountID(c.UserContext().ActingAs.Account.ID)
	if err != nil {
		return c.String(http.StatusNotFound, "Can't create Squid")
	}

	v := models.AppsView{
		Apps:        cbs,
		CurrentPage: 1,          // Always page 1 after creation
		TotalPages:  totalPages, // Pass calculated total pages
		Limit:       limit,      // Pass the limit used
		BaseView:    NewBaseViewWithSquidAndAccountName(c, newSquid, newName),
	}

	c.AddSuccessFlash("apps-create", "new app created") // Fixed typo
	cv := pages.AppsList(v)

	pushedURL := fmt.Sprintf("%s?page=1&limit=%d", paths.Apps, limit) // Use the same limit
	c.Response().Header().Set("HX-Trigger", "closeNewappModal")
	c.Response().Header().Set("HX-Push-Url", pushedURL)

	return HTML(c, cv)
}

func AppRefreshHandler(c LemcContext) error {
	appUUID := c.Param("uuid")
	accountID := c.UserContext().ActingAs.Account.ID

	log.Printf("Attempting to refresh app %s for account %d", appUUID, accountID)

	app, err := models.AppByUUIDAndAccountID(appUUID, accountID)
	if err != nil {
		log.Printf("Error fetching app by UUID %s for account %d: %v", appUUID, accountID, err)
		if errors.Is(err, sql.ErrNoRows) {
			return c.String(http.StatusNotFound, "app not found")
		}
		return c.String(http.StatusInternalServerError, "Error retrieving app")
	}

	// Fetch specific user permissions for this app using the model function
	userID := c.UserContext().ActingAs.ID
	appPerms, err := models.AppPermissionsByUserAccountAndApp(userID, accountID, app.ID)
	if err != nil {
		// Handle potential database errors returned by the model function
		log.Printf("Error fetching app permissions via model for user %d, app %d: %v", userID, app.ID, err)
		return c.String(http.StatusInternalServerError, "Error retrieving app permissions")
	}
	// Assign the permissions (even if they are the default zero-value ones)
	app.UserPerms = appPerms

	cb, err := models.CookbookByIDAndAccountID(app.CookbookID, accountID)
	if err != nil {
		log.Printf("Error fetching cookbook by ID %d for account %d (from app %s): %v", app.CookbookID, accountID, appUUID, err)
		if errors.Is(err, sql.ErrNoRows) {
			return c.String(http.StatusNotFound, "Associated cookbook not found")
		}
		return c.String(http.StatusInternalServerError, "Error retrieving associated cookbook")
	}

	app.YAMLShared = cb.YamlShared
	app.YAMLIndividual = cb.YamlIndividual

	tx, err := db.Db().Beginx()
	if err != nil {
		log.Printf("Error starting transaction for app refresh %s: %v", appUUID, err)
		return c.String(http.StatusInternalServerError, "Error starting transaction")
	}

	err = app.Update(tx)
	if err != nil {
		log.Printf("Error updating app %s during refresh: %v", appUUID, err)
		errR := tx.Rollback()
		if errR != nil {
			log.Printf("🔥 Failed to rollback transaction for app refresh %s: %v", appUUID, errR)
		}
		return c.String(http.StatusInternalServerError, "Error updating app")
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("Error committing transaction for app refresh %s: %v", appUUID, err)
		return c.String(http.StatusInternalServerError, "Error committing transaction")
	}

	log.Printf("Successfully refreshed app %s for account %d from cookbook %d", appUUID, accountID, app.CookbookID)

	apps, err := models.Apps(userID, accountID, 1, DefaultappLimit)
	if err != nil {
		log.Printf("Error fetching apps after refresh: %v", err)
		return c.String(http.StatusInternalServerError, "Error fetching updated apps")
	}

	totalapps, err := models.Countapps(userID, accountID)
	if err != nil {
		log.Printf("Error counting apps: %v", err)
		return c.String(http.StatusInternalServerError, "Error counting apps")
	}

	totalPages := 0
	if totalapps > 0 {
		totalPages = int(math.Ceil(float64(totalapps) / float64(DefaultappLimit)))
	}

	newSquid, newName, err := models.SquidAndNameByAccountID(c.UserContext().ActingAs.Account.ID)
	if err != nil {
		return c.String(http.StatusNotFound, "Can't create Squid")
	}

	av := models.AppsView{
		Apps:        apps,
		CurrentPage: 1,
		TotalPages:  totalPages,
		Limit:       DefaultappLimit,
		BaseView:    NewBaseViewWithSquidAndAccountName(c, newSquid, newName),
	}

	c.AddSuccessFlash("app-refresh", "app refreshed successfully")
	cv := pages.Apps(av)
	return HTML(c, cv)
}
