package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/jaredfolkins/letemcook/db"
	"github.com/jaredfolkins/letemcook/logger"
	"github.com/jmoiron/sqlx" // Import sqlx
)

func AiDbDumpMiddleware(dbConn *sqlx.DB) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Aif(func() {
				dumpState, err := db.DumpDatabaseState(dbConn)
				if err != nil {
					logger.Error("Failed to dump database state for AI log", "error", err, "path", r.URL.Path)
					return // Don't proceed if dumping failed
				}

				jsonData, err := json.MarshalIndent(dumpState, "", "  ") // Use Indent for readability
				if err != nil {
					logger.Error("Failed to marshal database dump to JSON for AI log", "error", err, "path", r.URL.Path)
					logger.Ai("Database state dump (non-JSON)", "method", r.Method, "path", r.URL.Path, "db_snapshot_raw", dumpState)
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
