package yeschef

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jaredfolkins/letemcook/models"
)

func DoStep(ctx context.Context, job *JobRecipe, st models.Step) error {
	userid, err := strconv.ParseInt(job.UserID, 10, 64)
	if err != nil {
		return err
	}

	xserver := XoxoX.CreadInstance(userid)

	// Prepare step-specific environment variables
	// Start with a copy of the job-level environment variables
	stepEnv := make([]string, len(job.Env))
	copy(stepEnv, job.Env)

	// Append system-defined step env vars
	stepEnv = append(stepEnv, PYTHON_UNBUFFERED)
	stepEnv = append(stepEnv, fmt.Sprintf(STEP_ID, st.Step))
	stepEnv = append(stepEnv, fmt.Sprintf(LEMC_HTML_ID, job.UUID, job.PageID, job.Scope))
	stepEnv = append(stepEnv, fmt.Sprintf(LEMC_CSS_ID, job.UUID, job.PageID, job.Scope))
	stepEnv = append(stepEnv, fmt.Sprintf(LEMC_JS_ID, job.UUID, job.PageID, job.Scope))

	// Append user-defined env vars from the step configuration
	if len(st.Env) > 0 {
		stepEnv = append(stepEnv, st.Env...)
	}

	// Work on a copy of the job to avoid races when multiple steps run concurrently
	jobCopy := *job
	jobCopy.StepID = fmt.Sprintf("%d", st.Step)

	jobCopy.ContainerTimeoutInSeconds, err = timeoutInSeconds(st.Timeout)
	if err != nil {
		return err
	}

	err = runContainer(xserver, &jobCopy, st.Image, stepEnv)
	if err != nil {
		e := fmt.Errorf("runContainer failed: %v", err)
		return e
	}

	return nil
}
