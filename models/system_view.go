package models

type SystemView struct {
	BaseView
}

type SystemAccountsView struct {
	BaseView
	Accounts    []Account
	CurrentPage int
	TotalPages  int
	Limit       int
}

type SystemImagesView struct {
	BaseView
	Images []string
}

type SystemJobsView struct {
	BaseView
	Jobs        []JobInfo
	CurrentPage int
	TotalPages  int
	Limit       int
}

type SystemSettingsView struct {
	BaseView
	Settings map[string]string
}
