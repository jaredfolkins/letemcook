package util

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	mathrand "math/rand"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

const FileMode fs.FileMode = DirPerm

// GenerateHash returns a random 16 byte hex string.
func GenerateHash() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// GenerateAlphabet returns a shuffled alphanumeric alphabet used for squid ids.
func GenerateAlphabet() string {
	r := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
	alpha := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	r.Shuffle(len(alpha), func(i, j int) { alpha[i], alpha[j] = alpha[j], alpha[i] })
	return string(alpha)
}

// SetupEnvironment initializes directories and the .env file.
func SetupEnvironment() error {
	syscall.Umask(0)
	envValue := EnvName()
	dataRoot := DataRoot()

	if err := os.MkdirAll(dataRoot, FileMode); err != nil {
		return err
	}

	data := filepath.Join(dataRoot, envValue)
	if err := os.MkdirAll(data, FileMode); err != nil {
		return err
	}

	envFile := filepath.Join(data, ".env")
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		f, err := os.Create(envFile)
		if err != nil {
			return err
		}
		defer f.Close()

		secret, err := GenerateHash()
		if err != nil {
			return err
		}
		api, err := GenerateHash()
		if err != nil {
			return err
		}

		f.WriteString(fmt.Sprintf("LEMC_DATA=%s\n", dataRoot))
		f.WriteString(fmt.Sprintf("LEMC_ENV=%s\n", envValue))
		f.WriteString("LEMC_FQDN=localhost\n")
		f.WriteString(fmt.Sprintf("LEMC_DEFAULT_THEME=%s\n", DefaultTheme))
		f.WriteString(fmt.Sprintf("LEMC_GLOBAL_API_KEY=%s\n", api))
		f.WriteString(fmt.Sprintf("LEMC_SECRET_KEY=%s\n", secret))
		f.WriteString(fmt.Sprintf("LEMC_SQUID_ALPHABET=%s\n", GenerateAlphabet()))
		f.WriteString("LEMC_DOCKER_HOST=unix:///var/run/docker.sock\n")
	}

	if err := godotenv.Load(envFile); err != nil {
		return fmt.Errorf("load env: %w", err)
	}

	if os.Getenv("LEMC_DOCKER_HOST") == "" {
		os.Setenv("LEMC_DOCKER_HOST", "unix:///var/run/docker.sock")
	}

	qf := QueuesPath()
	if err := os.MkdirAll(filepath.Join(qf, "now"), FileMode); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(qf, "in"), FileMode); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(qf, "every"), FileMode); err != nil {
		return err
	}

	lp := LockerPath()
	if err := os.MkdirAll(lp, FileMode); err != nil {
		return err
	}
	gi := filepath.Join(lp, ".gitignore")
	if _, err := os.Stat(gi); os.IsNotExist(err) {
		if f, err := os.Create(gi); err == nil {
			f.Close()
		} else {
			return err
		}
	}

	return nil
}

// DumpFS copies all files from the provided FS into the destination directory.
func DumpFS(src fs.FS, dest string) error {
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		if err := os.MkdirAll(dest, FileMode); err != nil {
			return err
		}
	}

	return fs.WalkDir(src, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}
		dst := filepath.Join(dest, path)
		if d.IsDir() {
			return os.MkdirAll(dst, FileMode)
		}
		content, err := fs.ReadFile(src, path)
		if err != nil {
			return err
		}
		return os.WriteFile(dst, content, FileMode)
	})
}

// SetupLogWriters returns writers for app and http logs based on env.
func SetupLogWriters(env, appPath, httpPath string) (io.Writer, io.Writer, func(), error) {
	cleanup := func() {}
	if strings.ToLower(env) != "production" {
		return os.Stdout, os.Stdout, cleanup, nil
	}

	appFile, err := os.OpenFile(appPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
	if err != nil {
		return nil, nil, cleanup, err
	}
	httpFile, err := os.OpenFile(httpPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
	if err != nil {
		appFile.Close()
		return nil, nil, cleanup, err
	}
	cleanup = func() { appFile.Close(); httpFile.Close() }
	return appFile, httpFile, cleanup, nil
}
