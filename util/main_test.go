package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	dataRoot := TestDataRoot()
	os.Setenv("LEMC_ENV", "test")
	os.Setenv("LEMC_DATA", dataRoot)
	envDir := filepath.Join(dataRoot, "test")
	_ = os.MkdirAll(envDir, 0o755)
	code := m.Run()
	os.Exit(code)
}
