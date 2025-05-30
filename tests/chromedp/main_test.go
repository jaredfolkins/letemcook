package main_test

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/jaredfolkins/letemcook/tests/testutil"
)

func TestMain(m *testing.M) {
	// No longer need shared server setup - each test creates its own instance
	log.Println("Starting ChromeDP tests with parallel infrastructure")

	// Ensure we have a clean environment
	testutil.ForceCleanupChrome()

	// Run tests
	code := m.Run()

	// Final cleanup
	testutil.ForceCleanupChrome()

	if code != 0 {
		log.Printf("Tests failed with exit code %d", code)
	} else {
		log.Println("All ChromeDP tests completed successfully")
	}

	os.Exit(code)
}

func init() {
	fmt.Println("Initializing ChromeDP test package with parallel infrastructure")
}
