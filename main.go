package main

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"log"
	"log/slog"
	mrand "math/rand"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/jaredfolkins/letemcook/db"
	"github.com/jaredfolkins/letemcook/embedded"
	"github.com/jaredfolkins/letemcook/handlers"
	"github.com/jaredfolkins/letemcook/logger"
	"github.com/jaredfolkins/letemcook/seed"
	"github.com/jaredfolkins/letemcook/util"
	"github.com/jaredfolkins/letemcook/yeschef"
	"github.com/joho/godotenv"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pressly/goose/v3"

	"path/filepath"
	"time"
)

const (
	FILE_MODE           fs.FileMode = 0744
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

func generateHash() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func generateRandom64BitNumber() (uint64, error) {
	var num uint64
	err := binary.Read(rand.Reader, binary.LittleEndian, &num)
	if err != nil {
		return 0, err
	}
	return num, nil
}

func generateAlphabet() string {
	mrand.Seed(time.Now().UnixNano())

	alphabet := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	mrand.Shuffle(len(alphabet), func(i, j int) {
		alphabet[i], alphabet[j] = alphabet[j], alphabet[i]
	})

	return string(alphabet)
}

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

	dir, err := os.Getwd()
	if err != nil {
		log.Fatal("init() error:", err)
	}

	envValue := os.Getenv("LEMC_ENV")
	if envValue == "" {
		envValue = LEMC_ENV
	}

	dataRoot := filepath.Join(dir, DATA_FOLDER)
	if _, err := os.Stat(dataRoot); os.IsNotExist(err) {
		err = os.Mkdir(dataRoot, FILE_MODE)
		if err != nil {
			log.Fatal("init() error:", err)
		}
		log.Println("folder created successfully:", dataRoot)
	}

	data := filepath.Join(dataRoot, envValue)
	if _, err := os.Stat(data); os.IsNotExist(err) {
		err = os.MkdirAll(data, FILE_MODE)
		if err != nil {
			log.Fatal("init() error:", err)
		}
		log.Println("folder created successfully:", data)
	}

	envfile := filepath.Join(data, ENV_FILE)
	if _, err := os.Stat(envfile); os.IsNotExist(err) {
		f, err := os.Create(envfile)
		if err != nil {
			log.Fatal("init() error:", err)
		}

		secret_key, err := generateHash()
		if err != nil {
			log.Fatal("init() error:", err)
		}

		api_key, err := generateHash()
		if err != nil {
			log.Fatal("init() error:", err)
		}

		f.WriteString(fmt.Sprintf("LEMC_DATA=%s\n", filepath.Join(dir, DATA_FOLDER)))
		f.WriteString(fmt.Sprintf("LEMC_ENV=%s\n", envValue))
		f.WriteString(fmt.Sprintf("LEMC_FQDN=%s\n", LEMC_FQDN))
		f.WriteString(fmt.Sprintf("LEMC_DEFAULT_THEME=%s\n", util.DefaultTheme))
		f.WriteString(fmt.Sprintf("LEMC_GLOBAL_API_KEY=%s\n", api_key))
		f.WriteString(fmt.Sprintf("LEMC_SECRET_KEY=%s\n", secret_key))
		f.WriteString(fmt.Sprintf("LEMC_SQUID_ALPHABET=%s\n", generateAlphabet()))
		f.WriteString(fmt.Sprintf("LEMC_DOCKER_HOST=%s\n", DEFAULT_DOCKER_HOST))
		log.Println(".env created successfully:", f.Name())
		f.Close()
	}

	err = godotenv.Load(envfile)
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	if os.Getenv("LEMC_DOCKER_HOST") == "" {
		os.Setenv("LEMC_DOCKER_HOST", DEFAULT_DOCKER_HOST)
	}

	qf := util.QueuesPath()
	if _, err := os.Stat(qf); os.IsNotExist(err) {
		err = os.Mkdir(qf, FILE_MODE)
		if err != nil {
			log.Fatal("init() error:", err)
		}
		log.Println("folder created successfully:", qf)
	}

	nowQf := filepath.Join(qf, NOW_QUEUE_FOLDER)
	if _, err := os.Stat(nowQf); os.IsNotExist(err) {
		err = os.Mkdir(nowQf, FILE_MODE)
		if err != nil {
			log.Fatal("init() error:", err)
		}
		log.Println("folder created successfully:", nowQf)
	}

	inQf := filepath.Join(qf, IN_QUEUE_FOLDER)
	if _, err := os.Stat(inQf); os.IsNotExist(err) {
		err = os.Mkdir(inQf, FILE_MODE)
		if err != nil {
			log.Fatal("init() error:", err)
		}
		log.Println("folder created successfully:", inQf)
	}

	everyQf := filepath.Join(qf, EVERY_QUEUE_FOLDER)
	if _, err := os.Stat(everyQf); os.IsNotExist(err) {
		err = os.Mkdir(everyQf, FILE_MODE)
		if err != nil {
			log.Fatal("init() error:", err)
		}
		log.Println("folder created successfully:", everyQf)
	}

	path := util.LockerPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.Mkdir(path, FILE_MODE)
		if err != nil {
			log.Fatal("init() error:", err)
		}
		log.Println("folder created successfully:", path)
	}

	path = filepath.Join(path, ".gitignore")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		f, err := os.Create(path)
		if err != nil {
			log.Fatal("init() error:", err)
		}
		defer f.Close()
		log.Println(".gitignore created successfully:", f.Name())
	}

}

func main() {
	var appLogWriter io.Writer = os.Stdout
	var httpLogWriter io.Writer = os.Stdout
	env := os.Getenv(LEMC_ENV)
	if env == LEMC_ENV {
		appFile, err := os.OpenFile(APP_LOG_FILE, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening app log file: %v", err)
		}
		defer appFile.Close()
		log.SetOutput(appFile)
		appLogWriter = appFile

		httpFile, err := os.OpenFile(HTTP_LOG_FILE, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening http log file: %v", err)
		}
		defer httpFile.Close()
		httpLogWriter = httpFile
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

	sessionPath := filepath.Join(os.Getenv("LEMC_DATA"), "sessions")
	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		err = os.Mkdir(sessionPath, FILE_MODE)
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

	// Create a directory to dump assets if it doesn't exist
	assetsPath := filepath.Join(os.Getenv("LEMC_DATA"), "assets")
	if _, err := os.Stat(assetsPath); os.IsNotExist(err) {
		err = os.MkdirAll(assetsPath, FILE_MODE)
		if err != nil {
			log.Fatalf("failed to create assets directory: %v", err)
		}
		log.Println("folder created successfully:", assetsPath)
	}

	err = fs.WalkDir(assetsFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory
		if path == "." {
			return nil
		}

		destPath := filepath.Join(assetsPath, path)

		if d.IsDir() {
			// Create directory
			if err := os.MkdirAll(destPath, FILE_MODE); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", destPath, err)
			}
			return nil
		}

		// Read file content
		content, err := fs.ReadFile(assetsFS, path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}

		// Write file content
		if err := os.WriteFile(destPath, content, FILE_MODE); err != nil {
			return fmt.Errorf("failed to write file %s: %w", destPath, err)
		}

		return nil
	})

	if err != nil {
		log.Printf("Error dumping assets: %v", err)
	} else {
		log.Println("Successfully dumped all assets to:", assetsPath)
	}

	var ap string
	if os.Getenv("LEMC_ENV") == "dev" || os.Getenv("LEMC_ENV") == "development" {
		ap = filepath.Join("embedded", "assets")
	} else {
		ap = filepath.Join(os.Getenv("LEMC_DATA"), "assets")
	}

	e.GET("/*", echo.WrapHandler(http.StripPrefix("/", http.FileServer(http.Dir(ap)))))

	e.HTTPErrorHandler = handlers.CustomHTTPErrorHandler

	migrationsFS, err := embedded.GetMigrationsFS()
	if err != nil {
		log.Fatalf("failed to get migrations filesystem: %v", err)
	}

	goose.SetBaseFS(migrationsFS)
	if err := goose.SetDialect("sqlite3"); err != nil {
		panic(err)
	}

	dbConn := db.Db()
	if dbConn == nil {
		log.Fatal("Database connection is nil")
	}
	if dbConn.DB == nil {
		log.Fatal("Database DB field is nil")
	}

	if err := goose.Up(dbConn.DB, "."); err != nil {
		panic(fmt.Sprintf("goose up failed: %v", err))
	}

	seed.SeedDatabaseIfDev(db.Db())

	handlers.Routes(e)

	yeschef.Start()

	port := portFromEnv()
	e.Logger.Fatal(e.Start(":" + port))
}
