package yeschef

import (
	"context"
	"testing"

	"github.com/jaredfolkins/letemcook/models"
)

// TestStepJobExecuteError ensures Execute returns an error when DoStep fails.
func TestStepJobExecuteError(t *testing.T) {
	job := &JobRecipe{} // missing UserID to force error in DoStep
	st := models.Step{Step: 1}
	sj := &StepJob{Step: st, RecipeJob: job}

	err := sj.Execute(context.Background())
	if err == nil {
		t.Fatalf("expected error when DoStep fails")
	}
}

// TestStepJobDescription verifies the description string.
func TestStepJobDescription(t *testing.T) {
	sj := &StepJob{Step: models.Step{Do: "echo"}}
	if sj.Description() != "StepJob: echo" {
		t.Fatalf("unexpected description: %s", sj.Description())
	}
}

// TestCreateTaskID tests all branches of CreateTaskID.
func TestCreateTaskID(t *testing.T) {
	job := &JobRecipe{UserID: "u", UUID: "id", PageID: "p"}
	got, err := job.CreateTaskID("queue")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "[userid:u][pageid:p][uuid:id][queue:queue]" {
		t.Fatalf("unexpected task id: %s", got)
	}

	if _, err := (&JobRecipe{UUID: "id", PageID: "p"}).CreateTaskID("q"); err == nil {
		t.Fatalf("expected error for missing UserID")
	}
	if _, err := (&JobRecipe{UserID: "u", PageID: "p"}).CreateTaskID("q"); err == nil {
		t.Fatalf("expected error for missing UUID")
	}
	if _, err := (&JobRecipe{UserID: "u", UUID: "id"}).CreateTaskID("q"); err == nil {
		t.Fatalf("expected error for missing PageID")
	}
}
