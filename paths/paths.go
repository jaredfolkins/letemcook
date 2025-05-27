package paths

const (
	LEMCPrefix = "/lemc"
	LEMCDir    = "lemc"

	// Basic paths
	Setup     = "/setup"
	Login     = "/lemc/login"
	Register  = "/lemc/register"
	Logout    = "/lemc/logout"
	Profile   = "/lemc/profile"
	Apps      = "/lemc/apps"
	Cookbooks = "/lemc/cookbooks"
	Home      = "/"
	NavTop    = "/navtop"

	// Account paths
	AccountSettings   = "/lemc/account/settings"
	AccountUsers      = "/lemc/account/users"
	AccountJobs       = "/lemc/account/jobs"
	AccountUserCreate = "/lemc/account/user/create"

	// System paths
	SystemSettings        = "/lemc/system/settings"
	SystemAccounts        = "/lemc/system/accounts"
	SystemImages          = "/lemc/system/images"
	SystemJobs            = "/lemc/system/jobs"
	SystemSettingsPartial = "/lemc/system/settings?partial=true"
	SystemImagesPull      = "/lemc/system/images/pull"

	// App paths
	AppCreate         = "/lemc/app/create"
	Impersonate       = "/lemc/impersonate"
	ImpersonateSearch = "/lemc/impersonate/search"

	// Cookbook paths
	CookbookCreate = "/lemc/cookbook/create"
	CookbookSearch = "/lemc/cookbook/search"

	// Profile paths
	ProfilePassword       = "/lemc/profile/password"
	ProfileSettingsHeckle = "/lemc/profile/settings/heckle"

	// Heckle paths
	Heckle       = "/heckle"
	HecklePublic = "/heckle/public"

	// Template patterns using fmt.Sprintf
	ImpersonatePattern = "/lemc/impersonate/%d/%d"
	NavTopPattern      = "/navtop?squid=%s&account=%s"
	LoginPattern       = "/lemc/login?squid=%s&account=%s"
	RegisterPattern    = "/lemc/register?squid=%s&account=%s"

	// Account template patterns
	AccountUserPattern                      = "/lemc/account/user/%d"
	AccountUsersPagePattern                 = "/lemc/account/users?page=%d&limit=%d"
	AccountUsersPagePartialPattern          = "/lemc/account/users?page=%d&limit=%d&partial=true"
	AccountJobsPagePattern                  = "/lemc/account/jobs?page=%d&limit=%d"
	AccountJobsPagePartialPattern           = "/lemc/account/jobs?page=%d&limit=%d&partial=true"
	AccountUserPermissionCanCreateApps      = "/lemc/account/user/%d/account/%d/permission/can_create_apps"
	AccountUserPermissionCanViewApps        = "/lemc/account/user/%d/account/%d/permission/can_view_apps"
	AccountUserPermissionCanCreateCookbooks = "/lemc/account/user/%d/account/%d/permission/can_create_cookbooks"
	AccountUserPermissionCanViewCookbooks   = "/lemc/account/user/%d/account/%d/permission/can_view_cookbooks"
	AccountUserPermissionCanAdminister      = "/lemc/account/user/%d/account/%d/permission/can_administer"
	AccountUserPermissionIsOwner            = "/lemc/account/user/%d/account/%d/permission/is_owner"

	// System template patterns
	SystemAccountsPagePattern        = "/lemc/system/accounts?page=%d&limit=%d"
	SystemAccountsPagePartialPattern = "/lemc/system/accounts?page=%d&limit=%d&partial=true"
	SystemJobsPagePattern            = "/lemc/system/jobs?page=%d&limit=%d"
	SystemJobsPagePartialPattern     = "/lemc/system/jobs?page=%d&limit=%d&partial=true"

	// App template patterns
	AppThumbnailDownloadPattern       = "/lemc/app/thumbnail/download/%s"
	AppIndexPattern                   = "/lemc/app/index/%s/%s"
	AppIndexPartialPattern            = "/lemc/app/index/%s/%s?partial=true"
	AppIndexIndividualPattern         = "/lemc/app/index/i/individual/%s"
	AppIndexIndividualPartialPattern  = "/lemc/app/index/i/individual/%s?partial=true"
	AppIndexSharedPattern             = "/lemc/app/index/s/shared/%s"
	AppIndexSharedPartialPattern      = "/lemc/app/index/s/shared/%s?partial=true"
	AppIndexAclsPattern               = "/lemc/app/index/a/acls/%s"
	AppIndexAclsPartialPattern        = "/lemc/app/index/a/acls/%s?partial=true"
	AppRefreshPattern                 = "/lemc/app/refresh/%s"
	AppsPagePattern                   = "/lemc/apps?page=%d&limit=%d"
	AppsPagePartialPattern            = "/lemc/apps?page=%d&limit=%d&partial=true"
	AppOnRegisterTogglePattern        = "/lemc/app/onregister/toggle/%s"
	AppAclSearchUsersPattern          = "/lemc/app/acl/search/users/%s"
	AppAclUserAddPattern              = "/lemc/app/acl/user/add/%s/%d"
	AppAclUserToggleIndividualPattern = "/lemc/app/acl/user/toggle/individual/%s/%d"
	AppAclUserToggleSharedPattern     = "/lemc/app/acl/user/toggle/shared/%s/%d"
	AppAclUserToggleAdminPattern      = "/lemc/app/acl/user/toggle/admin/%s/%d"
	AppAclUserDeletePattern           = "/lemc/app/acl/user/delete/%s/%d"
	AppJobStatusPattern               = "/lemc/app/job/status/uuid/%s/page/%d/scope/%s"
	AppJobPattern                     = "/lemc/app/job/%s/uuid/%s/page/%d/recipe/%s"

	// Cookbook template patterns
	CookbookThumbnailDownloadPattern = "/lemc/cookbook/thumbnail/download/%s?ts=%s"
	CookbookEditPattern              = "/lemc/cookbook/edit/%s/%s"
	CookbookEditPartialPattern       = "/lemc/cookbook/edit/%s/%s?partial=true"
	CookbookEditIndividualPattern    = "/lemc/cookbook/edit/individual/%s"
	CookbooksPagePattern             = "/lemc/cookbooks?page=%d&limit=%d"
	CookbooksPagePartialPattern      = "/lemc/cookbooks?page=%d&limit=%d&partial=true"
	CookbookConfigPattern            = "/lemc/cookbook/config/%s/%s/%s"
	CookbookConfigAllPattern         = "/lemc/cookbook/config/%s/all/%s"
	CookbookTogglePublishedPattern   = "/lemc/cookbook/toggle/published/%s"
	CookbookToggleDeletedPattern     = "/lemc/cookbook/toggle/deleted/%s"
	CookbookYamlDownloadPattern      = "/lemc/cookbook/yaml/download/%s/%s"
	CookbookJobStatusPattern         = "/lemc/cookbook/job/status/uuid/%s/page/%d/scope/%s"
	CookbookJobPattern               = "/lemc/cookbook/job/%s/uuid/%s/page/%d/recipe/%s"
	CookbookMetaUpdatePattern        = "/lemc/cookbook/meta/update/%s"
	CookbookThumbnailUploadPattern   = "/lemc/cookbook/thumbnail/upload/%s"
	CookbookYamlUploadPattern        = "/lemc/cookbook/yaml/upload/%s/%s"
	CookbookWikiImagePattern         = "/lemc/cookbook/wiki/image/%s/%s"
	CookbookWikiCreatePattern        = "/lemc/cookbook/wiki/create/%s"
	CookbookAclSearchUsersPattern    = "/lemc/cookbook/acl/search/users/%s"
	CookbookAclUserAddPattern        = "/lemc/cookbook/acl/user/add/%s/%d"
	CookbookAclUserToggleEditPattern = "/lemc/cookbook/acl/user/toggle/edit/%s/%d"
	CookbookAclUserDeletePattern     = "/lemc/cookbook/acl/user/delete/%s/%d"

	// Generic patterns
	GenericJobPattern = "/lemc/%s/job/%s/uuid/%s/page/%d/recipe/%s"

	// Theme and static asset patterns
	ThemeCssPattern  = "/themes/%s/public/css/compiled.css?v=%s"
	ThemeIconPattern = "/themes/%s/public/imgs/%sx%s.ico?v=%s"
	WebSocketPattern = "/ws"

	// Existing patterns
	LockerDownloadPattern = "/lemc/locker/uuid/%s/page/%d/scope/%s/filename/"
	McpSharedJobPattern   = "/lemc/app/job/shared/uuid/%s/page/%d/recipe/%s"
	SharedMount           = "/lemc/shared"
)
