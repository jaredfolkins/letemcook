package util

import (
	"log"
	"path/filepath"
	"runtime"
)

// RepoRoot returns the repository root directory path.
func RepoRoot() string {
	_, currentFile, _, _ := runtime.Caller(0)
	// test_utils.go is in util/, so we need to go up 2 directories to get to project root
	root := filepath.Dir(filepath.Dir(currentFile))
	log.Printf("RepoRoot: %s", root)
	return root
}

// TestDataRoot returns the data directory path used for tests.
func TestDataRoot() string {
	return filepath.Join(RepoRoot(), "tests", "data")
}
