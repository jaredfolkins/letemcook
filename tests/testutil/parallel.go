package testutil

import (
	"crypto/rand"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/sqids/sqids-go"
)

// TestInstance represents a unique test instance with its own resources
type TestInstance struct {
	ID          string
	Port        int
	DataDir     string
	TestDataDir string
	BaseURL     string
	Cleanup     func(failed bool)
}

// generateTestID creates an 8-character lowercase alphanumeric ID
func generateTestID() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 8)
	rand.Read(b)
	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}
	return string(b)
}

// findAvailablePort finds an available port starting from the given base port
func findAvailablePort(basePort int) (int, error) {
	for port := basePort; port < basePort+1000; port++ {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			ln.Close()
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available port found starting from %d", basePort)
}

// CreateTestInstance creates a new test instance with unique resources
func CreateTestInstance() (*TestInstance, error) {
	id := generateTestID()

	// Find available port starting from 15362
	port, err := findAvailablePort(15362)
	if err != nil {
		return nil, fmt.Errorf("find available port: %w", err)
	}

	// Create unique data directory
	baseDataDir := DataRoot()
	dataDir := filepath.Join(baseDataDir, fmt.Sprintf("test-%s", id))
	testDataDir := filepath.Join(dataDir, "test")

	// Ensure clean test directory
	_ = os.RemoveAll(dataDir)
	if err := os.MkdirAll(testDataDir, 0o755); err != nil {
		return nil, fmt.Errorf("create test data dir: %w", err)
	}

	baseURL := fmt.Sprintf("http://localhost:%d", port)

	instance := &TestInstance{
		ID:          id,
		Port:        port,
		DataDir:     dataDir,
		TestDataDir: testDataDir,
		BaseURL:     baseURL,
	}

	return instance, nil
}

// StartTestServerForInstance starts a test server for a specific test instance
func (ti *TestInstance) StartTestServer() (func(failed bool), error) {
	repoRoot := RepoRoot()

	// Set up environment for this specific test instance
	env := []string{
		"LEMC_ENV=test",
		"LEMC_DATA=" + ti.DataDir,
		"LEMC_SQUID_ALPHABET=abcdefghijklmnopqrstuvwxyz0123456789",
		"LEMC_PORT_TEST=" + strconv.Itoa(ti.Port),
	}

	// Add existing environment, filtering out conflicting vars
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "LEMC_ENV=") &&
			!strings.HasPrefix(e, "LEMC_DATA=") &&
			!strings.HasPrefix(e, "LEMC_SQUID_ALPHABET=") &&
			!strings.HasPrefix(e, "LEMC_PORT_TEST=") {
			env = append(env, e)
		}
	}

	// Start the server
	serverCmd := exec.Command("go", "run", ".")
	serverCmd.Dir = repoRoot
	serverCmd.Env = env

	// Capture output for debugging
	serverCmd.Stdout = os.Stdout
	serverCmd.Stderr = os.Stderr

	if err := serverCmd.Start(); err != nil {
		return nil, fmt.Errorf("start server: %w", err)
	}

	// Wait for server to be ready
	if err := ti.waitForServerReady(30 * time.Second); err != nil {
		// Kill the server if it failed to start properly
		if serverCmd.Process != nil {
			serverCmd.Process.Kill()
		}
		return nil, fmt.Errorf("server not ready: %w", err)
	}

	// Return cleanup function that always cleans up resources
	cleanup := func(failed bool) {
		if serverCmd != nil && serverCmd.Process != nil {
			// Always try graceful shutdown first
			serverCmd.Process.Signal(syscall.SIGTERM)

			// Wait for shutdown
			done := make(chan error, 1)
			go func() {
				done <- serverCmd.Wait()
			}()

			select {
			case <-done:
				// Graceful shutdown succeeded
			case <-time.After(2 * time.Second):
				// Force kill if graceful shutdown failed
				serverCmd.Process.Kill()
				<-done
			}
		}

		// Always clean up resources after test completion
		if failed {
			fmt.Printf("Test failed, cleaning up resources for instance %s\n", ti.ID)
		} else {
			fmt.Printf("Test passed, cleaning up resources for instance %s\n", ti.ID)
		}
		_ = os.RemoveAll(ti.DataDir)
	}

	ti.Cleanup = cleanup
	return cleanup, nil
}

// waitForServerReady waits for the server to be ready to accept requests
func (ti *TestInstance) waitForServerReady(timeout time.Duration) error {
	client := &http.Client{Timeout: 1 * time.Second}
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		resp, err := client.Get(ti.BaseURL + "/")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode < 500 {
				return nil // Server is ready
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for server to be ready")
}

// GetTestInstanceBaseURL returns the base URL for the test instance
func (ti *TestInstance) GetTestInstanceBaseURL() string {
	return ti.BaseURL
}

// LoadTestEnvForInstance loads test environment for a specific instance
func (ti *TestInstance) LoadTestEnvForInstance() (alphaSquid, bravoSquid string, err error) {
	// The alphabet should already be set via environment variables when the server started
	// Use the same alphabet that was configured for the server
	alphabet := "abcdefghijklmnopqrstuvwxyz0123456789"

	// Create squid generator
	s, err := sqids.New(sqids.Options{
		Blocklist: nil,
		MinLength: 4,
		Alphabet:  alphabet,
	})
	if err != nil {
		return "", "", err
	}

	// Generate squids for account IDs 1 and 2 (from constants in chromedp.go)
	alphaSquid, err = s.Encode([]uint64{AlphaAccountID})
	if err != nil {
		return "", "", err
	}

	bravoSquid, err = s.Encode([]uint64{BravoAccountID})
	if err != nil {
		return "", "", err
	}

	return alphaSquid, bravoSquid, nil
}
