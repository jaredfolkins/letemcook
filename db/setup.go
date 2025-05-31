//go:build test

package db

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jaredfolkins/letemcook/embedded"
	"github.com/jaredfolkins/letemcook/util"
	"github.com/pressly/goose/v3"
)

// SetupTestDB initializes an isolated SQLite database for testing and returns a teardown function.
func SetupTestDB(t *testing.T) func() {
	t.Helper()
	os.Setenv("LEMC_ENV", "test")
	dataRoot := util.TestDataRoot()
	os.Setenv("LEMC_DATA", dataRoot)

	envDir := filepath.Join(dataRoot, "test")
	if err := os.MkdirAll(envDir, 0o755); err != nil {
		t.Fatalf("prepare db dir: %v", err)
	}

	os.Remove(filepath.Join(envDir, "lemc_test.sqlite3"))

	mfs, err := embedded.GetMigrationsFS()
	if err != nil {
		t.Fatalf("migrations fs: %v", err)
	}
	goose.SetBaseFS(mfs)
	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatalf("set dialect: %v", err)
	}
	dbc := Db()
	if err := goose.Up(dbc.DB, "."); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return func() { dbc.Close() }
}
