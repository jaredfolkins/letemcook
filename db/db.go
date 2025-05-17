package db

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/jaredfolkins/letemcook/util"
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

	db, err = sqlx.Open("sqlite3", dbName())
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
	path := util.DataPath()
	switch env {
	case "dev", "development":
		return filepath.Join(path, devDb)
	case "test":
		return filepath.Join(path, testDb)
	}
	return filepath.Join(path, prodDb)
}
