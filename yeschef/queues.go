package yeschef

import (
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

// Recover scans the queue directory for previously persisted jobs and
// schedules them with the provided scheduler. This allows the system to
// restore jobs after an unexpected shutdown or crash.
func (jq *jobQueue) Recover(s *quartz.StdScheduler) error {
	// First, collect all jobs while holding the mutex
	var jobsToSchedule []quartz.ScheduledJob
	var filesToRemove []string
	var invalidFilesToRemove []string

	jq.mu.Lock()
	files, err := os.ReadDir(jq.Path)
	if err != nil {
		jq.mu.Unlock()
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		path := filepath.Join(jq.Path, file.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			logger.Errorf("Recover read job: %v", err)
			continue
		}

		var job quartz.ScheduledJob
		switch jq.Name {
		case NOW_QUEUE:
			job, err = unmarshalRecipeJob(data)
		case IN_QUEUE:
			job, err = unmarshalInStepJob(data)
		case EVERY_QUEUE:
			job, err = unmarshalEveryStepJob(data)
		default:
			logger.Errorf("Recover unknown queue name: %s", jq.Name)
			continue
		}
		if err != nil {
			logger.Errorf("Recover unmarshal job: %v", err)
			invalidFilesToRemove = append(invalidFilesToRemove, path)
			continue
		}

		// Validate the job before scheduling
		if err := ValidateScheduledJob(job); err != nil {
			logger.Errorf("Recover job validation failed: %v", err)
			invalidFilesToRemove = append(invalidFilesToRemove, path)
			continue
		}

		jobsToSchedule = append(jobsToSchedule, job)
		filesToRemove = append(filesToRemove, path)
	}

	// Remove valid files while still holding the mutex
	for _, path := range filesToRemove {
		if err := os.Remove(path); err != nil {
			logger.Errorf("Recover remove job file: %v", err)
		}
	}

	// Remove invalid files to prevent future problems
	for _, path := range invalidFilesToRemove {
		if err := os.Remove(path); err != nil {
			logger.Errorf("Recover remove invalid job file: %v", err)
		} else {
			logger.Infof("Removed invalid job file during recovery: %s", path)
		}
	}
	jq.mu.Unlock()

	// Now schedule jobs without holding the mutex
	for _, job := range jobsToSchedule {
		if err := s.ScheduleJob(job.JobDetail(), job.Trigger()); err != nil {
			logger.Errorf("Recover schedule job: %v", err)
		}
	}

	return nil
}

var _ quartz.JobQueue = (*jobQueue)(nil)

func NewQuartzQueue(name string) *jobQueue {
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

	// Store with .json extension for UI compatibility
	if err = os.WriteFile(fmt.Sprintf("%s/%d.json", jq.Path, job.NextRunTime()),
		serialized, util.FilePerm); err != nil {
		logger.Errorf("Failed to write job: %s", err)
		return err
	}
	return nil
}

func (jq *jobQueue) Pop() (quartz.ScheduledJob, error) {
	jq.mu.Lock()
	defer jq.mu.Unlock()

	// First validate system state
	if err := ValidateSystemState(); err != nil {
		logger.Errorf("System validation failed: %v", err)
		return nil, err
	}

	job, err := findHead(jq)
	if err != nil {
		return nil, err
	}

	// Validate the job before proceeding
	if err := ValidateScheduledJob(job); err != nil {
		logger.Errorf("Job validation failed: %v", err)
		// Remove the invalid job file to prevent repeated failures
		timestamp := job.NextRunTime()
		jsonPath := fmt.Sprintf("%s/%d.json", jq.Path, timestamp)
		plainPath := fmt.Sprintf("%s/%d", jq.Path, timestamp)

		// Try to remove both possible file formats
		if removeErr := os.Remove(jsonPath); removeErr != nil {
			os.Remove(plainPath) // Try plain format if json fails
		}

		return nil, err
	}

	// Try both .json and non-.json filenames for backward compatibility
	timestamp := job.NextRunTime()
	jsonPath := fmt.Sprintf("%s/%d.json", jq.Path, timestamp)
	plainPath := fmt.Sprintf("%s/%d", jq.Path, timestamp)

	// Try to remove .json file first, then plain file
	err = os.Remove(jsonPath)
	if err != nil {
		err = os.Remove(plainPath)
	}

	if err != nil {
		logger.Errorf("Failed to delete job: %s", err)
		return nil, err
	}

	key := job.JobDetail().JobKey().Name()
	logger.Debugf("Adding job to RunningMan: %s", key)
	XoxoX.RunningMan.Add(key)
	return job, nil
}

func (jq *jobQueue) Head() (quartz.ScheduledJob, error) {
	jq.mu.Lock()
	defer jq.mu.Unlock()

	job, err := findHead(jq)
	if err != nil && err.Error() != "no jobs found" {
		logger.Errorf("Failed to find job head: %s", err)
		return nil, err
	}

	if job != nil {
		// Validate the job before returning
		if err := ValidateScheduledJob(job); err != nil {
			logger.Errorf("Job validation failed in Head: %v", err)
			return nil, err
		}
	}

	return job, err
}

func findHead(jq *jobQueue) (quartz.ScheduledJob, error) {
	fileInfo, err := os.ReadDir(jq.Path)
	if err != nil {
		return nil, err
	}

	var lastUpdate int64 = math.MaxInt64
	var filename string
	for _, file := range fileInfo {
		if !file.IsDir() {
			name := file.Name()
			var timeStr string

			// Handle both .json and non-.json filenames for backward compatibility
			if strings.HasSuffix(name, ".json") {
				timeStr = strings.TrimSuffix(name, ".json")
			} else {
				timeStr = name
			}

			time, err := strconv.ParseInt(timeStr, 10, 64)
			if err == nil && time < lastUpdate {
				lastUpdate = time
				filename = name
			}
		}
	}

	if lastUpdate == math.MaxInt64 {
		return nil, errors.New("no jobs found")
	}

	data, err := os.ReadFile(fmt.Sprintf("%s/%s", jq.Path, filename))
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
	default:
		return nil, fmt.Errorf("unknown queue name: %s", jq.Name)
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
						logger.Errorf("Failed to unmarshal recipe job in Get: %v", erri)
						continue
					}
				case IN_QUEUE:
					job, erri = unmarshalInStepJob(data)
					if erri != nil {
						logger.Errorf("Failed to unmarshal in step job in Get: %v", erri)
						continue
					}
				case EVERY_QUEUE:
					job, erri = unmarshalEveryStepJob(data)
					if erri != nil {
						logger.Errorf("Failed to unmarshal every step job in Get: %v", erri)
						continue
					}
				default:
					return nil, fmt.Errorf("unknown queue name: %s", jq.Name)
				}

				// Validate the job before checking the key
				if err := ValidateScheduledJob(job); err != nil {
					logger.Errorf("Job validation failed in Get: %v", err)
					continue
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
						logger.Errorf("Failed to unmarshal recipe job in Remove: %v", erri)
						continue
					}
				case IN_QUEUE:
					job, erri = unmarshalInStepJob(data)
					if erri != nil {
						logger.Errorf("Failed to unmarshal in step job in Remove: %v", erri)
						continue
					}
				case EVERY_QUEUE:
					job, erri = unmarshalEveryStepJob(data)
					if erri != nil {
						logger.Errorf("Failed to unmarshal every step job in Remove: %v", erri)
						continue
					}
				default:
					return nil, fmt.Errorf("unknown queue name: %s", jq.Name)
				}

				// Validate the job before checking the key
				if err := ValidateScheduledJob(job); err != nil {
					logger.Errorf("Job validation failed in Remove: %v", err)
					// Still try to remove the file if the key matches
					if jobKey.Name() == file.Name() || strings.TrimSuffix(file.Name(), ".json") == jobKey.Name() {
						remErr := os.Remove(path)
						if remErr != nil {
							return nil, remErr
						}
						return nil, fmt.Errorf("removed invalid job file: %s", path)
					}
					continue
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
