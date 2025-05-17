package util

import (
	"os"
	"path/filepath"
)

func DataRoot() string {
	p := os.Getenv("LEMC_DATA")
	if p == "" {
		wd, _ := os.Getwd()
		p = filepath.Join(wd, "data")
	}
	return p
}

func EnvName() string {
	env := os.Getenv("LEMC_ENV")
	if env == "" {
		env = "production"
	}
	return env
}

func EnvPath() string {
	return filepath.Join(DataRoot(), EnvName())
}

func LockerPath() string {
	return filepath.Join(EnvPath(), "locker")
}

func QueuesPath() string {
	return filepath.Join(LockerPath(), "queues")
}

func AssetsPath() string {
	return filepath.Join(EnvPath(), "assets")
}
