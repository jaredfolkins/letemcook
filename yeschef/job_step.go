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
	err := DoStep(ctx, dij.RecipeJob, dij.Step)
	if err != nil {
		ctx.Done()
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
