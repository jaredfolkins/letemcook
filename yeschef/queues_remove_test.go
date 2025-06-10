package yeschef

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jaredfolkins/letemcook/models"
	"github.com/reugn/go-quartz/quartz"
)

// TestRemoveInvalidJobFile ensures Remove deletes an invalid job file and returns a useful error.
func TestRemoveInvalidJobFile(t *testing.T) {
	dir := t.TempDir()
	q := &jobQueue{Path: dir, Name: IN_QUEUE}

	// Create an invalid step job (missing RecipeJob)
	stepJob := &StepJob{Step: models.Step{Step: 1, Name: "", Image: ""}, RecipeJob: nil}
	jd := quartz.NewJobDetail(stepJob, quartz.NewJobKey("invalid"))
	sj := &scheduledLemcJob{jobDetail: jd, trigger: quartz.NewRunOnceTrigger(time.Second), nextRunTime: time.Now().Unix()}
	data, err := marshal(sj)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	path := filepath.Join(dir, "invalid.json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	_, err = q.Remove(quartz.NewJobKey("invalid"))
	if err == nil {
		t.Fatal("expected error from Remove")
	}
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Fatalf("expected file to be removed, stat err=%v", statErr)
	}

	if err.Error() == "removed invalid job file: <nil>" {
		t.Fatalf("unexpected error message: %v", err)
	}
}
