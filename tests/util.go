package tests

import (
	"log"
	"path/filepath"
	"runtime"
)

// RepoRoot returns the repository root directory path.
func RepoRoot() string {
	_, currentFile, _, _ := runtime.Caller(0)
	root := filepath.Dir(filepath.Dir(currentFile))
	log.Printf("RepoRoot: %s", root)
	return root
}

// DataRoot returns the data directory path used for tests.
func DataRoot() string {
	return filepath.Join(RepoRoot(), "data")
}
