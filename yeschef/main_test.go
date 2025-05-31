package yeschef

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jaredfolkins/letemcook/util"
)

func TestMain(m *testing.M) {
	dataRoot := util.TestDataRoot()
	os.Setenv("LEMC_ENV", "test")
	os.Setenv("LEMC_DATA", dataRoot)
	envDir := filepath.Join(dataRoot, "test")
	_ = os.MkdirAll(envDir, 0o755)
	code := m.Run()
	os.Exit(code)
}
