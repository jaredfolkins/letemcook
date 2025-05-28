package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jaredfolkins/letemcook/util"

	"github.com/jaredfolkins/letemcook/middleware"
	"github.com/jaredfolkins/letemcook/models"
	"github.com/jaredfolkins/letemcook/tests/testutil"
	"github.com/labstack/echo/v4"
)

// Helper functions

// testCleanup removes all test data to ensure test isolation
func testCleanup(t *testing.T) {
	t.Helper()
	queuesPath := util.QueuesPath()
	if err := os.RemoveAll(queuesPath); err != nil {
		t.Logf("Warning: failed to clean up test directory: %v", err)
	}
}

// setupTestEnv sets up the test environment and returns a cleanup function
func setupTestEnv(t *testing.T) func() {
	t.Helper()
	os.Setenv("LEMC_ENV", "test")
	os.Setenv("LEMC_DATA", testutil.DataRoot())
	testCleanup(t) // Clean up any existing data
	return func() { testCleanup(t) }
}

// testContextWithAccount creates a test context with a user belonging to given account ID
func testContextWithAccount(t *testing.T, accountID int64) LemcContext {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	user := &models.User{Account: &models.Account{ID: accountID}}
	uc := &models.UserContext{ActingAs: user, LoggedInAs: user}
	return middleware.SetUserContext(c, uc)
}

// writeJob writes a legacy format job to a file
func writeJob(t *testing.T, dir, name string, info persistedJobInfo) {
	t.Helper()
	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("failed to marshal job: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
		t.Fatalf("failed to write job file: %v", err)
	}
}

// writeYeschefStepJob writes a yeschef format step job to a file
func writeYeschefStepJob(t *testing.T, dir, filename string, jobType, recipeName, username string, userID int64, nextRunTime int64) {
	t.Helper()

	userIDStr := fmt.Sprintf("%d", userID)

	yeschefJob := map[string]interface{}{
		"job": map[string]interface{}{
			"Step": map[string]interface{}{
				"Step":    1,
				"Name":    "test step",
				"Image":   "test-image:latest",
				"Do":      getStepDo(jobType),
				"Timeout": "10.minutes",
			},
			"RecipeJob": map[string]interface{}{
				"JobType":    "cookbook",
				"UUID":       "test-uuid-123",
				"CookbookID": "1",
				"AppID":      "",
				"UserID":     userIDStr,
				"Username":   username,
				"PageID":     "1",
				"StepID":     "",
				"Scope":      "individual",
				"Recipe": map[string]interface{}{
					"Name":        recipeName,
					"Description": "A test recipe for " + jobType + " job type",
				},
			},
		},
		"job_key":       "[cookbook][individual][userid:" + userIDStr + "][page:1][uuid:test-uuid-123][queue:" + strings.ToLower(jobType) + "]",
		"description":   "StepJob: " + getStepDo(jobType),
		"group":         "[userid:" + userIDStr + "][page:1][uuid:test-uuid-123][group]",
		"trigger":       getTrigger(jobType),
		"next_run_time": nextRunTime,
	}

	jobData, err := json.Marshal(yeschefJob)
	if err != nil {
		t.Fatalf("failed to marshal yeschef job: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, filename), jobData, 0o644); err != nil {
		t.Fatalf("failed to write yeschef job file: %v", err)
	}
}

// writeYeschefRecipeJob writes a yeschef format recipe job to a file (used for NOW jobs)
func writeYeschefRecipeJob(t *testing.T, dir, filename string, recipeName, username string, userID int64, nextRunTime int64) {
	t.Helper()

	userIDStr := fmt.Sprintf("%d", userID)

	yeschefJob := map[string]interface{}{
		"job": map[string]interface{}{
			"JobType":    "cookbook",
			"UUID":       "test-uuid-456",
			"CookbookID": "1",
			"AppID":      "",
			"UserID":     userIDStr,
			"Username":   username,
			"PageID":     "1",
			"StepID":     "",
			"Scope":      "individual",
			"Recipe": map[string]interface{}{
				"Name":        recipeName,
				"Description": "A test recipe for NOW job type",
				"Steps": []map[string]interface{}{
					{
						"Step":    1,
						"Name":    "test step",
						"Image":   "test-image:latest",
						"Do":      "now",
						"Timeout": "10.minutes",
					},
				},
			},
		},
		"job_key":       "[cookbook][individual][userid:" + userIDStr + "][page:1][uuid:test-uuid-456][queue:now]",
		"description":   "[cookbook][individual][userid:" + userIDStr + "][page:1][uuid:test-uuid-456][queue:cookbook]",
		"group":         "[userid:" + userIDStr + "][page:1][uuid:test-uuid-456][group]",
		"trigger":       "RunOnceTrigger::100ms::expired",
		"next_run_time": nextRunTime,
	}

	jobData, err := json.Marshal(yeschefJob)
	if err != nil {
		t.Fatalf("failed to marshal yeschef recipe job: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, filename), jobData, 0o644); err != nil {
		t.Fatalf("failed to write yeschef recipe job file: %v", err)
	}
}

// Helper functions for job type specific data
func getStepDo(jobType string) string {
	switch jobType {
	case "NOW":
		return "now"
	case "IN":
		return "in.5.minutes"
	case "EVERY":
		return "every.30.seconds"
	default:
		return "now"
	}
}

func getTrigger(jobType string) string {
	switch jobType {
	case "NOW":
		return "RunOnceTrigger::100ms::expired"
	case "IN":
		return "RunOnceTrigger::5m0s::expired"
	case "EVERY":
		return "SimpleTrigger::30s"
	default:
		return "RunOnceTrigger::100ms::expired"
	}
}

// Test for legacy format jobs (should still work)
func TestGetJobsLegacyFormat(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	queuesPath := util.QueuesPath()
	nowDir := filepath.Join(queuesPath, "now")
	inDir := filepath.Join(queuesPath, "in")
	everyDir := filepath.Join(queuesPath, "every")

	os.MkdirAll(nowDir, 0o755)
	os.MkdirAll(inDir, 0o755)
	os.MkdirAll(everyDir, 0o755)

	// Create legacy format jobs
	j1 := persistedJobInfo{ID: "legacy-now", RecipeName: "legacy NOW recipe", Username: "user1", AccountID: 1, JobType: "NOW", Status: "Running", CreatedAt: time.Now()}
	j2 := persistedJobInfo{ID: "legacy-in", RecipeName: "legacy IN recipe", Username: "user1", AccountID: 1, JobType: "IN", Status: "Scheduled", CreatedAt: time.Now()}
	j3 := persistedJobInfo{ID: "legacy-every", RecipeName: "legacy EVERY recipe", Username: "user2", AccountID: 2, JobType: "EVERY", Status: "Scheduled", CreatedAt: time.Now()}

	writeJob(t, nowDir, "j1.json", j1)
	writeJob(t, inDir, "j2.json", j2)
	writeJob(t, everyDir, "j3.json", j3)

	ctx := testContextWithAccount(t, 1)
	jobs, total, err := getJobs(1, 10, ctx)
	if err != nil {
		t.Fatalf("getJobs returned error: %v", err)
	}

	if total != 2 || len(jobs) != 2 {
		t.Fatalf("expected 2 jobs for account 1, got total=%d len=%d", total, len(jobs))
	}

	// Verify job types are correct
	jobTypes := make(map[string]bool)
	for _, job := range jobs {
		jobTypes[job.Type] = true
	}

	if !jobTypes["NOW"] || !jobTypes["IN"] {
		t.Error("expected to find both NOW and IN job types for account 1")
	}
}

// Test for yeschef format NOW jobs
func TestGetJobsYeschefNOWFormat(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	queuesPath := util.QueuesPath()
	nowDir := filepath.Join(queuesPath, "now")
	os.MkdirAll(nowDir, 0o755)

	timestamp := time.Now().UnixNano()
	writeYeschefRecipeJob(t, nowDir, "1748466187227539001.json", "yeschef NOW recipe", "test-user", 1, timestamp)

	ctx := testContextWithAccount(t, 1)
	jobs, total, err := getJobs(1, 10, ctx)
	if err != nil {
		t.Fatalf("getJobs returned error: %v", err)
	}

	if total != 1 || len(jobs) != 1 {
		t.Fatalf("expected 1 job for account 1, got total=%d len=%d", total, len(jobs))
	}

	job := jobs[0]
	if job.RecipeName != "yeschef NOW recipe" {
		t.Errorf("expected recipe name 'yeschef NOW recipe', got '%s'", job.RecipeName)
	}
	if job.Type != "NOW" {
		t.Errorf("expected job type 'NOW', got '%s'", job.Type)
	}
	if job.Username != "test-user" {
		t.Errorf("expected username 'test-user', got '%s'", job.Username)
	}
	if job.Status != "Scheduled" {
		t.Errorf("expected status 'Scheduled', got '%s'", job.Status)
	}

	t.Logf("✅ Successfully parsed yeschef format NOW job: %s (Type: %s, Status: %s)",
		job.RecipeName, job.Type, job.Status)
}

// Test for yeschef format IN jobs
func TestGetJobsYeschefINFormat(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	queuesPath := util.QueuesPath()
	inDir := filepath.Join(queuesPath, "in")
	os.MkdirAll(inDir, 0o755)

	timestamp := time.Now().UnixNano() + 5*60*1000000000 // 5 minutes from now
	writeYeschefStepJob(t, inDir, "1748466187227539002.json", "IN", "yeschef IN recipe", "test-user", 1, timestamp)

	ctx := testContextWithAccount(t, 1)
	jobs, total, err := getJobs(1, 10, ctx)
	if err != nil {
		t.Fatalf("getJobs returned error: %v", err)
	}

	if total != 1 || len(jobs) != 1 {
		t.Fatalf("expected 1 job for account 1, got total=%d len=%d", total, len(jobs))
	}

	job := jobs[0]
	if job.RecipeName != "yeschef IN recipe" {
		t.Errorf("expected recipe name 'yeschef IN recipe', got '%s'", job.RecipeName)
	}
	if job.Type != "IN" {
		t.Errorf("expected job type 'IN', got '%s'", job.Type)
	}
	if job.Username != "test-user" {
		t.Errorf("expected username 'test-user', got '%s'", job.Username)
	}
	if job.Status != "Scheduled" {
		t.Errorf("expected status 'Scheduled', got '%s'", job.Status)
	}

	t.Logf("✅ Successfully parsed yeschef format IN job: %s (Type: %s, Status: %s)",
		job.RecipeName, job.Type, job.Status)
}

// Test for yeschef format EVERY jobs
func TestGetJobsYeschefEVERYFormat(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	queuesPath := util.QueuesPath()
	everyDir := filepath.Join(queuesPath, "every")
	os.MkdirAll(everyDir, 0o755)

	timestamp := time.Now().UnixNano() + 30*1000000000 // 30 seconds from now
	writeYeschefStepJob(t, everyDir, "1748466187227539003.json", "EVERY", "yeschef EVERY recipe", "test-user", 1, timestamp)

	ctx := testContextWithAccount(t, 1)
	jobs, total, err := getJobs(1, 10, ctx)
	if err != nil {
		t.Fatalf("getJobs returned error: %v", err)
	}

	if total != 1 || len(jobs) != 1 {
		t.Fatalf("expected 1 job for account 1, got total=%d len=%d", total, len(jobs))
	}

	job := jobs[0]
	if job.RecipeName != "yeschef EVERY recipe" {
		t.Errorf("expected recipe name 'yeschef EVERY recipe', got '%s'", job.RecipeName)
	}
	if job.Type != "EVERY" {
		t.Errorf("expected job type 'EVERY', got '%s'", job.Type)
	}
	if job.Username != "test-user" {
		t.Errorf("expected username 'test-user', got '%s'", job.Username)
	}
	if job.Status != "Scheduled" {
		t.Errorf("expected status 'Scheduled', got '%s'", job.Status)
	}

	t.Logf("✅ Successfully parsed yeschef format EVERY job: %s (Type: %s, Status: %s)",
		job.RecipeName, job.Type, job.Status)
}

// Test mixing legacy and yeschef formats
func TestGetJobsMixedFormats(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	queuesPath := util.QueuesPath()
	nowDir := filepath.Join(queuesPath, "now")
	inDir := filepath.Join(queuesPath, "in")
	everyDir := filepath.Join(queuesPath, "every")

	os.MkdirAll(nowDir, 0o755)
	os.MkdirAll(inDir, 0o755)
	os.MkdirAll(everyDir, 0o755)

	// Create one legacy job and one yeschef job for each type
	legacyNow := persistedJobInfo{ID: "legacy-now", RecipeName: "legacy NOW", Username: "user1", AccountID: 1, JobType: "NOW", Status: "Running", CreatedAt: time.Now()}
	writeJob(t, nowDir, "legacy-now.json", legacyNow)

	timestamp := time.Now().UnixNano()
	writeYeschefRecipeJob(t, nowDir, "yeschef-now.json", "yeschef NOW", "user1", 1, timestamp)
	writeYeschefStepJob(t, inDir, "yeschef-in.json", "IN", "yeschef IN", "user1", 1, timestamp+5*60*1000000000)
	writeYeschefStepJob(t, everyDir, "yeschef-every.json", "EVERY", "yeschef EVERY", "user1", 1, timestamp+30*1000000000)

	ctx := testContextWithAccount(t, 1)
	jobs, total, err := getJobs(1, 10, ctx)
	if err != nil {
		t.Fatalf("getJobs returned error: %v", err)
	}

	if total != 4 || len(jobs) != 4 {
		t.Fatalf("expected 4 jobs for account 1, got total=%d len=%d", total, len(jobs))
	}

	// Verify we have all job types
	jobTypes := make(map[string]int)
	for _, job := range jobs {
		jobTypes[job.Type]++
	}

	if jobTypes["NOW"] != 2 {
		t.Errorf("expected 2 NOW jobs, got %d", jobTypes["NOW"])
	}
	if jobTypes["IN"] != 1 {
		t.Errorf("expected 1 IN job, got %d", jobTypes["IN"])
	}
	if jobTypes["EVERY"] != 1 {
		t.Errorf("expected 1 EVERY job, got %d", jobTypes["EVERY"])
	}

	t.Logf("✅ Successfully parsed mixed format jobs: NOW=%d, IN=%d, EVERY=%d",
		jobTypes["NOW"], jobTypes["IN"], jobTypes["EVERY"])
}

// Test account filtering across different job types
func TestGetJobsAccountFiltering(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	queuesPath := util.QueuesPath()
	nowDir := filepath.Join(queuesPath, "now")
	inDir := filepath.Join(queuesPath, "in")
	everyDir := filepath.Join(queuesPath, "every")

	os.MkdirAll(nowDir, 0o755)
	os.MkdirAll(inDir, 0o755)
	os.MkdirAll(everyDir, 0o755)

	timestamp := time.Now().UnixNano()

	// Create jobs for different accounts
	writeYeschefRecipeJob(t, nowDir, "account1-now.json", "Account 1 NOW", "user1", 1, timestamp)
	writeYeschefStepJob(t, inDir, "account1-in.json", "IN", "Account 1 IN", "user1", 1, timestamp+5*60*1000000000)
	writeYeschefStepJob(t, everyDir, "account1-every.json", "EVERY", "Account 1 EVERY", "user1", 1, timestamp+30*1000000000)

	writeYeschefRecipeJob(t, nowDir, "account2-now.json", "Account 2 NOW", "user2", 2, timestamp)
	writeYeschefStepJob(t, inDir, "account2-in.json", "IN", "Account 2 IN", "user2", 2, timestamp+5*60*1000000000)

	// Test account 1
	ctx1 := testContextWithAccount(t, 1)
	jobs1, total1, err := getJobs(1, 10, ctx1)
	if err != nil {
		t.Fatalf("getJobs returned error for account 1: %v", err)
	}

	if total1 != 3 || len(jobs1) != 3 {
		t.Fatalf("expected 3 jobs for account 1, got total=%d len=%d", total1, len(jobs1))
	}

	// Test account 2
	ctx2 := testContextWithAccount(t, 2)
	jobs2, total2, err := getJobs(1, 10, ctx2)
	if err != nil {
		t.Fatalf("getJobs returned error for account 2: %v", err)
	}

	if total2 != 2 || len(jobs2) != 2 {
		t.Fatalf("expected 2 jobs for account 2, got total=%d len=%d", total2, len(jobs2))
	}

	t.Logf("✅ Account filtering works correctly: Account 1=%d jobs, Account 2=%d jobs", total1, total2)
}

// Test recursive directory scanning
func TestGetJobsRecursiveScanning(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	queuesPath := util.QueuesPath()
	nowDir := filepath.Join(queuesPath, "now")
	nestedDir := filepath.Join(nowDir, "nested", "deep")
	os.MkdirAll(nestedDir, 0o755)

	// Create jobs at different directory levels
	j1 := persistedJobInfo{ID: "root-job", RecipeName: "root job", Username: "user1", AccountID: 1, JobType: "NOW", Status: "Running", CreatedAt: time.Now()}
	writeJob(t, nowDir, "root.json", j1)

	timestamp := time.Now().UnixNano()
	writeYeschefRecipeJob(t, nestedDir, "nested.json", "nested job", "user1", 1, timestamp)

	ctx := testContextWithAccount(t, 1)
	jobs, total, err := getJobs(1, 10, ctx)
	if err != nil {
		t.Fatalf("getJobs returned error: %v", err)
	}

	if total != 2 || len(jobs) != 2 {
		t.Fatalf("expected 2 jobs for account 1, got total=%d len=%d", total, len(jobs))
	}

	t.Logf("✅ Recursive directory scanning works correctly: found %d jobs", total)
}

// Test GetJobs handler with mixed job types
func TestGetJobsHandlerMixedTypes(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	queuesPath := util.QueuesPath()
	nowDir := filepath.Join(queuesPath, "now")
	inDir := filepath.Join(queuesPath, "in")
	everyDir := filepath.Join(queuesPath, "every")

	os.MkdirAll(nowDir, 0o755)
	os.MkdirAll(inDir, 0o755)
	os.MkdirAll(everyDir, 0o755)

	// Create mixed format jobs
	legacyJob := persistedJobInfo{ID: "legacy", RecipeName: "Legacy Job", Username: "user1", AccountID: 1, JobType: "NOW", Status: "Running", CreatedAt: time.Now()}
	writeJob(t, nowDir, "legacy.json", legacyJob)

	timestamp := time.Now().UnixNano()
	writeYeschefStepJob(t, inDir, "yeschef-in.json", "IN", "YesChef IN Job", "user1", 1, timestamp)
	writeYeschefStepJob(t, everyDir, "yeschef-every.json", "EVERY", "YesChef EVERY Job", "user1", 1, timestamp)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/lemc/account/jobs?page=1&limit=10", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	user := &models.User{Account: &models.Account{ID: 1}}
	uc := &models.UserContext{ActingAs: user, LoggedInAs: user}
	ctx := middleware.SetUserContext(c, uc)

	if err := GetJobs(ctx); err != nil {
		t.Fatalf("GetJobs returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if strings.Contains(body, "No jobs found") {
		t.Fatalf("unexpected 'No jobs found' in response")
	}

	// Check that all job names are present
	expectedJobs := []string{"Legacy Job", "YesChef IN Job", "YesChef EVERY Job"}
	for _, jobName := range expectedJobs {
		if !strings.Contains(body, jobName) {
			t.Errorf("response missing expected job name: %s", jobName)
		}
	}

	t.Logf("✅ Handler correctly renders mixed job types")
}

// Demonstration test showing the fix for the original issue
func TestYeschefJobParsingFix(t *testing.T) {
	// This test demonstrates the fix for the issue where IN jobs (yeschef format)
	// were not showing up in the jobs list because they were being parsed as
	// empty persistedJobInfo structs instead of being processed as yeschef format.

	cleanup := setupTestEnv(t)
	defer cleanup()

	queuesPath := util.QueuesPath()
	inDir := filepath.Join(queuesPath, "in")
	os.MkdirAll(inDir, 0o755)

	// Create a yeschef format job file (this is the actual format used by the yeschef system)
	yeschefJobData := `{
		"job": {
			"Step": {
				"Step": 1,
				"Name": "terraform destroy",
				"Image": "us-east5-docker.pkg.dev/holidayhack2025/hhc25/lemc-gcp-terraform-debian:latest",
				"Do": "in.5.minutes",
				"Timeout": "10.minutes"
			},
			"RecipeJob": {
				"JobType": "cookbook",
				"UUID": "0196fe96-c31a-7678-b28d-8cad0afee525",
				"CookbookID": "17",
				"UserID": "1",
				"Username": "alpha-owner",
				"PageID": "1",
				"Scope": "individual",
				"Recipe": {
					"Name": "delayed terraform destroy",
					"Description": "This runs terraform destroy to teardown all assets"
				}
			}
		},
		"job_key": "[cookbook][individual][userid:1][page:1][uuid:0196fe96-c31a-7678-b28d-8cad0afee525][queue:in]",
		"description": "StepJob: in.5.minutes",
		"group": "[userid:1][page:1][uuid:0196fe96-c31a-7678-b28d-8cad0afee525][group]",
		"trigger": "RunOnceTrigger::5m0s::expired",
		"next_run_time": 1748466187227539000
	}`

	jobFile := filepath.Join(inDir, "1748466187227539000.json")
	if err := os.WriteFile(jobFile, []byte(yeschefJobData), 0o644); err != nil {
		t.Fatalf("failed to write yeschef job file: %v", err)
	}

	ctx := testContextWithAccount(t, 1)
	jobs, total, err := getJobs(1, 10, ctx)
	if err != nil {
		t.Fatalf("getJobs returned error: %v", err)
	}

	// Before the fix: total would be 0 because the yeschef job would be parsed
	// as an empty persistedJobInfo struct and filtered out
	// After the fix: total should be 1 because the yeschef job is properly parsed
	if total != 1 || len(jobs) != 1 {
		t.Fatalf("expected 1 job for account 1, got total=%d len=%d", total, len(jobs))
	}

	job := jobs[0]

	// Verify the job was parsed correctly from yeschef format
	if job.RecipeName != "delayed terraform destroy" {
		t.Errorf("expected recipe name 'delayed terraform destroy', got '%s'", job.RecipeName)
	}
	if job.Type != "IN" {
		t.Errorf("expected job type 'IN', got '%s'", job.Type)
	}
	if job.Username != "alpha-owner" {
		t.Errorf("expected username 'alpha-owner', got '%s'", job.Username)
	}
	if job.Status != "Scheduled" {
		t.Errorf("expected status 'Scheduled', got '%s'", job.Status)
	}

	t.Logf("✅ Successfully parsed yeschef format IN job: %s (Type: %s, Status: %s)",
		job.RecipeName, job.Type, job.Status)
}
