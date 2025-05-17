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
	repoRoot := filepath.Dir(filepath.Dir(currentFile))

	os.Setenv("LEMC_ENV", "test")
	os.Setenv("LEMC_DATA", filepath.Join(repoRoot, "data"))
	if os.Getenv("LEMC_PORT_TEST") == "" {
		os.Setenv("LEMC_PORT_TEST", "15362")
	}
	os.MkdirAll(filepath.Join(os.Getenv("LEMC_DATA"), os.Getenv("LEMC_ENV")), 0755)

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
