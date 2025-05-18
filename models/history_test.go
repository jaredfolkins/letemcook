package models

import (
	"os"
	"testing"

	"github.com/jaredfolkins/letemcook/db"
	"github.com/jaredfolkins/letemcook/embedded"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
)

var historyTestDB *sqlx.DB

func setupHistoryTests(m *testing.M) int {
	tmpDir, err := os.MkdirTemp("", "lemc_historytest")
	if err != nil {
		panic(err)
	}
	os.Setenv("LEMC_ENV", "test")
	os.Setenv("LEMC_DATA", tmpDir)

	migrationsFS, err := embedded.GetMigrationsFS()
	if err != nil {
		panic(err)
	}
	goose.SetBaseFS(migrationsFS)
	if err := goose.SetDialect("sqlite3"); err != nil {
		panic(err)
	}

	historyTestDB = db.Db()
	if err := goose.Up(historyTestDB.DB, "."); err != nil {
		panic(err)
	}

	code := m.Run()

	historyTestDB.Close()
	os.RemoveAll(tmpDir)
	return code
}

func TestTotalHistory(t *testing.T) {
	total, err := TotalHistory()
	if err != nil {
		t.Fatalf("TotalHistory returned error: %v", err)
	}
	if total != 0 {
		t.Fatalf("expected 0 history entries, got %d", total)
	}

	if _, err := historyTestDB.Exec("INSERT INTO cookbook_history (cookbook_id, yaml_shared, yaml_individual) VALUES (1, 's1', 'i1')"); err != nil {
		t.Fatalf("insert cookbook_history: %v", err)
	}
	if _, err := historyTestDB.Exec("INSERT INTO app_history (app_id, yaml_shared, yaml_individual) VALUES (2, 's2', 'i2')"); err != nil {
		t.Fatalf("insert app_history: %v", err)
	}

	total, err = TotalHistory()
	if err != nil {
		t.Fatalf("TotalHistory returned error after inserts: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected 2 history entries, got %d", total)
	}
}
