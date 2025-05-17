package testutil

import (
	"context"
	"fmt"
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
	testDataPath := filepath.Join(repoRoot, "data", "test")
	// Start with a clean test data directory
	os.RemoveAll(testDataPath)
	if err := os.MkdirAll(testDataPath, 0755); err != nil {
		return nil, fmt.Errorf("prepare test data dir: %w", err)
	}

	os.Setenv("LEMC_ENV", "test")
	os.Setenv("LEMC_DATA", filepath.Join(repoRoot, "data"))
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

	// allow server to start
	time.Sleep(2 * time.Second)

	shutdown := func() {
		cancel()
		_ = cmd.Process.Kill()
		cmd.Wait()
	}

	return shutdown, nil
}
