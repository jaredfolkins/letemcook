package yeschef

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types/filters"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/jaredfolkins/letemcook/util"
)

var lemc_env_rgx = regexp.MustCompile(`^lemc\.env;`)
var lemc_css_buffer_rgx = regexp.MustCompile(`^lemc\.css\.buffer;`)
var lemc_css_trunc_rgx = regexp.MustCompile(`^lemc\.css\.trunc;`)
var lemc_css_append_rgx = regexp.MustCompile(`^lemc\.css\.append;`)
var lemc_html_buffer_rgx = regexp.MustCompile(`^lemc\.html\.buffer;`)
var lemc_html_trunc_rgx = regexp.MustCompile(`^lemc\.html\.trunc;`)
var lemc_html_append_rgx = regexp.MustCompile(`^lemc\.html\.append;`)
var lemc_timeout_rgx = regexp.MustCompile(`^(\d+)\.(\w+)$`)
var lemc_js_trunc_rgx = regexp.MustCompile(`^lemc\.js\.trunc;`)
var lemc_js_exec_rgx = regexp.MustCompile(`^lemc\.js\.exec;`)

func msg(message, imageHash, imageName string, job *JobRecipe, jm *util.JobMeta, cf *util.ContainerFiles, lf *util.LogFile) {
	r := &Response{
		UUID:     jm.UUID,
		PageID:   jm.PageID,
		ViewType: job.Scope,
	}

	log.Printf("msg: %s", message)

	lf.StepWriteToLog(jm.StepID, message, imageHash, imageName)

	if lemc_css_trunc_rgx.MatchString(message) {
		r.Cmd = LEMC_CSS_TRUNC
		r.Msg = strings.TrimPrefix(message, LEMC_CSS_TRUNC)
		err := cf.Trunc(r.Msg, cf.Css)
		if err != nil {
			log.Printf("Error truncating file css: %v", err)
			return
		}
	} else if lemc_css_append_rgx.MatchString(message) {
		r.Cmd = LEMC_CSS_APPEND
		r.Msg = strings.TrimPrefix(message, LEMC_CSS_APPEND)
		err := cf.Append(r.Msg, cf.Css)
		if err != nil {
			log.Printf("Error appending file css: %v", err)
			return
		}
	} else if lemc_css_buffer_rgx.MatchString(message) {
		r.Cmd = LEMC_CSS_BUFFER
		r.Msg = strings.TrimPrefix(message, LEMC_CSS_BUFFER)
		err := cf.Append(r.Msg, cf.Css)
		if err != nil {
			log.Printf("Error appending file css: %v", err)
			return
		}
	}

	if lemc_html_trunc_rgx.MatchString(message) {
		r.Cmd = LEMC_HTML_TRUNC
		r.Msg = strings.TrimPrefix(message, LEMC_HTML_TRUNC)
		err := cf.Trunc(r.Msg, cf.Html)
		if err != nil {
			log.Printf("Error truncating file html: %v", err)
			return
		}
	} else if lemc_html_append_rgx.MatchString(message) {
		r.Cmd = LEMC_HTML_APPEND
		r.Msg = strings.TrimPrefix(message, LEMC_HTML_APPEND)
		err := cf.Append(r.Msg, cf.Html)
		if err != nil {
			log.Printf("Error appending file html: %v", err)
			return
		}
	} else if lemc_html_buffer_rgx.MatchString(message) {
		r.Cmd = LEMC_HTML_BUFFER
		r.Msg = strings.TrimPrefix(message, LEMC_HTML_BUFFER)
		err := cf.Append(r.Msg, cf.Html)
		if err != nil {
			log.Printf("Error appending file html: %v", err)
			return
		}
	}

	if lemc_js_exec_rgx.MatchString(message) {
		r.Cmd = LEMC_JS_EXEC
		r.Msg = strings.TrimPrefix(message, LEMC_JS_EXEC)
		err := cf.Trunc(r.Msg, cf.Js)
		if err != nil {
			log.Printf("Error processing js exec: %v", err)
			return
		}
	} else if lemc_js_trunc_rgx.MatchString(message) {
		r.Cmd = LEMC_JS_TRUNC
		r.Msg = strings.TrimPrefix(message, LEMC_JS_TRUNC)
		err := cf.Trunc(r.Msg, cf.Js)
		if err != nil {
			log.Printf("Error truncating js file: %v", err)
			return
		}
	}

	if lemc_env_rgx.MatchString(message) {
		s := strings.TrimSpace(strings.TrimPrefix(message, LEMC_ENV))
		job.Env = append(job.Env, s)
		return
	}

	jsonData, err := json.Marshal(r)
	if err != nil {
		log.Printf("Error converting struct to JSON: %v", err)
		return
	}

	// Determine target users based on job scope
	var targetUserIDs []int64

	switch job.Scope {
	case "individual":
		// Send only to the job's owner
		individualUserID, err := strconv.ParseInt(job.UserID, 10, 64)
		if err != nil {
			log.Printf("Error parsing individual UserID '%s' for job %s (UUID: %s): %v", job.UserID, job.StepID, job.UUID, err)
			return // Cannot proceed without a valid user ID
		}
		targetUserIDs = []int64{individualUserID}
		log.Printf("[Job %s] Sending individual message to UserID %d for UUID %s", job.StepID, individualUserID, job.UUID)

	case "shared":
		// Use the pre-populated list for shared jobs
		targetUserIDs = job.RecipientUserIDs
		if len(targetUserIDs) > 0 {
			log.Printf("[Job %s] Broadcasting shared message to %d recipients for UUID %s", job.StepID, len(targetUserIDs), job.UUID)
		}

	default:
		log.Printf("Warning: Unknown job scope '%s' encountered for job %s (UUID: %s). Message not sent.", job.Scope, job.StepID, job.UUID)
		return // Don't send if scope is unknown
	}

	// Check if we have any valid targets
	if targetUserIDs == nil || len(targetUserIDs) == 0 {
		log.Printf("No valid recipient user IDs determined for StepID %s (UUID: %s, Scope: %s). Message not sent.", job.StepID, job.UUID, job.Scope)
		return
	}

	// Send message to the appropriate user(s)
	for _, userID := range targetUserIDs {
		targetServer := XoxoX.ReadInstance(userID)
		if targetServer != nil {
			select {
			case targetServer.Radio <- jsonData:
			default:
				log.Printf("Warning: Radio channel full for user %d, skipping message for job %s", userID, job.StepID)
			}
		}
	}
}

type Response struct {
	PageID   string
	UUID     string
	ViewType string
	Cmd      string
	Msg      string
}

func timeoutInSeconds(timeout string) (int, error) {
	m := lemc_timeout_rgx.FindStringSubmatch(timeout)
	if len(m) < 2 {
		return 0, fmt.Errorf("do.in regex matches total failed")
	}

	digit, err := strconv.Atoi(m[1])
	if err != nil {
		return 0, err
	}

	var multiplier int
	switch m[2] {
	case "second":
		multiplier = 1
	case "seconds":
		multiplier = 1
	case "minute":
		multiplier = 60
	case "minutes":
		multiplier = 60
	}

	return digit * multiplier, nil
}

func createDockerContainerTagMap(job *JobRecipe, adminOrUsername string) map[string]string {
	tagMap := make(map[string]string)
	tagMap["UUID"] = job.UUID
	tagMap["STEP_ID"] = job.StepID
	tagMap["CONTRACT_ID"] = job.PageID
	tagMap["USER_ID"] = job.UserID
	tagMap["USERNAME"] = adminOrUsername
	tagMap["RECIPE_NAME"] = job.Recipe.Name
	tagMap["OWNED_BY"] = OWNED_BY
	return tagMap
}

func deletePreviousContainer(ctx context.Context, cli *client.Client, job *JobRecipe, adminOrUsername string) error {

	mapRecipe := createDockerContainerTagMap(job, adminOrUsername)

	filterArgs := filters.NewArgs()
	for key, value := range mapRecipe {
		filterArgs.Add("label", fmt.Sprintf("%s=%s", key, value))
	}

	listOptions := container.ListOptions{
		All:     true,
		Filters: filterArgs,
	}

	containers, err := cli.ContainerList(ctx, listOptions)
	if err != nil {
		log.Printf("Error listing containers: %v", err)
		return err
	}

	removeOpts := container.RemoveOptions{
		RemoveVolumes: true,
		Force:         true,
		RemoveLinks:   false,
	}

	for _, c := range containers {
		err = cli.ContainerRemove(ctx, c.ID, removeOpts)
		if err != nil {
			log.Printf("Error removing container : %v", err)
			return err
		}
	}
	return nil
}

func NewHostConfig(cf *util.ContainerFiles, is_admin bool) *container.HostConfig {
	var mounts []mount.Mount

	mounts = append(mounts, mount.Mount{
		Type:   mount.TypeBind,
		Source: cf.BindPerUserPublicDir,
		Target: filepath.Join("/", "lemc", util.PUBLIC),
	})

	mounts = append(mounts, mount.Mount{
		Type:   mount.TypeBind,
		Source: cf.BindPerUserPrivateDir,
		Target: filepath.Join("/", "lemc", util.PRIVATE),
	})

	mounts = append(mounts, mount.Mount{
		Type:   mount.TypeBind,
		Source: cf.BindGlobalDir,
		Target: filepath.Join("/", "lemc", util.GLOBAL_DIR),
	})

	if is_admin {
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: cf.BindSharedDir,
			Target: "/lemc/shared",
		})
	}

	return &container.HostConfig{
		Mounts: mounts,
	}
}

func runContainer(server *CmdServer, job *JobRecipe, uri string) error {
	var err error
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(
		client.WithHost(os.Getenv("LEMC_DOCKER_HOST")),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return err
	}

	jm := util.NewJobMetaFromEnv(job.Env)

	cf, err := util.NewContainerFiles(jm, job.Recipe.IsShared)
	if err != nil {
		return err
	}

	fm, err := util.NewFileMeta(jm, job.Recipe.IsShared)
	if err != nil {
		return err
	}

	err = cf.OpenFiles()
	if err != nil {
		return err
	}
	defer cf.CloseFiles()

	lf, err := fm.OpenLogFile(jm)
	if err != nil {
		return err
	}
	defer lf.CloseLogFile()

	err = deletePreviousContainer(ctx, cli, job, fm.IndividualUsernameOrSharedUsername)
	if err != nil {
		return err
	}

	parsed, err := url.Parse(uri)
	if err != nil {
		return fmt.Errorf("failed to parse URI: %w", err)
	}

	image_name := strings.TrimPrefix(parsed.Path, "/")

	imageInspect, _, err := cli.ImageInspectWithRaw(ctx, image_name)
	if err != nil {
		log.Printf("runContainer Error: cli.ImageInspectWithRaw: %s", err)
		return err
	}

	trimmedHash := strings.TrimPrefix(imageInspect.ID, "sha256:")
	imageHash := trimmedHash[:8]

	hostCfg := NewHostConfig(cf, job.Recipe.IsShared)

	cfg := &container.Config{
		StopTimeout:  &job.ContainerTimeoutInSeconds,
		Image:        image_name,
		Env:          job.Env,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
		Labels:       createDockerContainerTagMap(job, fm.IndividualUsernameOrSharedUsername),
	}

	resp, err := cli.ContainerCreate(ctx, cfg, hostCfg, nil, nil, jm.GenerateContainerName(job.Recipe.Name, fm.IndividualUsernameOrSharedUsername))
	if err != nil {
		log.Printf("runContainer Error: cli.ConainterCreate: %s", err)
		return err
	}

	err = cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		log.Printf("runContainer Error: cli.ConainterStart: %s", err)
		return err
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		logCfg := container.LogsOptions{
			ShowStderr: true,
			ShowStdout: true,
			Follow:     true,
		}

		out, err := cli.ContainerLogs(ctx, resp.ID, logCfg)
		if err != nil {
			log.Println("ContainerLogError:", err)
			return
		}
		defer out.Close()

		reader := bufio.NewReader(out)
		for {
			header := make([]byte, 8)
			_, err := io.ReadFull(reader, header)
			if err != nil {
				if err != io.EOF {
					log.Println("error not EOF:", err)
				}
				return
			}

			count := binary.BigEndian.Uint32(header[4:])
			readBuf := make([]byte, count)
			_, err = io.ReadFull(reader, readBuf)
			if err != nil {
				if !errors.Is(err, io.EOF) {
					log.Println("Tried to read docker output: ", err)
				}
				return
			}

			s := strings.TrimSpace(string(readBuf))
			msg(s, imageHash, image_name, job, jm, cf, lf)
		}
	}()

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)

	doneTimeout := make(chan struct{})
	go timeoutCleanup(ctx, cli, job, resp, doneTimeout)

	for {
		select {
		case <-statusCh:
			close(doneTimeout)
			removeOpts := container.RemoveOptions{
				RemoveVolumes: true,
				RemoveLinks:   false,
				Force:         true,
			}

			err = cli.ContainerRemove(ctx, resp.ID, removeOpts)
			if err != nil {
				log.Println(err)
				return err
			}
			wg.Wait()
			return nil
		case err := <-errCh:
			close(doneTimeout)
			errx := deletePreviousContainer(ctx, cli, job, fm.IndividualUsernameOrSharedUsername)
			if errx != nil {
				log.Println(errx)
			}
			if err != nil {
				log.Println(err)
				return err
			}
			return nil
		}
	}
}

func timeoutCleanup(ctx context.Context, cli *client.Client, job *JobRecipe, resp container.CreateResponse, doneTimeout chan struct{}) {
	select {
	case <-doneTimeout:
	case <-time.After(time.Duration(job.ContainerTimeoutInSeconds) * time.Second):
		log.Println("Image Timeout exceeded")
		timeout := 10
		stopOpts := container.StopOptions{
			Timeout: &timeout,
		}

		err := cli.ContainerStop(ctx, resp.ID, stopOpts)
		if err != nil {
			log.Println(err)
		}
	}
}
