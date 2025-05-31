package tests

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	// No longer need shared server setup - each test creates its own instance
	log.Println("Starting integration tests with parallel infrastructure")

	// Run tests
	code := m.Run()

	/*
		if code != 0 {
			log.Printf("Integration tests failed with exit code %d", code)
		} else {
			log.Println("All integration tests completed successfully")
		}
	*/

	// Mark test package as complete to trigger final cleanup
	MarkTestPackageComplete()

	// Give cleanup time to finish
	time.Sleep(300 * time.Millisecond)

	// More targeted cleanup of test-related processes only
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		// First, clean up our registered test processes (safest approach)
		CleanupRegisteredProcesses()

		// Then clean up any remaining test-specific processes as backup
		cleanupTestChrome()
		cleanupTestGoProcesses()
		cleanupTestLemcProcesses()
	}

	// Final cleanup delay
	time.Sleep(500 * time.Millisecond)

	// Force garbage collection
	runtime.GC()
	runtime.GC()

	log.Printf("Integration test cleanup completed: %d", code)
	os.Exit(code)
}

// cleanupTestChrome only kills Chrome processes that are clearly from chromedp tests
func cleanupTestChrome() {
	// Look for Chrome processes with chromedp-runner temp directories
	if runtime.GOOS == "darwin" {
		// Kill Chrome processes with chromedp-runner user-data-dir (very specific to our tests)
		exec.Command("pkill", "-f", "chromedp-runner").Run()
		// Kill headless Chrome with our specific temp dir pattern
		exec.Command("pkill", "-f", "--user-data-dir=.*chromedp-runner").Run()
		// Clean up any Chrome processes with remote-debugging-port=0 (chromedp specific)
		exec.Command("pkill", "-f", "--remote-debugging-port=0").Run()
	} else if runtime.GOOS == "linux" {
		// Similar patterns for Linux
		exec.Command("pkill", "-f", "chromedp-runner").Run()
		exec.Command("pkill", "-f", "--user-data-dir=.*chromedp-runner").Run()
		exec.Command("pkill", "-f", "--remote-debugging-port=0").Run()
	}
}

// cleanupTestGoProcesses only kills go processes from our test directory
func cleanupTestGoProcesses() {
	// Get our repo path to be more specific
	wd, err := os.Getwd()
	if err == nil {
		// Only kill go run processes in our specific test directory path
		repoPath := strings.TrimSuffix(wd, "/tests")
		if strings.Contains(repoPath, "letemcook") {
			exec.Command("pkill", "-f", "go run.*"+repoPath).Run()
		}
	}
}

// cleanupTestLemcProcesses only kills letemcook processes on test ports
func cleanupTestLemcProcesses() {
	// Kill processes listening on our test port ranges (15362+)
	if runtime.GOOS == "darwin" {
		// Use lsof to find processes on test ports and kill them
		cmd := exec.Command("sh", "-c", "lsof -ti:15362-15500 2>/dev/null | xargs kill -9 2>/dev/null || true")
		cmd.Run()
	} else if runtime.GOOS == "linux" {
		// Similar approach for Linux
		cmd := exec.Command("sh", "-c", "fuser -k 15362-15500/tcp 2>/dev/null || true")
		cmd.Run()
	}
}

func init() {
	fmt.Println("Initializing integration test package with parallel infrastructure")
}
