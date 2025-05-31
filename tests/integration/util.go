package tests

import (
	"log"
	"path/filepath"
	"runtime"
)

// RepoRoot returns the repository root directory path.
func RepoRoot() string {
	_, currentFile, _, _ := runtime.Caller(0)
	// util.go is in tests/integration/, so we need to go up 3 directories to get to project root
	root := filepath.Dir(filepath.Dir(filepath.Dir(currentFile)))
	log.Printf("RepoRoot: %s", root)
	return root
}

// DataRoot returns the data directory path used for tests.
func DataRoot() string {
	return filepath.Join(RepoRoot(), "tests", "data")
}
