package db

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
)

func DumpDatabaseState(db *sqlx.DB) (map[string]interface{}, error) {
	dump := make(map[string]interface{})
	schemaDump := make(map[string]string)
	dataDump := make(map[string][]map[string]interface{})

	tables := []struct {
		Name string `db:"name"`
		Sql  string `db:"sql"`
	}{}
	err := db.Select(&tables, "SELECT name, sql FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' AND name NOT LIKE 'goose_%'")
	if err != nil {
		return nil, fmt.Errorf("failed to query sqlite_master: %w", err)
	}

	for _, table := range tables {
		schemaDump[table.Name] = table.Sql

		query := fmt.Sprintf("SELECT * FROM %s", table.Name)
		rows := []map[string]interface{}{}

		rowsQuery, err := db.Queryx(query)
		if err != nil {
			log.Printf("Warning: Failed to query table %s for dump: %v", table.Name, err) // Log warning but continue
			dataDump[table.Name] = nil                                                    // Indicate error for this table
			continue
		}
               for rowsQuery.Next() {
			row := make(map[string]interface{})
			if err := rowsQuery.MapScan(row); err != nil {
				log.Printf("Warning: Failed to scan row from table %s for dump: %v", table.Name, err) // Log warning
				continue // Skip this row
			}
			for key, val := range row {
				if b, ok := val.([]byte); ok {
					row[key] = fmt.Sprintf("base64:%s", string(b)) // Simple base64 placeholder, consider real encoding
				}
			}
			rows = append(rows, row)
		}

               if err := rowsQuery.Err(); err != nil {
                       log.Printf("Warning: Error during row iteration for table %s: %v", table.Name, err)
               }

               if err := rowsQuery.Close(); err != nil {
                       log.Printf("Warning: failed to close query for table %s: %v", table.Name, err)
               }

               dataDump[table.Name] = rows
	}

	dump["schema"] = schemaDump
	dump["data"] = dataDump

	return dump, nil
}
