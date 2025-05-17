package handlers

import (
	"encoding/gob"

	"github.com/jaredfolkins/letemcook/db"
	"github.com/jaredfolkins/letemcook/middleware"
	"github.com/jaredfolkins/letemcook/models"
	"github.com/labstack/echo/v4"
)

func Routes(e *echo.Echo) {
	gob.Register(&models.User{})
	gob.Register(&models.UserContext{})

	e.Use(echo.WrapMiddleware(middleware.AiDbDumpMiddleware(db.Db())))

	e.Use(middleware.Before)
	e.Use(middleware.ThemeMiddleware)
	e.GET("/setup", Ctx(SetupHandler))
	e.POST("/setup", Ctx(PostSetupHandler))
	e.GET("/", Ctx(redirLoginHandler))
	e.GET("/ws", Ctx(Ws))
	e.GET("/mcp/app/:uuid", Ctx(McpSSE))
	e.POST("/mcp/app/:uuid", Ctx(McpPost))
	e.GET("/navtop", Ctx(GetNavtop))
	e.GET("/heckle", Ctx(GetHeckle))

	root := e.Group("")
	root.Use(middleware.BeforeNav, middleware.RedirIfNotSetup, middleware.RedirIfAuthd)
	root.GET("/login", Ctx(redirLoginHandler))
	root.GET("/register", Ctx(redirRegisterHandler))

	external := e.Group("/lemc")
	external.Use(middleware.BeforeNav, middleware.RedirIfNotSetup)
	external.GET("/login", middleware.RedirIfAuthd(Ctx(LoginHandler)))
	external.POST("/login", Ctx(PostLoginHandler))
	external.GET("/register", middleware.RedirIfAuthd(Ctx(RegisterHandler)))
	external.POST("/register", Ctx(PostRegisterHandler))

	lemc := e.Group("/lemc")
	lemc.Use(middleware.BeforeNav, middleware.RedirIfNotSetup, middleware.RedirIfNotAuthd)
	lemc.GET("/cookbooks", middleware.ApplyMiddlewares(Ctx(GetCookbooksHandler), middleware.CheckPermission(models.CanAccessCookbooksView, models.CanAdministerAccount, models.CanAdministerSystem)))
	lemc.GET("/apps", middleware.ApplyMiddlewares(Ctx(GetAppsHandler), middleware.CheckPermission(models.CanAccessAppsView, models.CanAdministerAccount, models.CanAdministerSystem)))
	lemc.GET("/locker/uuid/:uuid/page/:page/scope/:scope/filename/:filename", middleware.ApplyMiddlewares(Ctx(LockerDownload)))
	lemc.GET("/wiki/image/:view_type/:uuid/:filename", middleware.ApplyMiddlewares(Ctx(GetCookbookWikiImage), middleware.CheckPermission(models.CanAccessAppsView, models.CanAdministerAccount)))
	lemc.GET("/profile", Ctx(GetProfileHandler))
	lemc.POST("/profile/password", Ctx(PostChangePasswordHandler))
	lemc.POST("/profile/settings/heckle", Ctx(PostToggleHeckleHandler))

	lemc.POST("/logout", Ctx(PostLogoutHandler))

	account := lemc.Group("/account")
	account.GET("/users", middleware.ApplyMiddlewares(Ctx(GetAllUsers), middleware.CheckPermission(models.CanAdministerAccount, models.CanAdministerSystem)))
	account.GET("/user/:id", middleware.ApplyMiddlewares(Ctx(GetUserHandler), middleware.CheckPermission(models.CanAdministerAccount)))
	account.POST("/user/create", middleware.ApplyMiddlewares(Ctx(CreateUserHandler), middleware.CheckPermission(models.CanAdministerAccount)))
	account.PUT("/user/:user_id/account/:account_id/permission/:permission_name", middleware.ApplyMiddlewares(Ctx(PutUserAccountPermissionToggleHandler), middleware.CheckPermission(models.CanAdministerAccount)))

	account.GET("/settings", middleware.ApplyMiddlewares(Ctx(GetAccountSettingsHandler), middleware.CheckPermission(models.CanAdministerAccount))) // Basic logged-in check is enough for now
	account.POST("/settings", middleware.ApplyMiddlewares(Ctx(PostAccountSettingsHandler), middleware.CheckPermission(models.CanAdministerAccount)))
	account.GET("/jobs", middleware.ApplyMiddlewares(Ctx(GetJobs), middleware.CheckPermission(models.CanAdministerAccount))) // TODO: i need more permissions here

	app := lemc.Group("/app")
	app.GET("/job/status/uuid/:uuid/page/:page/scope/:scope", middleware.ApplyMiddlewares(Ctx(GetAppJobStatus))) // TODO: i need more permissions here
	app.PUT("/job/:view_type/uuid/:uuid/page/:page/recipe/:recipe", middleware.ApplyMiddlewares(Ctx(PutAppJob))) // TODO: i need more permissions here

	app.GET("/index/individual/:uuid", middleware.ApplyMiddlewares(Ctx(GetAppIndexIndividualHandler), middleware.CheckPermission(models.CanIndividualApp, models.CanAdministerAccount)))
	app.GET("/index/shared/:uuid", middleware.ApplyMiddlewares(Ctx(GetAppIndexSharedHandler), middleware.CheckPermission(models.CanSharedApp, models.CanAdministerAccount)))
	app.GET("/index/acls/:uuid", middleware.ApplyMiddlewares(Ctx(GetAppIndexAclsHandler), middleware.CheckPermission(models.CanAclApp, models.CanAdministerAccount)))

	app.PATCH("/onregister/toggle/:uuid", middleware.ApplyMiddlewares(Ctx(PatchAppOnRegisterToggleHandler), middleware.CheckPermission(models.CanAclApp, models.CanAdministerAccount)))

	app.POST("/create", middleware.ApplyMiddlewares(Ctx(PostAppCreate), middleware.CheckPermission(models.CanCreateApp)))
	app.GET("/thumbnail/download/:uuid", middleware.ApplyMiddlewares(Ctx(GetAppThumbnail), middleware.CheckPermission(models.CanAccessAppsView)))
	app.POST("/refresh/:uuid", middleware.ApplyMiddlewares(Ctx(AppRefreshHandler), middleware.CheckPermission(models.CanEditApp, models.CanAdministerAccount)))
	app.GET("/acl/search/users/:uuid", middleware.ApplyMiddlewares(Ctx(GetAclAppSearchHandler), middleware.CheckPermission(models.CanEditApp, models.CanAdministerAccount)))
	app.POST("/acl/user/add/:uuid/:uid", middleware.ApplyMiddlewares(Ctx(PostAclUserToappHandler), middleware.CheckPermission(models.CanAclApp, models.CanAdministerAccount)))
	app.PUT("/acl/user/toggle/individual/:uuid/:uid", middleware.ApplyMiddlewares(Ctx(PutappAclToggleIndividualHandler), middleware.CheckPermission(models.CanAclApp, models.CanAdministerAccount)))
	app.PUT("/acl/user/toggle/shared/:uuid/:uid", middleware.ApplyMiddlewares(Ctx(PutappAclToggleSharedHandler), middleware.CheckPermission(models.CanAclApp, models.CanAdministerAccount)))
	app.PUT("/acl/user/toggle/admin/:uuid/:uid", middleware.ApplyMiddlewares(Ctx(PutappAclToggleAdminHandler), middleware.CheckPermission(models.CanAclApp, models.CanAdministerAccount)))
	app.DELETE("/acl/user/delete/:uuid/:uid", middleware.ApplyMiddlewares(Ctx(DeleteUserFromappHandler), middleware.CheckPermission(models.CanAclApp, models.CanAdministerAccount)))

	cookbook := lemc.Group("/cookbook")
	cookbook.GET("/job/status/uuid/:uuid/page/:page/scope/:scope", middleware.ApplyMiddlewares(Ctx(GetCookbookJobStatus))) // TODO: i need more permissions here
	cookbook.PUT("/job/:view_type/uuid/:uuid/page/:page/recipe/:recipe", middleware.ApplyMiddlewares(Ctx(PutCookbookJob))) // TODO: i need more permissions here
	cookbook.GET("/search", middleware.ApplyMiddlewares(Ctx(GetCookbookSearchByName), middleware.CheckPermission(models.CanCreateCookbook, models.CanAdministerAccount)))
	cookbook.POST("/create", middleware.ApplyMiddlewares(Ctx(PostCookbookCreate), middleware.CheckPermission(models.CanCreateCookbook, models.CanAdministerAccount)))

	cookbook.GET("/edit/acls/:uuid", middleware.ApplyMiddlewares(Ctx(GetCookbookEditAclsView), middleware.CheckPermission(models.CanAccessCookbooksView, models.CanAdministerAccount)))
	cookbook.GET("/edit/describe/:uuid", middleware.ApplyMiddlewares(Ctx(GetCookbookEditView), middleware.CheckPermission(models.CanAccessCookbooksView, models.CanAdministerAccount)))
	cookbook.GET("/edit/individual/:uuid", middleware.ApplyMiddlewares(Ctx(GetCookbookEditIndividualView), middleware.CheckPermission(models.CanAccessCookbooksView, models.CanAdministerAccount)))
	cookbook.GET("/edit/shared/:uuid", middleware.ApplyMiddlewares(Ctx(GetCookbookEditSharedView), middleware.CheckPermission(models.CanAccessCookbooksView, models.CanAdministerAccount)))
	cookbook.POST("/meta/update/:uuid", middleware.ApplyMiddlewares(Ctx(PostCookbookMetaUpdateHandler), middleware.CheckPermission(models.CanEditCookbook, models.CanAdministerAccount)))
	cookbook.GET("/acl/search/users/:uuid", middleware.ApplyMiddlewares(Ctx(GetAclCookbookSearchHandler), middleware.CheckPermission(models.CanEditCookbook, models.CanAdministerAccount)))
	cookbook.POST("/acl/user/add/:uuid/:uid", middleware.ApplyMiddlewares(Ctx(PostAclUserToCookbookHandler), middleware.CheckPermission(models.CanAclApp, models.CanAdministerAccount)))
	cookbook.DELETE("/acl/user/delete/:uuid/:uid", middleware.ApplyMiddlewares(Ctx(DeleteUserFromCookbookHandler), middleware.CheckPermission(models.CanAclApp, models.CanAdministerAccount)))
	cookbook.PUT("/acl/user/toggle/edit/:uuid/:uid", middleware.ApplyMiddlewares(Ctx(PutAclToogleEditCookbookHandler), middleware.CheckPermission(models.CanAclApp, models.CanAdministerAccount)))
	cookbook.GET("/yaml/download/:view_type/:uuid", middleware.ApplyMiddlewares(Ctx(GetYamlDownload), middleware.CheckPermission(models.CanAccessCookbooksView, models.CanAdministerAccount)))
	cookbook.POST("/yaml/upload/:view_type/:uuid", middleware.ApplyMiddlewares(Ctx(PostYamlUpload), middleware.CheckPermission(models.CanEditCookbook, models.CanAdministerAccount)))
	cookbook.GET("/config/:view_type/:action/:uuid", middleware.ApplyMiddlewares(Ctx(GetCookbookYamlConfigHandler), middleware.CheckPermission(models.CanAccessCookbooksView, models.CanAdministerAccount)))
	cookbook.POST("/config/:view_type/:action/:uuid", middleware.ApplyMiddlewares(Ctx(PostCookbookYamlConfigHandler), middleware.CheckPermission(models.CanEditCookbook, models.CanAdministerAccount)))
	cookbook.POST("/wiki/create/:uuid", middleware.ApplyMiddlewares(Ctx(PostCookbookWikiSaveHandler), middleware.CheckPermission(models.CanEditCookbook, models.CanAdministerAccount)))
	cookbook.POST("/wiki/image/:view_type/:uuid", middleware.ApplyMiddlewares(PostCookbookWikiImage, middleware.CheckPermission(models.CanEditCookbook, models.CanAdministerAccount)))
	cookbook.GET("/thumbnail/download/:uuid", middleware.ApplyMiddlewares(Ctx(GetCookbookThumbnailImage), middleware.CheckPermission(models.CanAccessCookbooksView, models.CanAdministerAccount)))
	cookbook.POST("/thumbnail/upload/:uuid", middleware.ApplyMiddlewares(PostCookbookThumbnailImage, middleware.CheckPermission(models.CanEditCookbook, models.CanAdministerAccount)))
	cookbook.PATCH("/toggle/:toggle_type/:uuid", middleware.ApplyMiddlewares(Ctx(PatchCookbookToggle), middleware.CheckPermission(models.CanEditCookbook, models.CanAdministerAccount)))

	system := lemc.Group("/system")
	system.GET("/settings", middleware.ApplyMiddlewares(Ctx(GetSystemSettingsHandler), middleware.CheckPermission(models.CanAdministerSystem)))
	system.GET("/accounts", middleware.ApplyMiddlewares(Ctx(GetSystemAccountsHandler), middleware.CheckPermission(models.CanAdministerSystem)))
	system.GET("/images", middleware.ApplyMiddlewares(Ctx(GetSystemImagesHandler), middleware.CheckPermission(models.CanAdministerSystem)))
	system.GET("/jobs", middleware.ApplyMiddlewares(Ctx(GetSystemJobsHandler), middleware.CheckPermission(models.CanAdministerSystem)))

	e.Use(middleware.After)
}
