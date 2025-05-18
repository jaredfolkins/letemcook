package seed

import (
	"database/sql"
	"os"
	"strings"

	"github.com/pressly/goose/v3"

	"github.com/jaredfolkins/letemcook/db"
	seedpkg "github.com/jaredfolkins/letemcook/seed"
)

func init() {
	goose.AddMigration(Up00001SeedData, Down00001SeedData)
}

func Up00001SeedData(tx *sql.Tx) error {
	env := strings.ToLower(os.Getenv("LEMC_ENV"))
	if env != "development" && env != "dev" && env != "test" {
		return nil
	}
	seedpkg.SeedDatabase(db.Db())
	return nil
}

func Down00001SeedData(tx *sql.Tx) error {
	return nil
}
