package paths

const (
	LEMCPrefix = "/lemc"
	LEMCDir    = "lemc"

	Setup     = "/setup"
	Login     = LEMCPrefix + "/login"
	Register  = LEMCPrefix + "/register"
	Logout    = LEMCPrefix + "/logout"
	Profile   = LEMCPrefix + "/profile"
	Apps      = LEMCPrefix + "/apps"
	Cookbooks = LEMCPrefix + "/cookbooks"

	AccountSettings = LEMCPrefix + "/account/settings"
	AccountUsers    = LEMCPrefix + "/account/users"
	AccountJobs     = LEMCPrefix + "/account/jobs"

	SystemSettings = LEMCPrefix + "/system/settings"
	SystemAccounts = LEMCPrefix + "/system/accounts"
	SystemImages   = LEMCPrefix + "/system/images"
	SystemJobs     = LEMCPrefix + "/system/jobs"

	Impersonate = LEMCPrefix + "/impersonate"

	LockerDownloadPattern = LEMCPrefix + "/locker/uuid/%s/page/%d/scope/%s/filename/"
	McpSharedJobPattern   = LEMCPrefix + "/app/job/shared/uuid/%s/page/%d/recipe/%s"
	SharedMount           = LEMCPrefix + "/shared"
)
