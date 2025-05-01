package yeschef

import (
	"context"
	"log"
	"strings"

	"github.com/jaredfolkins/letemcook/models"
)

type JobRecipe struct {
	JobType                   string
	UUID                      string
	CookbookID                string
	AppID                     string
	UserID                    string
	Username                  string
	PageID                    string
	StepID                    string
	Scope                     string
	Env                       []string
	ContainerTimeoutInSeconds int
	Recipe                    models.Recipe
	RecipientUserIDs          []int64 // Populated for shared jobs
}

func (job *JobRecipe) Execute(ctx context.Context) error {
	key := LemcJobKey(job, NOW_QUEUE)
	defer XoxoX.RunningMan.Remove(key)
	log.Printf("JobRecipe: %v \n", key)
	/*
		if err := DeleteCronJobsByPageAndUUID(job.PageID, job.UUID); err != nil {
			return err
		}
	*/

	steps := make(map[int]models.Step)
	for _, st := range job.Recipe.Steps {
		steps[st.Step] = st
	}

	for _, st := range job.Recipe.Steps {
		do := strings.Trim(st.Do, "")
		if lemc_do_now_rgx.MatchString(do) {
			err := DoStep(ctx, job, st)
			if err != nil {
				ctx.Done()
				return err
			}
		} else if lemc_do_in_rgx.MatchString(do) {
			err := DoIn(st, job)
			if err != nil {
				ctx.Done()
				return err
			}
		} else if lemc_do_every_rgx.MatchString(do) {
			err := DoEvery(st, job)
			if err != nil {
				ctx.Done()
				return err
			}
		}
	}
	return nil
}

func (rj *JobRecipe) Description() string {
	desc, err := rj.CreateTaskID(rj.JobType)
	if err != nil {
		log.Fatalf("%s", err)
	}
	return desc
}
