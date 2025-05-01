package models

type AccountSettingsView struct {
	BaseView
	Settings        *AccountSettings
	AvailableThemes []string
}
