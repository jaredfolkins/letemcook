package models

import "time"

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

type ImageInfo struct {
	Name           string
	LastUpdated    time.Time
	Exists         bool
	NewerAvailable bool
}

type SystemImagesView struct {
	BaseView
	Images []ImageInfo
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
