package yeschef

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"sync"
	"time"

	"github.com/hibiken/asynq"
	"github.com/reugn/go-quartz/quartz"
)

var lemc_do_now_rgx = regexp.MustCompile(`^now$`)
var lemc_do_in_rgx = regexp.MustCompile(`^in\.\d+\.\w+$`)
var lemc_do_every_rgx = regexp.MustCompile(`^every\.\d+\.\w+$`)
var XoxoX *ChefsKiss

type ErrHandler struct{}

func (er ErrHandler) HandleError(ctx context.Context, task *asynq.Task, err error) {
	log.Printf("XoxoX Error: %s\n", err.Error())
}

type RunningMan struct {
	mu   sync.Mutex
	list map[string]bool
}

func NewRunningMan() *RunningMan {
	return &RunningMan{
		list: make(map[string]bool),
	}
}

func (rm *RunningMan) Add(key string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.list[key] = true
}

func (rm *RunningMan) IsRunning(key string) bool {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	return rm.list[key]
}

func (rm *RunningMan) Remove(key string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	delete(rm.list, key)
}

func Start() {
	nowJq := NewQuartzQueue(NOW_QUEUE)
	nowScheduler := NewQuartzScheduler(nowJq)

	inJq := NewQuartzQueue(IN_QUEUE)
	inScheduler := NewQuartzScheduler(inJq)

	everyJq := NewQuartzQueue(EVERY_QUEUE)
	everyScheduler := NewQuartzScheduler(everyJq)

	XoxoX = &ChefsKiss{
		mu:             sync.RWMutex{},
		apps:           make(map[int64]*CmdServer),
		RunningMan:     NewRunningMan(),
		NowQueue:       nowJq,
		NowScheduler:   nowScheduler,
		InQueue:        inJq,
		InScheduler:    inScheduler,
		EveryQueue:     everyJq,
		EveryScheduler: everyScheduler,
	}
}

func NewQuartzScheduler(queue *jobQueue) *quartz.StdScheduler {
	config := quartz.StdSchedulerOptions{
		OutdatedThreshold: time.Hour * 48, // considering file system I/O latency
		WorkerLimit:       10,
	}
	scheduler := quartz.NewStdSchedulerWithOptions(config, queue, nil)
	ctx := context.Background()
	scheduler.Start(ctx)
	return scheduler
}

const jobGroupTemplt = "[userid:%s][page:%s][uuid:%s][group]"

func jobGroup(userid, pageid, uuid string) string {
	return fmt.Sprintf(jobGroupTemplt, userid, pageid, uuid)
}

const jobKeyAppSharedTemplt = "[app][shared][appid:%s][page:%s][uuid:%s][queue:%s]"
const jobKeyCookbookSharedTemplt = "[cookbook][shared][cookbookid:%s][page:%s][uuid:%s][queue:%s]"

const jobKeyAppIndividualTemplt = "[app][individual][userid:%s][page:%s][uuid:%s][queue:%s]"
const jobKeyCookbookIndividualTemplt = "[cookbook][individual][userid:%s][page:%s][uuid:%s][queue:%s]"

func LemcJobKey(recipe *JobRecipe, name string) string {
	if recipe == nil {
		return ""
	}

	switch recipe.Scope {
	case "shared":
		if len(recipe.AppID) > 0 {
			return fmt.Sprintf(jobKeyAppSharedTemplt, recipe.AppID, recipe.PageID, recipe.UUID, name)
		} else if len(recipe.CookbookID) > 0 {
			return fmt.Sprintf(jobKeyCookbookSharedTemplt, recipe.CookbookID, recipe.PageID, recipe.UUID, name)
		}
	case "individual":
		if len(recipe.AppID) > 0 {
			return fmt.Sprintf(jobKeyAppIndividualTemplt, recipe.UserID, recipe.PageID, recipe.UUID, name)
		} else if len(recipe.CookbookID) > 0 {
			return fmt.Sprintf(jobKeyCookbookIndividualTemplt, recipe.UserID, recipe.PageID, recipe.UUID, name)
		}
	}

	panices := fmt.Sprintf("LemcJobKey: recipe.Scope is not valid: %s", recipe.Scope)
	panic(panices)
	// old
	//return fmt.Sprintf(jobKeyAppIndividualTemplt, userid, pageid, uuid, name)
}

func BuildTags(env []string, do string) []string {
	tags := []string{}
	for _, envVar := range env {
		tags = append(tags, envVar)
	}

	tags = append(tags, fmt.Sprintf("DO=%s", do))
	return tags
}

func Ts(m string) (time.Duration, error) {
	var ts time.Duration
	if m == "seconds" || m == "second" {
		ts = time.Second
	} else if m == "minutes" || m == "minute" {
		ts = time.Minute
	} else if m == "hours" || m == "hour" {
		ts = time.Hour
	} else {
		return ts, fmt.Errorf("time signature not detected")
	}

	return ts, nil
}
