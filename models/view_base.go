package models

type BaseView struct {
	Theme               string
	CacheBuster         string
	Title               string
	Username            string
	IsError             bool
	IsProtected         bool
	IsSetup             bool
	AccountSquid        string
	AccountName         string
	UserContext         *UserContext
	Env                 string
	RegistrationEnabled bool
	ShowAppsNav         bool
	ShowCookbooksNav    bool
	ShowAccountNav      bool
	ShowSystemNav       bool
	ActiveNav           string
}

type HomeView struct {
	BaseView BaseView
}

type AuthorView struct {
	BaseView BaseView
}

type PageCache struct {
	Html string
	Css  string
	Js   string
}

type CoreView struct {
	Cookbook                 *Cookbook
	CookbookAcls             []CookbookAcl
	CookbookAclSearchResults []CookbookAcl
	CookbookSearchResults    []Cookbook
	App                      *App
	AppAcls                  []AppAcl
	AppAclSearchResults      []AppAcl

	YamlDefault          YamlDefault
	YamlDefaultNoStorage YamlDefaultNoStorage
	ViewType             string // admin, user, acls
	BaseView             BaseView
}

type CookbooksView struct {
	Cookbooks   []Cookbook
	CurrentPage int // Added for pagination
	TotalPages  int // Added for pagination
	Limit       int // Added for pagination
	BaseView    BaseView
}

type AppsView struct {
	Apps        []App
	CurrentPage int // Added for pagination
	TotalPages  int // Added for pagination
	Limit       int // Added for pagination
	BaseView    BaseView
}

type JobView struct {
	BaseView BaseView
}

type LoginView struct {
	BaseView BaseView
}

type RegisterView struct {
	BaseView BaseView
}

type SetupView struct {
	BaseView BaseView
}
