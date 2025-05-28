package yeschef

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/reugn/go-quartz/job"
	"github.com/reugn/go-quartz/quartz"
)

func marshal(job quartz.ScheduledJob) ([]byte, error) {
	var serialized serializedJob
	serialized.Job = job.JobDetail().Job()
	serialized.Description = job.JobDetail().Job().Description()
	serialized.JobKey = job.JobDetail().JobKey().Name()
	serialized.Options = job.JobDetail().Options()
	serialized.Trigger = job.Trigger().Description()
	serialized.Group = job.JobDetail().JobKey().Group()
	serialized.NextRunTime = job.NextRunTime()
	return json.Marshal(serialized)
}

func unmarshalEveryStepJob(encoded []byte) (quartz.ScheduledJob, error) {
	var nj serializedStepJob
	if err := json.Unmarshal(encoded, &nj); err != nil {
		return nil, err
	}

	if nj.Job == nil {
		return nil, fmt.Errorf("job is nil in serializedStepJob")
	}

	jobKey := quartz.NewJobKeyWithGroup(nj.JobKey, nj.Group)
	jobDetail := quartz.NewJobDetailWithOptions(nj.Job, jobKey, nj.Options)
	triggerOpts := strings.Split(nj.Trigger, quartz.Sep)

	if len(triggerOpts) < 2 {
		return nil, fmt.Errorf("invalid trigger format: %s", nj.Trigger)
	}

	interval, err := time.ParseDuration(triggerOpts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid interval format: %s", triggerOpts[1])
	}

	var trigger quartz.Trigger
	trigger = quartz.NewSimpleTrigger(interval)

	return &scheduledLemcJob{
		jobDetail:   jobDetail,
		trigger:     trigger,
		nextRunTime: nj.NextRunTime,
	}, nil
}

func unmarshalRecipeJob(encoded []byte) (quartz.ScheduledJob, error) {
	var nj serializedRecipeJob
	if err := json.Unmarshal(encoded, &nj); err != nil {
		return nil, err
	}

	if nj.Job == nil {
		return nil, fmt.Errorf("job is nil in serializedRecipeJob")
	}

	jk := quartz.NewJobKeyWithGroup(nj.JobKey, nj.Group)
	isolatedJob := job.NewIsolatedJob(nj.Job)
	jobDetail := quartz.NewJobDetailWithOptions(isolatedJob, jk, nj.Options)
	triggerOpts := strings.Split(nj.Trigger, quartz.Sep)

	if len(triggerOpts) < 2 {
		return nil, fmt.Errorf("invalid trigger format: %s", nj.Trigger)
	}

	interval, err := time.ParseDuration(triggerOpts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid interval format: %s", triggerOpts[1])
	}

	var trigger quartz.Trigger
	trigger = quartz.NewRunOnceTrigger(interval)
	if len(triggerOpts) == 3 {
		if triggerOpts[2] == "expired" {
			trigger.(*quartz.RunOnceTrigger).Expired = true
		}
	}

	return &scheduledLemcJob{
		jobDetail:   jobDetail,
		trigger:     trigger,
		nextRunTime: nj.NextRunTime,
	}, nil
}

func unmarshalInStepJob(encoded []byte) (quartz.ScheduledJob, error) {
	var nj serializedStepJob
	if err := json.Unmarshal(encoded, &nj); err != nil {
		return nil, err
	}

	if nj.Job == nil {
		return nil, fmt.Errorf("job is nil in serializedStepJob")
	}

	jobKey := quartz.NewJobKeyWithGroup(nj.JobKey, nj.Group)
	jobDetail := quartz.NewJobDetailWithOptions(nj.Job, jobKey, nj.Options)
	triggerOpts := strings.Split(nj.Trigger, quartz.Sep)

	if len(triggerOpts) < 2 {
		return nil, fmt.Errorf("invalid trigger format: %s", nj.Trigger)
	}

	interval, err := time.ParseDuration(triggerOpts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid interval format: %s", triggerOpts[1])
	}

	var trigger quartz.Trigger
	trigger = quartz.NewRunOnceTrigger(interval)

	if len(triggerOpts) == 3 {
		if triggerOpts[2] == "expired" {
			trigger.(*quartz.RunOnceTrigger).Expired = true
		}
	}

	return &scheduledLemcJob{
		jobDetail:   jobDetail,
		trigger:     trigger,
		nextRunTime: nj.NextRunTime,
	}, nil
}
