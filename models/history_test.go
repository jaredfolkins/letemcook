package models

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jaredfolkins/letemcook/db"
	"github.com/jaredfolkins/letemcook/embedded"
	"github.com/jaredfolkins/letemcook/tests/testutil"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
)

var historyTestDB *sqlx.DB

func TestMain(m *testing.M) {
	dataRoot := testutil.DataRoot()
	os.Setenv("LEMC_ENV", "test")
	os.Setenv("LEMC_DATA", dataRoot)
	envDir := filepath.Join(dataRoot, "test")
	os.Remove(filepath.Join(envDir, "lemc_test.sqlite3"))

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

	// Insert prerequisite data for foreign key constraints
	if _, err := historyTestDB.Exec("INSERT INTO accounts (id, squid, name) VALUES (1, 'testsquid', 'test-account')"); err != nil {
		panic("insert account: " + err.Error())
	}
	if _, err := historyTestDB.Exec("INSERT INTO users (id, username, email, hash) VALUES (1, 'testuser', 'user@example.com', 'hash')"); err != nil {
		panic("insert user: " + err.Error())
	}
	// Note: is_published and is_deleted default to false. api_key is NOT NULL.
	if _, err := historyTestDB.Exec("INSERT INTO cookbooks (id, account_id, owner_id, uuid, name, description, yaml_shared, yaml_individual, api_key) VALUES (1, 1, 1, 'cookbook-uuid1', 'test-cookbook', '', '', '', 'testapikey1')"); err != nil {
		panic("insert cookbook: " + err.Error())
	}
	// Note: is_active and is_deleted default to false. api_key is NOT NULL.
	if _, err := historyTestDB.Exec("INSERT INTO apps (id, account_id, owner_id, cookbook_id, uuid, name, description, yaml_shared, yaml_individual, api_key) VALUES (2, 1, 1, 1, 'app-uuid2', 'test-app', '', '', '', 'testapikey2')"); err != nil {
		panic("insert app: " + err.Error())
	}

	code := m.Run()

	historyTestDB.Close()
	os.Exit(code)
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
