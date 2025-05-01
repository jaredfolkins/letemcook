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

	job.StepID = fmt.Sprintf("%d", st.Step)
	job.Env = append(job.Env, PYTHON_UNBUFFERED)
	job.Env = append(job.Env, fmt.Sprintf(STEP_ID, job.StepID))
	job.Env = append(job.Env, fmt.Sprintf(LEMC_HTML_ID, job.UUID, job.PageID, job.Scope))
	job.Env = append(job.Env, fmt.Sprintf(LEMC_CSS_ID, job.UUID, job.PageID, job.Scope))
	job.Env = append(job.Env, fmt.Sprintf(LEMC_JS_ID, job.UUID, job.PageID, job.Scope))

	job.ContainerTimeoutInSeconds, err = timeoutInSeconds(st.Timeout)
	if err != nil {
		ctx.Done()
		return err
	}

	err = runContainer(xserver, job, st.Image)
	if err != nil {
		e := fmt.Errorf("runContainer failed: %v", err)
		ctx.Done()
		return e
	}

	return nil
}
