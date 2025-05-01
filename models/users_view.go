package models

type UsersView struct {
	BaseView           // Embed BaseView for common layout data
	Users       []User // List of users to display
	CurrentPage int    // Current page number for pagination
	TotalPages  int    // Total number of pages
	Limit       int    // Number of users per page
}
