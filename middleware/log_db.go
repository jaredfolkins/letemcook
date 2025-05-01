package middleware

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/jaredfolkins/letemcook/logger" // Correct import path (new location)
)

func placeholderDump(database *sql.DB) (interface{}, error) {
	return map[string]interface{}{"users_count": 1, "products_count": 0}, nil
}

func LogDatabaseStateMiddleware(database *sql.DB) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Aif(func() {
				state, err := placeholderDump(database) // Use the placeholder
				if err != nil {
					logger.Error("Error dumping database state for AI log", "error", err, "path", r.URL.Path)
					return
				}

				jsonData, err := json.Marshal(state)
				if err != nil {
					logger.Error("Error marshalling database state to JSON for AI log", "error", err, "path", r.URL.Path)
					logger.Ai("Database state dump (non-JSON)", "state", state, "path", r.URL.Path)
					return
				}

				logger.Ai(
					"Database state dump for request",
					"method", r.Method,
					"path", r.URL.Path,
					"db_snapshot_json", string(jsonData),
				)
			})

			next.ServeHTTP(w, r)
		})
	}
}
