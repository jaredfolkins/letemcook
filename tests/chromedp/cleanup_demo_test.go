package main_test

import (
	"os"
	"testing"

	"github.com/jaredfolkins/letemcook/tests/testutil"
)

func TestCleanupDemo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Parallel()

	testutil.SimpleTestWrapper(t, func(t *testing.T, instance *testutil.TestInstance) {
		t.Logf("Test instance %s created with data dir: %s", instance.ID, instance.DataDir)

		// Verify the data directory exists during the test
		if _, err := os.Stat(instance.DataDir); os.IsNotExist(err) {
			t.Errorf("Expected data directory %s to exist during test", instance.DataDir)
		} else {
			t.Logf("✓ Data directory %s exists during test", instance.DataDir)
		}

		// Create a test file to verify cleanup
		testFile := instance.DataDir + "/test-was-here.txt"
		if err := os.WriteFile(testFile, []byte("test was here"), 0644); err != nil {
			t.Errorf("Failed to create test file: %v", err)
		} else {
			t.Logf("✓ Created test file %s", testFile)
		}

		// The cleanup function will remove this directory after the test
		t.Logf("Test completing - cleanup will happen automatically")
	})

	// Note: By the time we get here, cleanup has already happened
	// The defer in SimpleTestWrapper ensures cleanup runs before this point
}
