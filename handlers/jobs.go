package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

// YesChef job structures - imported from yeschef package structures
type yeschefStepJob struct {
	Job         *yeschefStepJobData `json:"job"`
	JobKey      string              `json:"job_key"`
	Description string              `json:"description"`
	Group       string              `json:"group"`
	Trigger     string              `json:"trigger"`
	NextRunTime int64               `json:"next_run_time"`
}

type yeschefStepJobData struct {
	Step      *yeschefStep      `json:"Step"`
	RecipeJob *yeschefRecipeJob `json:"RecipeJob"`
}

type yeschefStep struct {
	Step    int    `json:"Step"`
	Name    string `json:"Name"`
	Image   string `json:"Image"`
	Do      string `json:"Do"`
	Timeout string `json:"Timeout"`
}

type yeschefRecipeJob struct {
	JobType    string         `json:"JobType"`
	UUID       string         `json:"UUID"`
	CookbookID string         `json:"CookbookID"`
	AppID      string         `json:"AppID"`
	UserID     string         `json:"UserID"`
	Username   string         `json:"Username"`
	PageID     string         `json:"PageID"`
	StepID     string         `json:"StepID"`
	Scope      string         `json:"Scope"`
	Recipe     *yeschefRecipe `json:"Recipe"`
}

type yeschefRecipe struct {
	Name        string `json:"Name"`
	Description string `json:"Description"`
}

type yeschefRecipeJobFile struct {
	Job         *yeschefRecipeJob `json:"job"`
	JobKey      string            `json:"job_key"`
	Description string            `json:"description"`
	Group       string            `json:"group"`
	Trigger     string            `json:"trigger"`
	NextRunTime int64             `json:"next_run_time"`
}

func extractJobInfoFromYeschef(fileData []byte, filename string, dirPath string) (*persistedJobInfo, error) {
	// Determine job type from directory path
	jobType := "UNKNOWN"
	if strings.Contains(dirPath, "/now/") || strings.HasSuffix(dirPath, "/now") {
		jobType = "NOW"
	} else if strings.Contains(dirPath, "/in/") || strings.HasSuffix(dirPath, "/in") {
		jobType = "IN"
	} else if strings.Contains(dirPath, "/every/") || strings.HasSuffix(dirPath, "/every") {
		jobType = "EVERY"
	}

	// Try to parse as step job first
	var stepJob yeschefStepJob
	if err := json.Unmarshal(fileData, &stepJob); err == nil && stepJob.Job != nil && stepJob.Job.RecipeJob != nil {
		userID, _ := strconv.ParseInt(stepJob.Job.RecipeJob.UserID, 10, 64)

		recipeName := "Unknown Recipe"
		if stepJob.Job.RecipeJob.Recipe != nil {
			recipeName = stepJob.Job.RecipeJob.Recipe.Name
		} else if stepJob.Job.Step != nil {
			recipeName = stepJob.Job.Step.Name
		}

		// Extract timestamp from filename for ID and creation time
		var timestamp int64
		var err error
		if strings.HasSuffix(filename, ".json") {
			timestamp, err = strconv.ParseInt(strings.TrimSuffix(filename, ".json"), 10, 64)
		} else {
			timestamp, err = strconv.ParseInt(filename, 10, 64)
		}
		if err != nil {
			timestamp = time.Now().UnixNano()
		}

		createdAt := time.Unix(0, timestamp)
		scheduledAt := time.Unix(0, stepJob.NextRunTime)

		return &persistedJobInfo{
			ID:          filename,
			RecipeName:  recipeName,
			Username:    stepJob.Job.RecipeJob.Username,
			AccountID:   userID,
			JobType:     jobType,
			Status:      "Scheduled",
			CreatedAt:   createdAt,
			ScheduledAt: &scheduledAt,
		}, nil
	}

	// Try to parse as recipe job
	var recipeJob yeschefRecipeJobFile
	if err := json.Unmarshal(fileData, &recipeJob); err == nil && recipeJob.Job != nil {
		userID, _ := strconv.ParseInt(recipeJob.Job.UserID, 10, 64)

		recipeName := "Unknown Recipe"
		if recipeJob.Job.Recipe != nil {
			recipeName = recipeJob.Job.Recipe.Name
		}

		// Extract timestamp from filename for ID and creation time
		var timestamp int64
		var err error
		if strings.HasSuffix(filename, ".json") {
			timestamp, err = strconv.ParseInt(strings.TrimSuffix(filename, ".json"), 10, 64)
		} else {
			timestamp, err = strconv.ParseInt(filename, 10, 64)
		}
		if err != nil {
			timestamp = time.Now().UnixNano()
		}

		createdAt := time.Unix(0, timestamp)
		scheduledAt := time.Unix(0, recipeJob.NextRunTime)

		return &persistedJobInfo{
			ID:          filename,
			RecipeName:  recipeName,
			Username:    recipeJob.Job.Username,
			AccountID:   userID,
			JobType:     jobType,
			Status:      "Scheduled",
			CreatedAt:   createdAt,
			ScheduledAt: &scheduledAt,
		}, nil
	}

	return nil, fmt.Errorf("unable to parse as yeschef job format")
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

		// Accept both .json files and files without extensions (for backward compatibility)
		filename := d.Name()
		if !strings.HasSuffix(filename, ".json") {
			// Check if it's a timestamp-like filename (all digits)
			if _, err := strconv.ParseInt(filename, 10, 64); err != nil {
				return nil // Skip non-numeric filenames without .json extension
			}
		}

		fileData, readErr := os.ReadFile(path)
		if readErr != nil {
			log.Printf("Error reading job file '%s': %v", path, readErr)
			return nil
		}

		// Try to parse as old persistedJobInfo format first
		var jobData persistedJobInfo
		if unmarshalErr := json.Unmarshal(fileData, &jobData); unmarshalErr == nil {
			// Check if the parsed data is actually valid (not just empty struct)
			if jobData.ID != "" && jobData.AccountID != 0 {
				loadedJobs = append(loadedJobs, jobData)
				return nil
			}
		}

		// Try to parse as yeschef format
		if jobInfo, err := extractJobInfoFromYeschef(fileData, filename, path); err == nil {
			loadedJobs = append(loadedJobs, *jobInfo)
		} else {
			log.Printf("Error parsing job file '%s' as either format: %v", path, err)
		}

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
