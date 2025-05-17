package models

import (
	"time"

	"github.com/jaredfolkins/letemcook/db"
)

// CookbookHistory represents a version of a cookbook that has been archived.
type CookbookHistory struct {
	Created        time.Time `db:"created" json:"created"`
	Updated        time.Time `db:"updated" json:"updated"`
	ID             int64     `db:"id" json:"id"`
	CookbookID     int64     `db:"cookbook_id" json:"cookbook_id"`
	YamlShared     string    `db:"yaml_shared" json:"yaml_shared"`
	YamlIndividual string    `db:"yaml_individual" json:"yaml_individual"`
}

// AppHistory represents a version of an app that has been archived.
type AppHistory struct {
	Created        time.Time `db:"created" json:"created"`
	Updated        time.Time `db:"updated" json:"updated"`
	ID             int64     `db:"id" json:"id"`
	AppID          int64     `db:"app_id" json:"app_id"`
	YamlShared     string    `db:"yaml_shared" json:"yaml_shared"`
	YamlIndividual string    `db:"yaml_individual" json:"yaml_individual"`
}

// TotalHistory returns the combined number of history entries from
// cookbook_history and app_history tables.
func TotalHistory() (int, error) {
	query := `SELECT (SELECT COUNT(*) FROM cookbook_history) + (SELECT COUNT(*) FROM app_history)`
	var total int
	if err := db.Db().Get(&total, query); err != nil {
		return 0, err
	}
	return total, nil
}
