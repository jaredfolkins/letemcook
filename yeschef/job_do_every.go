package yeschef

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/jaredfolkins/letemcook/models"
	"github.com/reugn/go-quartz/quartz"
)

func DoEvery(st models.Step, job *JobRecipe) error {

	log.Printf("DoEvery: %v\n", st)
	re := regexp.MustCompile(`^every\.(\d+)\.(\w+)$`)
	m := re.FindStringSubmatch(st.Do)
	if len(m) < 2 {
		return fmt.Errorf("do.every regex matches total failed")
	}

	digit, err := strconv.Atoi(m[1])
	if err != nil {
		return err
	}

	ts, err := Ts(m[2])
	if err != nil {
		return err
	}

	k := LemcJobKey(job, EVERY_QUEUE)
	nowKey := LemcJobKey(job, NOW_QUEUE)
	if XoxoX.RunningMan.IsRunning(nowKey) {
		return fmt.Errorf("error: a NOW job is already running for this recipe")
	}

	dij := &StepJob{Step: st, RecipeJob: job}
	kg := quartz.NewJobKeyWithGroup(k, jobGroup(job.UserID, job.PageID, job.UUID))
	detail := quartz.NewJobDetail(dij, kg)
	err = XoxoX.EveryScheduler.ScheduleJob(detail, quartz.NewSimpleTrigger(time.Duration(digit)*ts))
	if err != nil {
		return err
	}

	return nil
}
