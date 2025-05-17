package util

import (
	"os"
	"path/filepath"
)

// DataPath returns the environment specific data directory.
func DataPath() string {
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

// LockerPath returns the path to the locker directory.
func LockerPath() string {
	return filepath.Join(DataPath(), "locker")
}

// QueuesPath returns the path to the job queues directory.
func QueuesPath() string {
	return filepath.Join(LockerPath(), "queues")
}

// SessionsPath returns the path to the sessions directory.
func SessionsPath() string {
	return filepath.Join(DataPath(), "sessions")
}

// AssetsPath returns the path to the assets directory.
func AssetsPath() string {
	return filepath.Join(DataPath(), "assets")
}
