package yeschef

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/jaredfolkins/letemcook/models"
	"github.com/reugn/go-quartz/quartz"
)

func DoIn(st models.Step, job *JobRecipe) error {

	re := regexp.MustCompile(`^in\.(\d+)\.(\w+)$`)
	m := re.FindStringSubmatch(st.Do)
	if len(m) < 2 {
		return fmt.Errorf("do.in regex matches total failed")
	}

	digit, err := strconv.Atoi(m[1])
	if err != nil {
		return err
	}

	ts, err := Ts(m[2])
	if err != nil {
		return err
	}

	k := LemcJobKey(job, IN_QUEUE)

	dij := &StepJob{Step: st, RecipeJob: job}
	kg := quartz.NewJobKeyWithGroup(k, jobGroup(job.UserID, job.PageID, job.UUID))
	detail := quartz.NewJobDetail(dij, kg)
	err = XoxoX.InScheduler.ScheduleJob(detail, quartz.NewRunOnceTrigger(time.Duration(digit)*ts))
	if err != nil {
		return err
	}

	return nil
}
