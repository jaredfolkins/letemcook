package db

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jaredfolkins/letemcook/embedded"
	"github.com/pressly/goose/v3"
)

// SetupTestDB initializes an isolated SQLite database for testing and returns a teardown function.
func SetupTestDB(t *testing.T) func() {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("LEMC_DATA", tmp)
	t.Setenv("LEMC_ENV", "test")
	t.Setenv("LEMC_SQUID_ALPHABET", "abcdefghijklmnopqrstuvwxyz0123456789")

	// Ensure the environment-specific directory exists for the SQLite file.
	if err := os.MkdirAll(filepath.Join(tmp, "test"), 0o755); err != nil {
		t.Fatalf("prepare db dir: %v", err)
	}

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
