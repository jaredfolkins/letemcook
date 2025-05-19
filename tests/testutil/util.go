package testutil

import (
	"path/filepath"
	"runtime"
)

// RepoRoot returns the repository root directory path.
func RepoRoot() string {
	_, currentFile, _, _ := runtime.Caller(0)
	return filepath.Dir(filepath.Dir(filepath.Dir(currentFile)))
}

// DataRoot returns the data directory path used for tests.
func DataRoot() string {
	return filepath.Join(RepoRoot(), "data")
}
