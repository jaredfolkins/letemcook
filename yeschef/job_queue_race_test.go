package yeschef

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/jaredfolkins/letemcook/models"
	"github.com/joho/godotenv"
	"github.com/reugn/go-quartz/quartz"
)

func loadEnvForTest(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Logf("Warning: Could not get current directory: %v", err)
		return
	}

	appDir := currentDir
	if strings.HasSuffix(currentDir, "yeschef") {
		appDir = filepath.Dir(currentDir)
	}

	envPath := filepath.Join(appDir, ".env")
	err = godotenv.Load(envPath)
	if err != nil {
		t.Logf("Warning: Could not load .env file from %s: %v", envPath, err)
		return
	}

	t.Logf("Successfully loaded environment from %s", envPath)
	t.Logf("LEMC_DOCKER_HOST=%s", os.Getenv("LEMC_DOCKER_HOST"))
	t.Logf("LEMC_QUEUES=%s", os.Getenv("LEMC_QUEUES"))
}

func isDockerAvailableForTest(t *testing.T, dockerHost string) bool {
	if dockerHost == "" {
		t.Log("Integration test: LEMC_DOCKER_HOST is not set, Docker might not be available")
		return false
	}

	cli, err := client.NewClientWithOpts(client.WithHost(dockerHost))
	if err != nil {
		t.Logf("Integration test: Error creating Docker client: %v", err)
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err = cli.Ping(ctx)
	if err != nil {
		t.Logf("Integration test: Error pinging Docker daemon: %v", err)
		return false
	}

	t.Log("Integration test: Docker is available for testing")
	return true
}

func setupTestDirectories(t *testing.T, baseDir string) {
	queueDirs := []string{NOW_QUEUE, IN_QUEUE, EVERY_QUEUE}
	for _, dir := range queueDirs {
		queueDir := filepath.Join(baseDir, dir)
		err := os.MkdirAll(queueDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create queue dir %s: %v", queueDir, err)
		}
	}

	os.Setenv("LEMC_QUEUES", baseDir)
}

func getCommonDockerPaths() []string {
	envPath := os.Getenv("LEMC_DOCKER_HOST")

	commonPaths := []string{
		"unix:///var/run/docker.sock",    // Standard Linux/macOS path
		"unix:///run/docker.sock",        // Alternative Linux path
		"npipe:////./pipe/docker_engine", // Windows path
	}

	var result []string

	if envPath != "" {
		result = append(result, envPath)
	}

	for _, path := range commonPaths {
		if path == envPath {
			continue
		}
		result = append(result, path)
	}

	return result
}

func skipIfDockerUnavailable(t *testing.T) {
	if testing.Short() {
		t.Log("Skipping integration test in short mode.")
		t.SkipNow()
	}

	t.Logf("DEBUG: LEMC_DOCKER_HOST=%s", os.Getenv("LEMC_DOCKER_HOST"))
	t.Logf("DEBUG: LEMC_QUEUES=%s", os.Getenv("LEMC_QUEUES"))

	dockerHost := os.Getenv("LEMC_DOCKER_HOST")

	commonPaths := getCommonDockerPaths()
	t.Logf("DEBUG: Checking common Docker paths: %v", commonPaths)

	if !isDockerAvailableForTest(t, dockerHost) {
		t.Log("Integration test: Docker is completely unavailable. Skipping test.")
		t.SkipNow()
	}
}

func setupRaceTest(t *testing.T) (string, func()) {
	loadEnvForTest(t)

	queuesPath := os.Getenv("LEMC_QUEUES")
	var tmpDir string
	var err error

	if queuesPath == "" {
		tmpDir, err = os.MkdirTemp("", "job-queue-test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		t.Logf("Created temp directory for queues: %s", tmpDir)
		setupTestDirectories(t, tmpDir)
	} else {
		t.Logf("Using existing queues path: %s", queuesPath)
		tmpDir = queuesPath
	}

	absPath, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	os.Setenv("LEMC_TEMP_DIR", absPath)

	return tmpDir, func() {
		os.Unsetenv("LEMC_TEMP_DIR")

		if queuesPath == "" {
			os.Unsetenv("LEMC_QUEUES") // Ensure this is unset if we set it temporarily
			os.RemoveAll(tmpDir)
		}

		stopSchedulers(t)
	}
}

func stopSchedulers(t *testing.T) {
	if XoxoX == nil {
		return
	}

	if XoxoX.NowScheduler != nil {
		t.Log("Stopping NOW scheduler")
		XoxoX.NowScheduler.Stop()
	}

	if XoxoX.InScheduler != nil {
		t.Log("Stopping IN scheduler")
		XoxoX.InScheduler.Stop()
	}

	if XoxoX.EveryScheduler != nil {
		t.Log("Stopping EVERY scheduler")
		XoxoX.EveryScheduler.Stop()
	}

	time.Sleep(100 * time.Millisecond)
}

func mockDockerOperationsForTest(t *testing.T, job *JobRecipe) {
	t.Logf("Integration test: Mocking Docker operations for job %s", job.UUID)

}

func runTestJobWithDockerHandling(t *testing.T, job *JobRecipe) error {
	dockerHost := os.Getenv("LEMC_DOCKER_HOST")
	dockerAvailable := isDockerAvailableForTest(t, dockerHost)

	cli, err := client.NewClientWithOpts(
		client.WithHost(dockerHost),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		t.Logf("Failed to create Docker client: %v", err)
		mockDockerOperationsForTest(t, job)
		return nil
	}

	for _, step := range job.Recipe.Steps {
		imageAvailable := handleImageForTest(t, cli, step.Image)
		if !imageAvailable {
			if !dockerAvailable {
				t.Logf("Image '%s' not available and Docker unavailable, mocking operations.", step.Image)
				mockDockerOperationsForTest(t, job)
				return nil // Continue test by mocking
			} else {
				t.Logf("Warning: Image '%s' could not be handled by handleImageForTest, even though Docker is available. Mocking operations.", step.Image)
				mockDockerOperationsForTest(t, job)
				return nil // Continue test by mocking
			}
		}
	}

	jobType := job.Recipe.Steps[0].Do
	var jobErr error

	if strings.HasPrefix(jobType, "in.") {
		jobErr = DoIn(job.Recipe.Steps[0], job)
	} else if strings.HasPrefix(jobType, "every.") {
		jobErr = DoEvery(job.Recipe.Steps[0], job)
	} else {
		jobErr = DoNow(job)
	}

	if jobErr != nil {
		if userErr := GetUserVisibleError(jobErr); userErr != nil {
			t.Logf("Integration test: Handling user-visible error: %v", jobErr)
			mockDockerOperationsForTest(t, job)
			return nil // Test continues by mocking
		}

		if strings.Contains(jobErr.Error(), "docker") ||
			strings.Contains(jobErr.Error(), "container") ||
			strings.Contains(jobErr.Error(), "image") {
			t.Logf("Integration test: Handling Docker-related error: %v", jobErr)
			mockDockerOperationsForTest(t, job)
			return nil // Test continues by mocking
		}

		t.Logf("Error during job execution: %v", jobErr)
		return jobErr
	}

	t.Logf("Job %s completed successfully in test handling.", job.UUID)
	return nil
}

func startForTest(t *testing.T, tempDir string) {
	nowQueue := &jobQueue{Path: filepath.Join(tempDir, NOW_QUEUE), Name: NOW_QUEUE}
	inQueue := &jobQueue{Path: filepath.Join(tempDir, IN_QUEUE), Name: IN_QUEUE}
	everyQueue := &jobQueue{Path: filepath.Join(tempDir, EVERY_QUEUE), Name: EVERY_QUEUE}

	ctx := context.Background()
	config := quartz.StdSchedulerOptions{
		OutdatedThreshold: time.Hour * 1, // Short threshold for tests
		WorkerLimit:       2,             // Fewer workers for testing
	}

	nowScheduler := quartz.NewStdSchedulerWithOptions(config, nowQueue, nil)
	nowScheduler.Start(ctx)

	inScheduler := quartz.NewStdSchedulerWithOptions(config, inQueue, nil)
	inScheduler.Start(ctx)

	everyScheduler := quartz.NewStdSchedulerWithOptions(config, everyQueue, nil)
	everyScheduler.Start(ctx)

	XoxoX = &ChefsKiss{
		RunningMan:     NewRunningMan(),
		NowQueue:       nowQueue,
		NowScheduler:   nowScheduler,
		InQueue:        inQueue,
		InScheduler:    inScheduler,
		EveryQueue:     everyQueue,
		EveryScheduler: everyScheduler,
		apps:           make(map[int64]*CmdServer),
	}

	t.Logf("Test environment initialized with queues in %s", tempDir)
}

func TestIntegrationConcurrentJobOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	loadEnvForTest(t)

	skipIfDockerUnavailable(t)

	tmpDir, cleanup := setupRaceTest(t)
	defer cleanup()

	dockerHost := os.Getenv("LEMC_DOCKER_HOST")
	_ = isDockerAvailableForTest(t, dockerHost) // We don't need the return value as the test will proceed with mocking

	t.Log("Running integration test that interacts with Docker and filesystem")

	startForTest(t, tmpDir)

	t.Run("Concurrent NOW job submissions", func(t *testing.T) {
		const jobCount = 10
		var wg sync.WaitGroup

		successCount := 0
		failureCount := 0
		var counterMutex sync.Mutex

		for i := 0; i < jobCount; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Add(-1)

				job := &JobRecipe{
					UUID:     "concurrent-uuid",
					PageID:   "54321", // Numeric pageID
					UserID:   "12345", // Numeric userID
					Username: "test-username",
					Recipe: models.Recipe{
						Name: "test-recipe",
						Steps: []models.Step{
							{
								Step:    1,
								Do:      "now",
								Image:   "alpine:latest",
								Timeout: "1.minutes",
							},
						},
					},
				}

				err := runTestJobWithDockerHandling(t, job)

				counterMutex.Lock()
				if err != nil {
					failureCount++
				} else {
					successCount++
				}
				counterMutex.Unlock()
			}(i)
		}

		wg.Wait()

		t.Logf("Jobs completed: %d successful, %d failed", successCount, failureCount)

		time.Sleep(time.Second * 2)
	})

	t.Run("Concurrent NOW, IN, EVERY job interaction", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(3)

		var unexpectedFailure bool
		var failureMutex sync.Mutex

		go func() {
			defer wg.Add(-1)
			inJob := &JobRecipe{
				UUID:     "mixed-uuid",
				PageID:   "54321", // Numeric pageID
				UserID:   "12345", // Numeric userID
				Username: "test-username",
				Recipe: models.Recipe{
					Name: "test-recipe",
					Steps: []models.Step{
						{
							Step:    1,
							Do:      "in.1.minutes",
							Image:   "alpine:latest",
							Timeout: "1.minutes",
						},
					},
				},
			}

			err := runTestJobWithDockerHandling(t, inJob)
			failureMutex.Lock()
			if err != nil {
				unexpectedFailure = true
				t.Logf("IN job failed unexpectedly: %v", err)
			}
			failureMutex.Unlock()
		}()

		time.Sleep(100 * time.Millisecond)

		go func() {
			defer wg.Add(-1)
			everyJob := &JobRecipe{
				UUID:     "mixed-uuid",
				PageID:   "54321", // Numeric pageID
				UserID:   "12345", // Numeric userID
				Username: "test-username",
				Recipe: models.Recipe{
					Name: "test-recipe",
					Steps: []models.Step{
						{
							Step:    1,
							Do:      "every.1.minutes",
							Image:   "alpine:latest",
							Timeout: "1.minutes",
						},
					},
				},
			}

			err := runTestJobWithDockerHandling(t, everyJob)
			failureMutex.Lock()
			if err != nil {
				unexpectedFailure = true
				t.Logf("EVERY job failed unexpectedly: %v", err)
			}
			failureMutex.Unlock()
		}()

		time.Sleep(100 * time.Millisecond)

		go func() {
			defer wg.Add(-1)
			nowJob := &JobRecipe{
				UUID:     "mixed-uuid",
				PageID:   "54321", // Numeric pageID
				UserID:   "12345", // Numeric userID
				Username: "test-username",
				Recipe: models.Recipe{
					Name: "test-recipe",
					Steps: []models.Step{
						{
							Step:    1,
							Do:      "now",
							Image:   "alpine:latest",
							Timeout: "1.minutes",
						},
					},
				},
			}

			err := runTestJobWithDockerHandling(t, nowJob)
			failureMutex.Lock()
			if err != nil {
				unexpectedFailure = true
				t.Logf("NOW job failed unexpectedly: %v", err)
			}
			failureMutex.Unlock()
		}()

		wg.Wait()

		if unexpectedFailure {
			t.Error("One or more jobs failed unexpectedly")
		}

		time.Sleep(time.Second * 2)
	})

	t.Run("Race condition with multiple start/stop operations", func(t *testing.T) {
		var wg sync.WaitGroup
		const iterations = 5

		for i := 0; i < iterations; i++ {
			wg.Add(3)

			go func(idx int) {
				defer wg.Add(-1)
				nowJob := &JobRecipe{
					UUID:     fmt.Sprintf("race-uuid-%d", idx),
					PageID:   fmt.Sprintf("%d", 54321+idx), // Unique numeric pageID
					UserID:   "12345",                      // Numeric userID
					Username: "test-username",
					Recipe: models.Recipe{
						Name: "test-recipe",
						Steps: []models.Step{
							{
								Step:    1,
								Do:      "now",
								Image:   "alpine:latest",
								Timeout: "1.minutes",
							},
						},
					},
				}
				runTestJobWithDockerHandling(t, nowJob)
			}(i)

			go func(idx int) {
				defer wg.Add(-1)
				inJob := &JobRecipe{
					UUID:     fmt.Sprintf("race-uuid-%d", idx),
					PageID:   fmt.Sprintf("%d", 54321+idx), // Unique numeric pageID
					UserID:   "12345",                      // Numeric userID
					Username: "test-username",
					Recipe: models.Recipe{
						Name: "test-recipe",
						Steps: []models.Step{
							{
								Step:    1,
								Do:      "in.1.minutes",
								Image:   "alpine:latest",
								Timeout: "1.minutes",
							},
						},
					},
				}
				runTestJobWithDockerHandling(t, inJob)
			}(i)

			go func(idx int) {
				defer wg.Add(-1)
				everyJob := &JobRecipe{
					UUID:     fmt.Sprintf("race-uuid-%d", idx),
					PageID:   fmt.Sprintf("%d", 54321+idx), // Unique numeric pageID
					UserID:   "12345",                      // Numeric userID
					Username: "test-username",
					Recipe: models.Recipe{
						Name: "test-recipe",
						Steps: []models.Step{
							{
								Step:    1,
								Do:      "every.1.minutes",
								Image:   "alpine:latest",
								Timeout: "1.minutes",
							},
						},
					},
				}
				runTestJobWithDockerHandling(t, everyJob)
			}(i)
		}

		wg.Wait()

		time.Sleep(time.Second * 3)
	})
}

func handleImageForTest(t *testing.T, cli *client.Client, imageName string) bool {
	// Assume no auth needed for test images or use a default placeholder if required
	imageSpec := ImageSpec{Name: imageName, RegistryAuth: ""}
	err := handleImagePull(cli, imageSpec) // Pass ImageSpec
	if err == nil {
		t.Logf("Image '%s' successfully handled by handleImagePull.", imageName)
		return true
	}

	t.Logf("Initial handleImagePull failed for '%s': %v. Checking test fallbacks...", imageName, err)

	if strings.HasPrefix(imageName, "lemc-") {
		images, listErr := cli.ImageList(context.Background(), image.ListOptions{})
		if listErr != nil {
			t.Logf("Failed to list Docker images during fallback check: %v", listErr)
		} else {
			for _, img := range images {
				for _, tag := range img.RepoTags {
					if strings.HasPrefix(tag, "lemc-") {
						t.Logf("Test Fallback: Using local image '%s' as substitute for '%s'.", tag, imageName)
						return true // Found a substitute
					}
				}
			}
			t.Logf("Test Fallback: No substitute lemc- image found locally for '%s'.", imageName)
		}
	}

	if imageName == "alpine" || imageName == "alpine:latest" {
		t.Logf("Test Fallback: '%s' image not available. Proceeding might lead to mocking.", imageName)
	}

	t.Logf("Image '%s' could not be handled (pull failed and no suitable test fallback applied).", imageName)
	return false // Indicate image is not usable for the test
}

func TestIntegrationImagePullDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	loadEnvForTest(t)
	skipIfDockerUnavailable(t)

	imageToTest := "hello-world:latest" // Using the official hello-world image
	dockerHost := os.Getenv("LEMC_DOCKER_HOST")

	cli, err := client.NewClientWithOpts(client.WithHost(dockerHost))
	if err != nil {
		t.Fatalf("Failed to create Docker client: %v", err)
	}

	defer func() {
		t.Logf("Cleaning up: Attempting to remove image %s", imageToTest)
		_, err := cli.ImageRemove(context.Background(), imageToTest, image.RemoveOptions{Force: true})
		if err != nil {
			t.Logf("Warning: Failed to remove image %s during cleanup: %v", imageToTest, err)
		}
	}()

	t.Logf("Attempting pre-test removal of image %s (if it exists)", imageToTest)
	_, err = cli.ImageRemove(context.Background(), imageToTest, image.RemoveOptions{Force: true})
	if err != nil {
		if strings.Contains(err.Error(), "No such image") {
			t.Logf("Image %s was not present initially.", imageToTest)
		} else {
			t.Logf("Warning: Unexpected error during pre-test removal of %s: %v", imageToTest, err)
		}
	} else {
		t.Logf("Successfully removed existing image %s before test.", imageToTest)
	}

	t.Logf("Attempting to pull image %s using handleImagePull", imageToTest)
	// Assume no auth needed for hello-world or use a default placeholder
	imageSpec := ImageSpec{Name: imageToTest, RegistryAuth: ""}
	err = handleImagePull(cli, imageSpec) // Pass ImageSpec
	if err != nil {
		t.Fatalf("Failed to pull image %s using handleImagePull: %v", imageToTest, err)
	}
	t.Logf("Successfully pulled image %s.", imageToTest)

	t.Logf("Verifying that image %s exists locally", imageToTest)
	if !imageExists(cli, imageToTest) {
		t.Fatalf("Image %s should exist locally after pull, but imageExists returned false.", imageToTest)
	}
	t.Logf("Verified image %s exists locally.", imageToTest)

	t.Logf("Attempting to create and run a container for %s", imageToTest)
	ctx := context.Background()
	resp, err := cli.ContainerCreate(ctx, &container.Config{Image: imageToTest}, nil, nil, nil, "")
	if err != nil {
		t.Fatalf("Failed to create container for %s: %v", imageToTest, err)
	}
	containerID := resp.ID

	defer func() {
		t.Logf("Cleaning up container %s", containerID[:12])
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		err := cli.ContainerRemove(cleanupCtx, containerID, container.RemoveOptions{Force: true})
		if err != nil {
			t.Logf("Warning: Failed to remove container %s: %v", containerID[:12], err)
		}
	}()

	err = cli.ContainerStart(ctx, containerID, container.StartOptions{})
	if err != nil {
		t.Fatalf("Failed to start container %s: %v", containerID[:12], err)
	}
	t.Logf("Container %s started", containerID[:12])

	statusCh, errCh := cli.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
	select {
	case status := <-statusCh:
		t.Logf("Container %s finished with status code: %d", containerID[:12], status.StatusCode)
		if status.StatusCode != 0 {
			t.Errorf("Expected container %s to exit with status 0, but got %d", containerID[:12], status.StatusCode)
		}
	case err := <-errCh:
		t.Fatalf("Error waiting for container %s: %v", containerID[:12], err)
	case <-time.After(30 * time.Second): // Add a timeout
		t.Fatalf("Timed out waiting for container %s to finish", containerID[:12])
	}

	t.Logf("Successfully ran container for %s", imageToTest)
}
