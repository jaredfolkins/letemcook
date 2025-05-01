package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/jaredfolkins/letemcook/models"
	"github.com/jaredfolkins/letemcook/views/pages"
	"github.com/labstack/echo/v4"
)

type persistedJobInfo struct {
	ID          string     `json:"id"`
	RecipeName  string     `json:"recipe_name"`
	Username    string     `json:"username"`
	AccountID   int64      `json:"account_id"` // Important for filtering later
	JobType     string     `json:"job_type"`   // e.g., "NOW", "IN", "EVERY"
	Status      string     `json:"status"`     // e.g., "Scheduled", "Running", "Completed"
	CreatedAt   time.Time  `json:"created_at"`
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"` // Pointer, might be null
}

func getJobs(page, limit int, c LemcContext) ([]models.JobInfo, int, error) { // Removed userID param, using context 'c' instead
	userCtx := c.UserContext()
	if userCtx == nil || userCtx.LoggedInAs == nil {
		return nil, 0, fmt.Errorf("user context not available")
	}
	expectedAccountID := userCtx.ActingAs.Account.ID

	jobDataDir := os.Getenv("LEMC_QUEUES")
	if jobDataDir == "" {
		jobDataDir = "data/queues" // Default path
		log.Println("LEMC_QUEUES environment variable not set, using default: data/queues")
	}
	loadedJobs := []persistedJobInfo{}

	files, err := os.ReadDir(jobDataDir)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Job data directory not found: %s. Returning empty job list.", jobDataDir)
			return []models.JobInfo{}, 0, nil
		} else {
			log.Printf("Error reading job data directory '%s': %v", jobDataDir, err)
			return nil, 0, fmt.Errorf("failed to read job data directory: %w", err)
		}
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue // Skip directories and non-json files
		}

		filePath := filepath.Join(jobDataDir, file.Name())
		fileData, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("Error reading job file '%s': %v", filePath, err)
			continue // Skip this file on read error
		}

		var jobData persistedJobInfo
		err = json.Unmarshal(fileData, &jobData)
		if err != nil {
			log.Printf("Error unmarshalling job file '%s': %v", filePath, err)
			continue // Skip this file on parse error
		}

		loadedJobs = append(loadedJobs, jobData)
	}

	filteredJobs := []models.JobInfo{}
	for _, jobData := range loadedJobs { // Iterate over persistedJobInfo
		shouldInclude := (jobData.AccountID == expectedAccountID)

		if shouldInclude {
			jobInfo := models.JobInfo{
				ID:         jobData.ID,
				RecipeName: jobData.RecipeName,
				Username:   jobData.Username,
				Type:       jobData.JobType,
				Status:     jobData.Status,
				CreatedAt:  jobData.CreatedAt,
			}
			if jobData.ScheduledAt != nil {
				jobInfo.ScheduledAt = *jobData.ScheduledAt
			}
			filteredJobs = append(filteredJobs, jobInfo) // Append the mapped struct
		}
	}

	totalJobs := len(filteredJobs)
	start := (page - 1) * limit
	end := start + limit
	if start >= totalJobs {
		return []models.JobInfo{}, totalJobs, nil // Page out of bounds
	}
	if end > totalJobs {
		end = totalJobs
	}

	return filteredJobs[start:end], totalJobs, nil
}

func GetJobs(c LemcContext) error { // Changed context type to LemcContext
	userCtx := c.UserContext()                       // Use context method
	if userCtx == nil || userCtx.LoggedInAs == nil { // Changed User to LoggedInAs
		return c.Redirect(http.StatusFound, "/lemc/login")
	}

	pageParam := c.QueryParam("page")
	limitParam := c.QueryParam("limit")
	page, err := strconv.Atoi(pageParam)
	if err != nil || page < 1 {
		page = 1 // Default to page 1
	}
	limit, err := strconv.Atoi(limitParam)
	if err != nil || limit < 1 {
		limit = 10 // Default limit
	}

	jobs, totalJobs, err := getJobs(page, limit, c) // Pass context 'c' to getJobs, removed userID
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve jobs: %v", err))
	}

	totalPages := (totalJobs + limit - 1) / limit // Calculate total pages

	baseView := NewBaseView(c) // Use the helper to create BaseView
	baseView.Title = "Jobs"

	viewData := models.JobsView{
		BaseView:    baseView, // Assign the prepared BaseView
		Jobs:        jobs,
		CurrentPage: page,
		TotalPages:  totalPages,
		Limit:       limit,
	}

	isPartial := c.QueryParam("partial") == "true" // Check for HTMX partial request

	jobsPage := pages.JobsPage(viewData) // Generate component once

	if isPartial {
		return HTML(c, jobsPage) // Use HTML helper
	} else {
		return HTML(c, pages.JobsIndex(viewData, jobsPage)) // Use HTML helper
	}
}
