package handlers

import (
	"encoding/json"
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
	"github.com/labstack/echo/v4"
)

// helper to create a context with a user belonging to given account ID
func testContextWithAccount(t *testing.T, accountID int64) LemcContext {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	user := &models.User{Account: &models.Account{ID: accountID}}
	uc := &models.UserContext{ActingAs: user, LoggedInAs: user}
	return middleware.SetUserContext(c, uc)
}

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

func TestGetJobsRecursiveFiltering(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LEMC_DATA", tmp)
	t.Setenv("LEMC_ENV", "test")

	nowDir := filepath.Join(util.QueuesPath(), "now")
	nestedDir := filepath.Join(nowDir, "nested")
	os.MkdirAll(nestedDir, 0o755)
	otherDir := filepath.Join(util.QueuesPath(), "every")
	os.MkdirAll(otherDir, 0o755)

	j1 := persistedJobInfo{ID: "1", RecipeName: "r1", Username: "u1", AccountID: 1, JobType: "NOW", Status: "Running", CreatedAt: time.Now()}
	j2 := persistedJobInfo{ID: "2", RecipeName: "r2", Username: "u1", AccountID: 1, JobType: "NOW", Status: "Running", CreatedAt: time.Now()}
	j3 := persistedJobInfo{ID: "3", RecipeName: "r3", Username: "u2", AccountID: 2, JobType: "EVERY", Status: "Running", CreatedAt: time.Now()}

	writeJob(t, nowDir, "j1.json", j1)
	writeJob(t, nestedDir, "j2.json", j2)
	writeJob(t, otherDir, "j3.json", j3)

	ctx := testContextWithAccount(t, 1)
	jobs, total, err := getJobs(1, 10, ctx)
	if err != nil {
		t.Fatalf("getJobs returned error: %v", err)
	}

	if total != 2 || len(jobs) != 2 {
		t.Fatalf("expected 2 jobs for account 1, got total=%d len=%d", total, len(jobs))
	}

	ctx2 := testContextWithAccount(t, 2)
	jobs, total, err = getJobs(1, 10, ctx2)
	if err != nil {
		t.Fatalf("getJobs returned error: %v", err)
	}
	if total != 1 || len(jobs) != 1 {
		t.Fatalf("expected 1 job for account 2, got total=%d len=%d", total, len(jobs))
	}
}

func TestGetJobsHandlerRendersJobs(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LEMC_DATA", tmp)
	t.Setenv("LEMC_ENV", "test")

	nowDir := filepath.Join(util.QueuesPath(), "now")
	if err := os.MkdirAll(nowDir, 0o755); err != nil {
		t.Fatalf("failed to create now dir: %v", err)
	}

	j1 := persistedJobInfo{ID: "1", RecipeName: "r1", Username: "u1", AccountID: 1, JobType: "NOW", Status: "Running", CreatedAt: time.Now()}
	j2 := persistedJobInfo{ID: "2", RecipeName: "r2", Username: "u1", AccountID: 1, JobType: "NOW", Status: "Running", CreatedAt: time.Now()}

	writeJob(t, nowDir, "j1.json", j1)
	writeJob(t, nowDir, "j2.json", j2)

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
		t.Fatalf("unexpected 'No jobs found' in response: %s", body)
	}
	if !strings.Contains(body, j1.RecipeName) || !strings.Contains(body, j2.RecipeName) {
		t.Fatalf("response missing expected job names: %s", body)
	}
}
