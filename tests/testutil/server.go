package testutil

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// StartTestServer starts the LEMC server for testing and returns a shutdown function.
func StartTestServer() (func(), error) {
	// Determine repository root based on this file location
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("failed to determine caller")
	}
	// Repo root is three directories up from this file: tests/testutil/server.go
	repoRoot := filepath.Dir(filepath.Dir(filepath.Dir(currentFile)))

	tempRoot, err := os.MkdirTemp("", "lemc_testdata_")
	if err != nil {
		return nil, fmt.Errorf("create temp root: %w", err)
	}
	testDataPath := filepath.Join(tempRoot, "test")
	if err := os.MkdirAll(testDataPath, 0755); err != nil {
		return nil, fmt.Errorf("prepare test data dir: %w", err)
	}

	os.Setenv("LEMC_ENV", "test")
	os.Setenv("LEMC_DATA", tempRoot)
	if os.Getenv("LEMC_PORT_TEST") == "" {
		os.Setenv("LEMC_PORT_TEST", "15362")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "go", "run", "main.go")
	cmd.Dir = repoRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("start server: %w", err)
	}

	if err := waitForServerReady(os.Getenv("LEMC_PORT_TEST"), 10*time.Second); err != nil {
		cancel()
		_ = cmd.Process.Kill()
		cmd.Wait()
		return nil, err
	}

	shutdown := func() {
		cancel()
		_ = cmd.Process.Kill()
		cmd.Wait()
		os.RemoveAll(tempRoot)
	}

	return shutdown, nil
}

// waitForServerReady polls the server until it responds or the timeout is reached.
func waitForServerReady(port string, timeout time.Duration) error {
	baseURL := "http://localhost:" + port
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(baseURL + "/")
		if err == nil {
			resp.Body.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("server not ready after %v", timeout)
}
