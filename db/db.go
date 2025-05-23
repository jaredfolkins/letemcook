package db

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/jmoiron/sqlx"

	_ "github.com/mattn/go-sqlite3"
)

var db *sqlx.DB

const (
	devDb  string = "lemc_dev.sqlite3"
	testDb string = "lemc_test.sqlite3"
	prodDb        = "lemc_prod.sqlite3"
)

func Db() *sqlx.DB {
	var err error
	if db != nil {
		return db
	}

	name := dbName()
	// Ensure the directory for the SQLite database exists. Without this
	// `sqlx.Open` will succeed but the first query will fail with
	// "unable to open database file" if the parent directory does not
	// exist.
	if err = os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
		log.Fatalf("\U0001F525 failed to prepare database directory: %s", err)
		return nil
	}

	db, err = sqlx.Open("sqlite3", name)
	if err != nil {
		log.Fatalf("🔥 failed to connect to the database: %s", err)
		return nil
	}

	db.SetMaxOpenConns(25)                 // Maximum number of open connections
	db.SetMaxIdleConns(25)                 // Maximum number of idle connections
	db.SetConnMaxLifetime(5 * time.Minute) // Maximum amount of time a connection may be reused

	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		log.Fatalf("🔥 failed to to enable foreign key constraint: %s", err)
		return nil
	}

	log.Println("🚀 Connected Successfully to the Database")

	return db
}

func dbName() string {
	env := os.Getenv("LEMC_ENV")
	path := dataPath()
	switch env {
	case "dev", "development":
		return filepath.Join(path, devDb)
	case "test":
		return filepath.Join(path, testDb)
	}
	return filepath.Join(path, prodDb)
}

// dataPath replicates util.DataPath locally to avoid an import cycle.
// It builds the environment specific data directory location.
func dataPath() string {
	base := os.Getenv("LEMC_DATA")
	if base == "" {
		base = "./data"
	}
	env := os.Getenv("LEMC_ENV")
	if env == "" {
		env = "development"
	}
	return filepath.Join(base, env)
}
