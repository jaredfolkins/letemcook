package testutil

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

var (
	serverCmd *exec.Cmd
	serverMu  sync.Mutex
)

// StartTestServer starts the LEMC server for testing and returns a shutdown function.
func StartTestServer() (func(), error) {
	serverMu.Lock()
	defer serverMu.Unlock()

	repoRoot := RepoRoot()
	dataRoot := DataRoot()
	testDataPath := filepath.Join(dataRoot, "test")

	// Ensure a completely clean test directory - remove everything
	_ = os.RemoveAll(testDataPath)
	if err := os.MkdirAll(testDataPath, 0o755); err != nil {
		return nil, fmt.Errorf("prepare test data dir: %w", err)
	}

	// Clean environment setup for each test run
	os.Setenv("LEMC_ENV", "test")
	os.Setenv("LEMC_DATA", dataRoot)
	os.Setenv("LEMC_SQUID_ALPHABET", "abcdefghijklmnopqrstuvwxyz0123456789")
	if os.Getenv("LEMC_PORT_TEST") == "" {
		os.Setenv("LEMC_PORT_TEST", "15362")
	}

	// Start the server
	serverCmd = exec.Command("go", "run", ".")
	serverCmd.Dir = repoRoot
	serverCmd.Env = append(os.Environ(),
		"LEMC_ENV=test",
		"LEMC_DATA="+dataRoot,
		"LEMC_SQUID_ALPHABET=abcdefghijklmnopqrstuvwxyz0123456789",
		"LEMC_PORT_TEST=15362",
	)

	// Capture output for debugging
	serverCmd.Stdout = os.Stdout
	serverCmd.Stderr = os.Stderr

	// Create a context with timeout for server startup
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := serverCmd.Start(); err != nil {
		return nil, fmt.Errorf("start server: %w", err)
	}

	// Wait for server to be ready
	if err := WaitForServerReady(ctx); err != nil {
		// Kill the server if it failed to start properly
		if serverCmd.Process != nil {
			serverCmd.Process.Kill()
		}
		return nil, fmt.Errorf("server not ready: %w", err)
	}

	// Return cleanup function
	return func() {
		serverMu.Lock()
		defer serverMu.Unlock()

		if serverCmd != nil && serverCmd.Process != nil {
			// Try graceful shutdown first
			serverCmd.Process.Signal(os.Interrupt)

			// Wait a moment for graceful shutdown
			done := make(chan error, 1)
			go func() {
				done <- serverCmd.Wait()
			}()

			select {
			case <-done:
				// Graceful shutdown succeeded
			case <-time.After(3 * time.Second):
				// Force kill if graceful shutdown takes too long
				serverCmd.Process.Kill()
				<-done // Wait for the process to actually exit
			}

			serverCmd = nil
		}

		// Clean up test data
		_ = os.RemoveAll(testDataPath)
	}, nil
}

// WaitForServerReady waits for the server to be ready to accept requests
func WaitForServerReady(ctx context.Context) error {
	baseURL := GetBaseURL()
	client := &http.Client{Timeout: 1 * time.Second}

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for server to be ready")
		case <-ticker.C:
			resp, err := client.Get(baseURL + "/")
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode < 500 {
					return nil // Server is ready
				}
			}
		}
	}
}
