package models

import (
	"time"
)

type JobInfo struct {
	ID          string    // Unique identifier (e.g., quartz JobKey string)
	RecipeName  string    // Name or ID of the associated recipe
	Username    string    // User associated with the job
	Type        string    // Type of job (NOW, IN, EVERY)
	Status      string    // Current status (Scheduled, Running, Completed, Failed, Unknown)
	CreatedAt   time.Time // When the job definition was created or first scheduled
	ScheduledAt time.Time // Next scheduled run time (for IN/EVERY)
}

type JobsView struct {
	BaseView
	Jobs        []JobInfo // The list of jobs for the current page
	CurrentPage int       // Current page number
	TotalPages  int       // Total number of pages
	Limit       int       // Number of items per page
}
