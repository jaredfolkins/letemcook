package yeschef

import (
	"fmt"
	"os"

	"github.com/reugn/go-quartz/logger"
	"github.com/reugn/go-quartz/quartz"
)

func (jq *jobQueue) ScheduledJobs(matchers []quartz.Matcher[quartz.ScheduledJob]) ([]quartz.ScheduledJob, error) {
	jq.mu.Lock()
	defer jq.mu.Unlock()

	logger.Trace("ScheduledJobs")
	var jobs []quartz.ScheduledJob
	fileInfo, err := os.ReadDir(jq.Path)
	if err != nil {
		return nil, err
	}

	for _, file := range fileInfo {
		if !file.IsDir() {
			data, err := os.ReadFile(fmt.Sprintf("%s/%s", jq.Path, file.Name()))
			if err == nil {
				var erri error
				var job quartz.ScheduledJob

				switch jq.Name {
				case NOW_QUEUE:
					job, erri = unmarshalRecipeJob(data)
					if erri != nil {
						return nil, erri
					}
				case IN_QUEUE:
					job, erri = unmarshalInStepJob(data)
					if erri != nil {
						return nil, erri
					}
				case EVERY_QUEUE:
					job, erri = unmarshalEveryStepJob(data)
					if erri != nil {
						return nil, erri
					}
				default:
					return nil, fmt.Errorf("unknown queue name: %s", jq.Name)
				}

				if isMatch(job, matchers) {
					jobs = append(jobs, job)
				}
			}
		}
	}

	return jobs, nil
}

func isMatch(job quartz.ScheduledJob, matchers []quartz.Matcher[quartz.ScheduledJob]) bool {
	for _, matcher := range matchers {
		if !matcher.IsMatch(job) {
			return false
		}
	}
	return true
}

var _ quartz.ScheduledJob = (*scheduledLemcJob)(nil)

func (job *scheduledLemcJob) JobDetail() *quartz.JobDetail {
	return job.jobDetail
}
func (job *scheduledLemcJob) Trigger() quartz.Trigger {
	return job.trigger
}
func (job *scheduledLemcJob) NextRunTime() int64 {
	return job.nextRunTime
}
