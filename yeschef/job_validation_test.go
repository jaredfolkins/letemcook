package yeschef

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jaredfolkins/letemcook/models"
	"github.com/reugn/go-quartz/quartz"
)

func TestJobValidation_NilJobDetail_ShouldNotPanic(t *testing.T) {
	// Setup a temporary queue for testing
	tmpDir := t.TempDir()
	queue := &jobQueue{
		Path: tmpDir,
		Name: IN_QUEUE,
	}

	// Create a malformed job file that will result in nil JobDetail
	malformedJobData := []byte(`{"job": null, "job_key": "test", "trigger": "test"}`)
	jobFile := filepath.Join(tmpDir, "1234567890.json")
	err := os.WriteFile(jobFile, malformedJobData, 0644)
	if err != nil {
		t.Fatalf("Failed to write test job file: %v", err)
	}

	// Attempt to pop - this should not panic
	job, err := queue.Pop()

	// We expect an error due to validation, not a panic
	if err == nil {
		t.Errorf("Expected error from Pop() with invalid job, got nil")
	}

	if job != nil {
		t.Errorf("Expected nil job from Pop() with invalid job, got %v", job)
	}
}

func TestJobValidation_ValidJob_ShouldPass(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	queue := &jobQueue{
		Path: tmpDir,
		Name: IN_QUEUE,
	}

	// Initialize XoxoX if not already done
	if XoxoX == nil {
		XoxoX = &ChefsKiss{
			RunningMan: NewRunningMan(),
		}
	}

	// Create a valid step job
	stepJob := &StepJob{
		Step: models.Step{
			Step:    1,
			Name:    "test-step",
			Image:   "test-image",
			Do:      "echo hello",
			Timeout: "30s",
		},
		RecipeJob: &JobRecipe{
			JobType:    "IN",
			UUID:       "test-uuid",
			CookbookID: "test-cookbook",
			UserID:     "123",
			Username:   "test-user",
			PageID:     "test-page",
			StepID:     "1",
			Scope:      "individual",
			Recipe: models.Recipe{
				Name:        "test-recipe",
				Description: "test description",
			},
		},
	}

	// Create job detail and scheduled job
	jobKey := quartz.NewJobKey("test-job")
	jobDetail := quartz.NewJobDetail(stepJob, jobKey)
	trigger := quartz.NewRunOnceTrigger(time.Minute)

	scheduledJob := &scheduledLemcJob{
		jobDetail:   jobDetail,
		trigger:     trigger,
		nextRunTime: time.Now().Add(time.Minute).UnixNano(),
	}

	// Push valid job
	err := queue.Push(scheduledJob)
	if err != nil {
		t.Fatalf("Failed to push valid job: %v", err)
	}

	// Pop should succeed without panic
	poppedJob, err := queue.Pop()
	if err != nil {
		t.Errorf("Expected no error from Pop() with valid job, got %v", err)
	}

	if poppedJob == nil {
		t.Errorf("Expected valid job from Pop(), got nil")
	}
}

func TestJobValidation_ExpiredJob_ShouldBeHandled(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	queue := &jobQueue{
		Path: tmpDir,
		Name: IN_QUEUE,
	}

	// Initialize XoxoX if not already done
	if XoxoX == nil {
		XoxoX = &ChefsKiss{
			RunningMan: NewRunningMan(),
		}
	}

	// Create job data that represents an expired IN job with proper trigger format
	expiredJobData := []byte(`{
		"job": {
			"Step": {
				"step": 1,
				"name": "expired-step",
				"image": "test-image",
				"do": "echo expired",
				"timeout": "30s"
			},
			"RecipeJob": {
				"JobType": "IN",
				"UUID": "expired-uuid",
				"CookbookID": "test-cookbook",
				"UserID": "123",
				"Username": "test-user",
				"PageID": "test-page",
				"StepID": "1",
				"Scope": "individual",
				"Recipe": {
					"recipe": "expired-recipe",
					"description": "expired test"
				}
			}
		},
		"job_key": "expired-job",
		"description": "expired job",
		"group": "test-group",
		"trigger": "RunOnceTrigger⇶1m0s⇶expired",
		"next_run_time": 1000000000
	}`)

	jobFile := filepath.Join(tmpDir, "1000000000.json")
	err := os.WriteFile(jobFile, expiredJobData, 0644)
	if err != nil {
		t.Fatalf("Failed to write expired job file: %v", err)
	}

	// Pop should handle expired job gracefully
	job, err := queue.Pop()

	// The behavior for expired jobs should be defined - either return error or handle gracefully
	// For now, we just ensure it doesn't panic
	t.Logf("Expired job handling: job=%v, err=%v", job != nil, err)
}
