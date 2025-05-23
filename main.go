package main

import (
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/jaredfolkins/letemcook/db"
	"github.com/jaredfolkins/letemcook/embedded"
	"github.com/jaredfolkins/letemcook/handlers"
	"github.com/jaredfolkins/letemcook/logger"
	"github.com/jaredfolkins/letemcook/util"
	"github.com/jaredfolkins/letemcook/yeschef"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pressly/goose/v3"
)

const (
	FILE_MODE           fs.FileMode = util.DirPerm
	LEMC_FQDN                       = "localhost"
	DEFAULT_PORT                    = "5362"
	APP_LOG_FILE                    = "app.log"
	HTTP_LOG_FILE                   = "http.log"
	ENV_SECRET_KEY                  = "LEMC_SECRET_KEY"
	LEMC_ENV                        = "production"
	ENV_PORT_DEV                    = "LEMC_PORT_DEV"
	ENV_PORT_TEST                   = "LEMC_PORT_TEST"
	ENV_PORT_PROD                   = "LEMC_PORT_PROD"
	DEFAULT_DOCKER_HOST             = "unix:///var/run/docker.sock"

	LOCKER_FOLDER      = "locker"
	ASSETS_FOLDER      = "assets"
	QUEUES_FOLDER      = "queues"
	NOW_QUEUE_FOLDER   = "now"
	IN_QUEUE_FOLDER    = "in"
	EVERY_QUEUE_FOLDER = "every"
	DATA_FOLDER        = "data"

	ENV_FILE = ".env"
)

func portFromEnv() string {
	env := strings.ToLower(os.Getenv("LEMC_ENV"))
	var port string
	switch env {
	case "dev", "development":
		port = os.Getenv(ENV_PORT_DEV)
	case "test":
		port = os.Getenv(ENV_PORT_TEST)
	default:
		port = os.Getenv(ENV_PORT_PROD)
	}
	if port == "" {
		port = DEFAULT_PORT
	}
	return port
}

func init() {
	if err := util.SetupEnvironment(); err != nil {
		log.Fatal("init error:", err)
	}
}

func main() {
	env := strings.ToLower(os.Getenv("LEMC_ENV"))
	appLogWriter, httpLogWriter, cleanup, err := util.SetupLogWriters(env, APP_LOG_FILE, HTTP_LOG_FILE)
	if err != nil {
		log.Fatalf("log setup: %v", err)
	}
	defer cleanup()
	if env == LEMC_ENV {
		log.SetOutput(appLogWriter)
	}

	logger.InitWithWriter(slog.LevelDebug, appLogWriter)

	e := echo.New()
	e.Debug = true
	e.Logger.SetOutput(appLogWriter)

	e.Use(middleware.Recover())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `${remote_ip} - ${user} [${time}] "${method} ${uri} ${protocol}" ${status} ${bytes_out} "${referer}" "${user_agent}"` + "\n",
		Output: httpLogWriter,
	}))

	sessionPath := util.SessionsPath()
	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		err = os.MkdirAll(sessionPath, FILE_MODE)
		if err != nil {
			log.Fatalf("failed to create session directory: %v", err)
		}
		log.Println("folder created successfully:", sessionPath)
	}
	store := sessions.NewFilesystemStore(sessionPath, []byte(os.Getenv(ENV_SECRET_KEY)))
	store.MaxLength(4096 * 4) // Optional: Set max length, e.g., 16KB

	e.Use(session.Middleware(store))

	assetsFS, err := embedded.GetAssetsFS()
	if err != nil {
		log.Fatal("Failed to get embedded assets FS:", err)
	}

	assetsPath := util.AssetsPath()
	if err := util.DumpFS(assetsFS, assetsPath); err != nil {
		log.Printf("Error dumping assets: %v", err)
	} else {
		log.Println("Successfully dumped all assets to:", assetsPath)
	}

	// Serve assets from the environment specific assets directory
	ap := util.AssetsPath()

	e.GET("/*", echo.WrapHandler(http.StripPrefix("/", http.FileServer(http.Dir(ap)))))

	e.HTTPErrorHandler = handlers.CustomHTTPErrorHandler

	dbConn := db.Db()
	if dbConn == nil || dbConn.DB == nil {
		log.Fatal("Database connection is nil")
	}

	migrationsFS, err := embedded.GetMigrationsFS()
	if err != nil {
		log.Fatalf("failed to get migrations filesystem: %v", err)
	}

	goose.SetBaseFS(migrationsFS)
	if err := goose.SetDialect("sqlite3"); err != nil {
		panic(err)
	}

	if err := goose.Up(dbConn.DB, "."); err != nil {
		panic(fmt.Sprintf("goose up failed: %v", err))
	}

	if env == "development" || env == "dev" || env == "test" {
		seedFS, err := embedded.GetSeedFS()
		if err != nil {
			log.Fatalf("failed to get seed filesystem: %v", err)
		}

		goose.SetBaseFS(seedFS)
		if err := goose.SetDialect("sqlite3"); err != nil {
			panic(err)
		}

		// Check if seed data already exists by querying for a known record
		var count int
		err = dbConn.QueryRow("SELECT COUNT(*) FROM users WHERE username = 'alpha-owner'").Scan(&count)
		if err != nil {
			log.Printf("Error checking for existing seed data: %v", err)
		}

		if count == 0 {
			err = goose.Up(dbConn.DB, ".", goose.WithNoVersioning())
			if err != nil {
				if !strings.Contains(err.Error(), "UNIQUE constraint failed") {
					panic(fmt.Sprintf("goose seed up failed: %v", err))
				}
			}
		}

	}

	user := os.Geteuid()
	group := os.Getegid()
	log.Println("user", user, "group", group)
	handlers.Routes(e)

	yeschef.Start()

	port := portFromEnv()
	e.Logger.Fatal(e.Start(":" + port))
}
