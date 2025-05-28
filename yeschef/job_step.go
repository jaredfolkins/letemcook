package yeschef

import (
	"context"
	"fmt"
	"log"

	"github.com/jaredfolkins/letemcook/models"
)

type StepJob struct {
	Step      models.Step
	RecipeJob *JobRecipe
}

func (dij *StepJob) Execute(ctx context.Context) error {
	log.Println("StepJob: Execute")
	log.Printf("StepJob: %v %v %v\n", dij.Step.Image, dij.Step.Step, dij.Step.Do)

	execCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Only check for NOW job conflicts if we have a valid recipe with all required fields
	// This prevents panics in test scenarios with incomplete JobRecipe structs
	if dij.RecipeJob != nil && dij.RecipeJob.Scope != "" && dij.RecipeJob.UserID != "" && dij.RecipeJob.UUID != "" && dij.RecipeJob.PageID != "" {
		// Check if a NOW job is already running for this recipe before executing the delayed step
		nowKey := LemcJobKey(dij.RecipeJob, NOW_QUEUE)

		XoxoX.RunningMan.mu.Lock()
		if XoxoX.RunningMan.list[nowKey] {
			XoxoX.RunningMan.mu.Unlock()
			return fmt.Errorf("error: a NOW job is already running for this recipe")
		}

		// Mark this delayed job as a running NOW job during execution
		XoxoX.RunningMan.list[nowKey] = true
		XoxoX.RunningMan.mu.Unlock()

		// Ensure we clean up the running status when done
		defer func() {
			XoxoX.RunningMan.mu.Lock()
			delete(XoxoX.RunningMan.list, nowKey)
			XoxoX.RunningMan.mu.Unlock()
		}()
	}

	err := DoStep(execCtx, dij.RecipeJob, dij.Step)
	if err != nil {
		cancel()
		return err
	}
	return nil
}

func (dij *StepJob) Description() string {
	return fmt.Sprintf("StepJob: %s", dij.Step.Do)
}

func (job *JobRecipe) CreateTaskID(queue string) (string, error) {
	var s string
	if len(job.UserID) == 0 {
		return s, fmt.Errorf("JobRecipe UserID missing")
	} else if len(job.UUID) == 0 {
		return s, fmt.Errorf("JobRecipe UUID missing")
	} else if len(job.PageID) == 0 {
		return s, fmt.Errorf("JobRecipe PageID missing")
	}

	s = fmt.Sprintf("[userid:%s][pageid:%s][uuid:%s][queue:%s]", job.UserID, job.PageID, job.UUID, queue)
	return s, nil
}
