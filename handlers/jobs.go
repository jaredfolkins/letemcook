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
	"github.com/jaredfolkins/letemcook/paths"
	"github.com/jaredfolkins/letemcook/util"
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

	jobDataDir := util.QueuesPath()
	loadedJobs := []persistedJobInfo{}

	err := filepath.WalkDir(jobDataDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// Skip paths we can't stat/read but continue walking
			log.Printf("Error accessing path '%s': %v", path, err)
			return nil
		}

		if d.IsDir() {
			return nil
		}

		if filepath.Ext(d.Name()) != ".json" {
			return nil
		}

		fileData, readErr := os.ReadFile(path)
		if readErr != nil {
			log.Printf("Error reading job file '%s': %v", path, readErr)
			return nil
		}

		var jobData persistedJobInfo
		if unmarshalErr := json.Unmarshal(fileData, &jobData); unmarshalErr != nil {
			log.Printf("Error unmarshalling job file '%s': %v", path, unmarshalErr)
			return nil
		}

		loadedJobs = append(loadedJobs, jobData)
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Job data directory not found: %s. Returning empty job list.", jobDataDir)
			return []models.JobInfo{}, 0, nil
		}
		log.Printf("Error walking job directory '%s': %v", jobDataDir, err)
		return nil, 0, fmt.Errorf("failed to read job data directory: %w", err)
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
		return c.Redirect(http.StatusFound, paths.Login)
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
	baseView.ActiveNav = "account"
	baseView.ActiveSubNav = paths.AccountJobs

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

func getAllJobs(page, limit int) ([]models.JobInfo, int, error) {
	jobDataDir := util.QueuesPath()
	loadedJobs := []persistedJobInfo{}
	err := filepath.WalkDir(jobDataDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			log.Printf("Error accessing path '%s': %v", path, err)
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(d.Name()) != ".json" {
			return nil
		}
		fileData, rErr := os.ReadFile(path)
		if rErr != nil {
			log.Printf("Error reading job file '%s': %v", path, rErr)
			return nil
		}
		var jobData persistedJobInfo
		if uErr := json.Unmarshal(fileData, &jobData); uErr != nil {
			log.Printf("Error unmarshalling job file '%s': %v", path, uErr)
			return nil
		}
		loadedJobs = append(loadedJobs, jobData)
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return []models.JobInfo{}, 0, nil
		}
		return nil, 0, fmt.Errorf("failed to read job data directory: %w", err)
	}
	jobs := make([]models.JobInfo, len(loadedJobs))
	for i, jd := range loadedJobs {
		jobs[i] = models.JobInfo{
			ID:         jd.ID,
			RecipeName: jd.RecipeName,
			Username:   jd.Username,
			Type:       jd.JobType,
			Status:     jd.Status,
			CreatedAt:  jd.CreatedAt,
		}
		if jd.ScheduledAt != nil {
			jobs[i].ScheduledAt = *jd.ScheduledAt
		}
	}
	total := len(jobs)
	start := (page - 1) * limit
	end := start + limit
	if start >= total {
		return []models.JobInfo{}, total, nil
	}
	if end > total {
		end = total
	}
	return jobs[start:end], total, nil
}
