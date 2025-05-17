package yeschef

import (
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/jaredfolkins/letemcook/util"
	"github.com/reugn/go-quartz/logger"
	"github.com/reugn/go-quartz/quartz"
)

var _ quartz.Job = (*JobRecipe)(nil)
var _ quartz.Job = (*StepJob)(nil)

type scheduledLemcJob struct {
	jobDetail   *quartz.JobDetail
	trigger     quartz.Trigger
	nextRunTime int64
}

type serializedRecipeJob struct {
	Job         *JobRecipe               `json:"job"`
	JobKey      string                   `json:"job_key"`
	Description string                   `json:"description"`
	Group       string                   `json:"group"`
	Options     *quartz.JobDetailOptions `json:"job_options"`
	Trigger     string                   `json:"trigger"`
	NextRunTime int64                    `json:"next_run_time"`
}

type serializedStepJob struct {
	Job         *StepJob                 `json:"job"`
	JobKey      string                   `json:"job_key"`
	Description string                   `json:"description"`
	Group       string                   `json:"group"`
	Options     *quartz.JobDetailOptions `json:"job_options"`
	Trigger     string                   `json:"trigger"`
	NextRunTime int64                    `json:"next_run_time"`
}

type serializedJob struct {
	Job         quartz.Job               `json:"job"`
	JobKey      string                   `json:"job_key"`
	Description string                   `json:"description"`
	Group       string                   `json:"group"`
	Options     *quartz.JobDetailOptions `json:"job_options"`
	Trigger     string                   `json:"trigger"`
	NextRunTime int64                    `json:"next_run_time"`
}

type jobQueue struct {
	mu   sync.Mutex
	Path string
	Name string
}

var _ quartz.JobQueue = (*jobQueue)(nil)

func NewQuartzQueue(name string) *jobQueue {
	// err := godotenv.Load()
	// if err != nil {
	// \tlogger.Errorf("Error loading .env file: %s", err)
	// }
	dataFolder := util.QueuesPath()
	if _, err := os.Stat(dataFolder); os.IsNotExist(err) {
		err = os.MkdirAll(dataFolder, FILE_MODE)
		if err != nil {
			log.Fatalf("failed to create queues directory: %v", err)
		}
	}

	path := filepath.Join(dataFolder, name)
	return &jobQueue{Path: path, Name: name}
}

func (jq *jobQueue) Push(job quartz.ScheduledJob) error {
	jq.mu.Lock()
	defer jq.mu.Unlock()

	logger.Debugf("Push job: %s", job.JobDetail().JobKey())

	var err error
	var serialized []byte
	serialized, err = marshal(job)
	if err != nil {
		return err
	}

	if err = os.WriteFile(fmt.Sprintf("%s/%d", jq.Path, job.NextRunTime()),
		serialized, FILE_MODE); err != nil {
		logger.Errorf("Failed to write job: %s", err)
		return err
	}
	return nil
}

func (jq *jobQueue) Pop() (quartz.ScheduledJob, error) {
	jq.mu.Lock()
	defer jq.mu.Unlock()

	job, err := findHead(jq)

	if err == nil {
		err = os.Remove(fmt.Sprintf("%s/%d", jq.Path, job.NextRunTime()))
		if err != nil {
			logger.Errorf("Failed to delete job: %s", err)
			return nil, err
		}

		key := job.JobDetail().JobKey().Name()
		logger.Debugf("Adding job to RunningMan: %s", key)
		XoxoX.RunningMan.Add(key)
		return job, nil
	}

	return nil, nil
}

func (jq *jobQueue) Head() (quartz.ScheduledJob, error) {
	jq.mu.Lock()
	defer jq.mu.Unlock()

	job, err := findHead(jq)
	if err != nil && err.Error() != "no jobs found" {
		logger.Errorf("Failed to find job head: %s", err)
	}
	return job, err
}

func findHead(jq *jobQueue) (quartz.ScheduledJob, error) {
	fileInfo, err := os.ReadDir(jq.Path)
	if err != nil {
		return nil, err
	}

	var lastUpdate int64 = math.MaxInt64
	for _, file := range fileInfo {
		if !file.IsDir() {
			time, err := strconv.ParseInt(file.Name(), 10, 64)
			if err == nil && time < lastUpdate {
				lastUpdate = time
			}
		}
	}

	if lastUpdate == math.MaxInt64 {
		return nil, errors.New("no jobs found")
	}

	data, err := os.ReadFile(fmt.Sprintf("%s/%d", jq.Path, lastUpdate))
	if err != nil {
		return nil, err
	}

	var job quartz.ScheduledJob

	switch jq.Name {
	case NOW_QUEUE:
		var err error
		job, err = unmarshalRecipeJob(data)
		if err != nil {
			return nil, err
		}
	case IN_QUEUE:
		job, err = unmarshalInStepJob(data)
		if err != nil {
			return nil, err
		}
	case EVERY_QUEUE:
		job, err = unmarshalEveryStepJob(data)
		if err != nil {
			return nil, err
		}
	}

	return job, nil
}

func (jq *jobQueue) Get(jobKey *quartz.JobKey) (quartz.ScheduledJob, error) {
	jq.mu.Lock()
	defer jq.mu.Unlock()

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
				}

				if jobKey.Equals(job.JobDetail().JobKey()) {
					return job, nil
				}
			}
		}
	}
	return nil, errors.New("no jobs found")
}

func (jq *jobQueue) Remove(jobKey *quartz.JobKey) (quartz.ScheduledJob, error) {
	jq.mu.Lock()
	defer jq.mu.Unlock()

	logger.Debugf("Removing job: %s", jobKey.Name())

	fileInfo, err := os.ReadDir(jq.Path)
	if err != nil {
		return nil, err
	}
	for _, file := range fileInfo {
		if !file.IsDir() {
			path := fmt.Sprintf("%s/%s", jq.Path, file.Name())
			data, err := os.ReadFile(path)
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
				}

				if jobKey.Name() == job.JobDetail().JobKey().Name() {
					if err = os.Remove(path); err == nil {
						return job, nil
					}
				}
			}
		}
	}
	return nil, errors.New("no jobs found")
}

func (jq *jobQueue) Size() (int, error) {
	jq.mu.Lock()
	defer jq.mu.Unlock()

	files, err := os.ReadDir(jq.Path)
	if err != nil {
		return 0, err
	}
	return len(files), nil
}

func (jq *jobQueue) Clear() error {
	jq.mu.Lock()
	defer jq.mu.Unlock()

	logger.Infof("Clearing queue: %s", jq.Name)
	return os.RemoveAll(jq.Path)
}
