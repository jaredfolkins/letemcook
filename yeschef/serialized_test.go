package yeschef

import (
	"reflect"
	"testing"
	"time"

	"github.com/jaredfolkins/letemcook/models"
	"github.com/reugn/go-quartz/quartz"
)

func newStepJob() *StepJob {
	return &StepJob{Step: models.Step{Step: 1, Do: "run"}, RecipeJob: &JobRecipe{}}
}

func TestMarshalUnmarshalEveryStepJob(t *testing.T) {
	jd := quartz.NewJobDetail(newStepJob(), quartz.NewJobKey("k"))
	trg := quartz.NewSimpleTrigger(time.Second)
	sj := &scheduledLemcJob{jobDetail: jd, trigger: trg, nextRunTime: 5}
	b, err := marshal(sj)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	j, err := unmarshalEveryStepJob(b)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if j.JobDetail().JobKey().Name() != "k" {
		t.Errorf("jobkey %s", j.JobDetail().JobKey().Name())
	}
}

func TestMarshalUnmarshalRecipeJob(t *testing.T) {
	rj := &JobRecipe{UUID: "u", JobType: "now"}
	jd := quartz.NewJobDetail(rj, quartz.NewJobKey("rk"))
	trg := quartz.NewRunOnceTrigger(time.Second)
	sj := &scheduledLemcJob{jobDetail: jd, trigger: trg, nextRunTime: 7}
	b, err := marshal(sj)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	j, err := unmarshalRecipeJob(b)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if j.JobDetail().JobKey().Name() != "rk" {
		t.Errorf("jobkey %s", j.JobDetail().JobKey().Name())
	}
}

func TestMarshalUnmarshalInStepJob(t *testing.T) {
	jd := quartz.NewJobDetail(newStepJob(), quartz.NewJobKey("ik"))
	trg := quartz.NewRunOnceTrigger(time.Second)
	sj := &scheduledLemcJob{jobDetail: jd, trigger: trg, nextRunTime: 9}
	b, err := marshal(sj)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	j, err := unmarshalInStepJob(b)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if j.JobDetail().JobKey().Name() != "ik" {
		t.Errorf("jobkey %s", j.JobDetail().JobKey().Name())
	}
	if !reflect.DeepEqual(j.JobDetail().Job().(*StepJob).Step, sj.jobDetail.Job().(*StepJob).Step) {
		t.Errorf("step mismatch")
	}
}
