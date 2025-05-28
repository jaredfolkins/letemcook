package yeschef

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/reugn/go-quartz/quartz"
)

func TestNewQuartzQueue(t *testing.T) {
	q := NewQuartzQueue("test")
	if q.Name != "test" {
		t.Fatalf("unexpected name: %s", q.Name)
	}
}

func TestJobQueueSizeEmpty(t *testing.T) {
	q := &jobQueue{Path: t.TempDir(), Name: "temp"}
	if n, err := q.Size(); err != nil || n != 0 {
		t.Fatalf("expected empty queue, got n=%d err=%v", n, err)
	}
}

func TestJobQueueRecover(t *testing.T) {
	tmpDir := t.TempDir()
	q := &jobQueue{Path: tmpDir, Name: NOW_QUEUE}
	config := quartz.StdSchedulerOptions{}

	job := &JobRecipe{UUID: "u", JobType: "now", UserID: "test-user", PageID: "test-page"}
	jd := quartz.NewJobDetail(job, quartz.NewJobKey("recover"))
	trg := quartz.NewRunOnceTrigger(time.Hour)
	next := time.Now().Add(time.Hour).Unix()
	sj := &scheduledLemcJob{jobDetail: jd, trigger: trg, nextRunTime: next}

	if err := q.Push(sj); err != nil {
		t.Fatalf("push: %v", err)
	}

	origPath := filepath.Join(tmpDir, fmt.Sprintf("%d.json", next))
	if _, err := os.Stat(origPath); err != nil {
		t.Fatalf("job file missing: %v", err)
	}

	newScheduler := quartz.NewStdSchedulerWithOptions(config, q, nil)
	if err := q.Recover(newScheduler); err != nil {
		t.Fatalf("recover: %v", err)
	}

	if _, err := os.Stat(origPath); !os.IsNotExist(err) {
		t.Fatalf("original job file still exists")
	}

	jobs, err := q.ScheduledJobs(nil)
	if err != nil {
		t.Fatalf("scheduled jobs: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	if jobs[0].JobDetail().JobKey().Name() != "recover" {
		t.Fatalf("wrong job key: %s", jobs[0].JobDetail().JobKey().Name())
	}
}
