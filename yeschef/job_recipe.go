package yeschef

import (
	"context"
	"log"
	"strconv"
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
	execCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	key := LemcJobKey(job, NOW_QUEUE)
	defer XoxoX.RunningMan.Remove(key)
	log.Printf("JobRecipe: %v \n", key)

	var srv *McpServer
	if job.AppID != "" {
		if id, err := strconv.ParseInt(job.AppID, 10, 64); err == nil {
			srv = XoxoX.ReadMcpAppInstance(id)
		}
	}
	if srv != nil {
		srv.broadcast([]byte("--MCP JOB STARTED--"))
	}
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
			err := DoStep(execCtx, job, st)
			if err != nil {
				cancel()
				return err
			}
		} else if lemc_do_in_rgx.MatchString(do) {
			err := DoIn(st, job)
			if err != nil {
				cancel()
				return err
			}
		} else if lemc_do_every_rgx.MatchString(do) {
			err := DoEvery(st, job)
			if err != nil {
				cancel()
				return err
			}
		}
	}
	if srv != nil {
		srv.broadcast([]byte("--MCP JOB FINISHED--"))
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
