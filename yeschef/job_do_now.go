package yeschef

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/jaredfolkins/letemcook/util"
	"github.com/reugn/go-quartz/quartz"
)

func PerRecipeDownloadAllContainers() {

}

func imageExists(cli *client.Client, imageRef string) bool {
	normalizedName, _, _, err := util.NormalizeImageName(imageRef)
	if err != nil {
		log.Printf("Error parsing image name '%s' in imageExists: %v", imageRef, err)
		return false // Cannot check existence if parsing fails
	}

	images, err := cli.ImageList(context.Background(), image.ListOptions{})
	if err != nil {
		log.Printf("Error listing images in imageExists: %v", err)
		return false // Cannot determine existence if listing fails
	}

	for _, img := range images {
		for _, repoTag := range img.RepoTags {
			existingNormalized, _, _, parseErr := util.NormalizeImageName(repoTag)
			if parseErr == nil && existingNormalized == normalizedName {
				log.Printf("Image exists: Found matching tag '%s' for reference '%s'", repoTag, normalizedName)
				return true
			}
		}
	}

	log.Printf("Image does not exist locally: %s", normalizedName)
	return false
}

func PerRecipeDeleteRunningContainersByStep(cli *client.Client, job *JobRecipe, step_id string) error {

	if len(step_id) == 0 {
		return fmt.Errorf("error: missing step_id")
	}

	err := deletePreviousContainer(context.Background(), cli, job, job.Recipe.UsernameOrAdmin())
	if err != nil {
		return err
	}
	return nil
}

func PerRecipeDeleteAnyExistingJobs(jr *JobRecipe) {
	nowkey := LemcJobKey(jr, NOW_QUEUE)
	inkey := LemcJobKey(jr, IN_QUEUE)
	everykey := LemcJobKey(jr, EVERY_QUEUE)

	XoxoX.NowScheduler.DeleteJob(quartz.NewJobKey(nowkey))
	XoxoX.NowQueue.Remove(quartz.NewJobKey(nowkey))

	XoxoX.InScheduler.DeleteJob(quartz.NewJobKey(inkey))
	XoxoX.InQueue.Remove(quartz.NewJobKey(inkey))

	XoxoX.EveryScheduler.DeleteJob(quartz.NewJobKey(everykey))
	XoxoX.EveryQueue.Remove(quartz.NewJobKey(everykey))
}

func DoNow(jr *JobRecipe) error {
	cli, err := client.NewClientWithOpts(
		client.WithHost(os.Getenv("LEMC_DOCKER_HOST")),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err = cli.Ping(ctx)
	if err != nil {
		return fmt.Errorf("docker daemon is not accessible: %w", err)
	}

	var missingImages []string

	for _, st := range jr.Recipe.Steps {
		if !imageExists(cli, st.Image) {
			log.Printf("Image %s not found locally, attempting to pull", st.Image)
			err := handleImagePull(cli, st.Image)
			if err != nil {
				log.Printf("Error pulling image %s: %v", st.Image, err)
				missingImages = append(missingImages, st.Image)
			}
		}
	}

	if len(missingImages) == 0 {
		for _, st := range jr.Recipe.Steps {
			if !imageExists(cli, st.Image) {
				missingImages = append(missingImages, st.Image)
			}
		}
	}

	if len(missingImages) > 0 {
		errorMsg := fmt.Sprintf("Cannot execute job: Missing Docker images: %s", strings.Join(missingImages, ", "))

		log.Println(errorMsg)

		return NewUserVisibleError("MISSING_IMAGES", errorMsg, map[string]interface{}{
			"images": missingImages,
		})
	}

	key := LemcJobKey(jr, NOW_QUEUE)
	kg := quartz.NewJobKeyWithGroup(key, jobGroup(jr.UserID, jr.PageID, jr.UUID))
	jdo := &quartz.JobDetailOptions{
		Replace: true,
	}

	XoxoX.RunningMan.mu.Lock()
	if XoxoX.RunningMan.list[key] {
		XoxoX.RunningMan.mu.Unlock()
		return fmt.Errorf("error: a recipe is already running")
	}

	XoxoX.RunningMan.list[key] = true
	XoxoX.RunningMan.mu.Unlock()

	PerRecipeDeleteAnyExistingJobs(jr)

	for _, st := range jr.Recipe.Steps {
		err := PerRecipeDeleteRunningContainersByStep(cli, jr, fmt.Sprintf("%d", st.Step))
		if err != nil {
			log.Printf("Warning: Failed to clean up existing containers for step %d: %v", st.Step, err)
		}
	}

	detail := quartz.NewJobDetailWithOptions(jr, kg, jdo)
	if err := XoxoX.NowScheduler.ScheduleJob(detail, quartz.NewRunOnceTrigger(time.Millisecond*100)); err != nil {
		XoxoX.RunningMan.mu.Lock()
		delete(XoxoX.RunningMan.list, key)
		XoxoX.RunningMan.mu.Unlock()
		return fmt.Errorf("failed to schedule job: %w", err)
	}

	return nil
}
