package yeschef

import (
	"fmt"

	"github.com/reugn/go-quartz/quartz"
)

// JobValidationError represents an error that occurs during job validation
type JobValidationError struct {
	Field   string
	Message string
}

func (e *JobValidationError) Error() string {
	return fmt.Sprintf("job validation error for %s: %s", e.Field, e.Message)
}

// ValidateScheduledJob performs comprehensive validation of a ScheduledJob
// to prevent nil pointer dereferences and other runtime errors.
func ValidateScheduledJob(job quartz.ScheduledJob) error {
	if job == nil {
		return &JobValidationError{
			Field:   "job",
			Message: "scheduled job is nil",
		}
	}

	// Validate JobDetail
	jobDetail := job.JobDetail()
	if jobDetail == nil {
		return &JobValidationError{
			Field:   "job_detail",
			Message: "job detail is nil",
		}
	}

	// Validate JobKey
	jobKey := jobDetail.JobKey()
	if jobKey == nil {
		return &JobValidationError{
			Field:   "job_key",
			Message: "job key is nil",
		}
	}

	// Validate JobKey name
	if jobKey.Name() == "" {
		return &JobValidationError{
			Field:   "job_key_name",
			Message: "job key name is empty",
		}
	}

	// Validate Job instance
	jobInstance := jobDetail.Job()
	if jobInstance == nil {
		return &JobValidationError{
			Field:   "job_instance",
			Message: "job instance is nil",
		}
	}

	// Validate Trigger
	trigger := job.Trigger()
	if trigger == nil {
		return &JobValidationError{
			Field:   "trigger",
			Message: "trigger is nil",
		}
	}

	// Additional validation based on job type
	switch j := jobInstance.(type) {
	case *StepJob:
		if err := validateStepJob(j); err != nil {
			return err
		}
	case *JobRecipe:
		if err := validateJobRecipe(j); err != nil {
			return err
		}
	}

	return nil
}

// validateStepJob validates a StepJob instance
func validateStepJob(stepJob *StepJob) error {
	if stepJob == nil {
		return &JobValidationError{
			Field:   "step_job",
			Message: "step job is nil",
		}
	}

	// Validate Step
	if stepJob.Step.Name == "" {
		return &JobValidationError{
			Field:   "step_name",
			Message: "step name is empty",
		}
	}

	if stepJob.Step.Image == "" {
		return &JobValidationError{
			Field:   "step_image",
			Message: "step image is empty",
		}
	}

	// Validate RecipeJob
	if stepJob.RecipeJob == nil {
		return &JobValidationError{
			Field:   "recipe_job",
			Message: "recipe job is nil",
		}
	}

	return validateJobRecipe(stepJob.RecipeJob)
}

// validateJobRecipe validates a JobRecipe instance
func validateJobRecipe(recipeJob *JobRecipe) error {
	if recipeJob == nil {
		return &JobValidationError{
			Field:   "recipe_job",
			Message: "recipe job is nil",
		}
	}

	if recipeJob.UUID == "" {
		return &JobValidationError{
			Field:   "recipe_uuid",
			Message: "recipe UUID is empty",
		}
	}

	if recipeJob.UserID == "" {
		return &JobValidationError{
			Field:   "recipe_user_id",
			Message: "recipe user ID is empty",
		}
	}

	if recipeJob.PageID == "" {
		return &JobValidationError{
			Field:   "recipe_page_id",
			Message: "recipe page ID is empty",
		}
	}

	if recipeJob.Scope == "" {
		return &JobValidationError{
			Field:   "recipe_scope",
			Message: "recipe scope is empty",
		}
	}

	// Validate scope is one of the allowed values
	if recipeJob.Scope != "individual" && recipeJob.Scope != "shared" {
		return &JobValidationError{
			Field:   "recipe_scope",
			Message: fmt.Sprintf("recipe scope '%s' is not valid (must be 'individual' or 'shared')", recipeJob.Scope),
		}
	}

	return nil
}

// ValidateSystemState validates that required system components are available
func ValidateSystemState() error {
	if XoxoX == nil {
		return &JobValidationError{
			Field:   "xoxox",
			Message: "XoxoX global instance is nil",
		}
	}

	if XoxoX.RunningMan == nil {
		return &JobValidationError{
			Field:   "running_man",
			Message: "RunningMan instance is nil",
		}
	}

	return nil
}
