package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSetupEnvironment(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("LEMC_DATA", tmp)
	t.Setenv("LEMC_ENV", "test")
	t.Setenv("LEMC_SQUID_ALPHABET", "abcdefghijklmnopqrstuvwxyz0123456789")

	if err := SetupEnvironment(); err != nil {
		t.Fatalf("setup: %v", err)
	}

	envPath := filepath.Join(tmp, "test")
	if _, err := os.Stat(envPath); err != nil {
		t.Fatalf("env path missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(envPath, ".env")); err != nil {
		t.Fatalf("env file missing: %v", err)
	}
	if v := os.Getenv("LEMC_DOCKER_HOST"); v == "" {
		t.Fatalf("docker host not set")
	}
}

func TestDumpFS(t *testing.T) {
	src := t.TempDir()
	os.WriteFile(filepath.Join(src, "file.txt"), []byte("ok"), 0o644)
	dest := t.TempDir()

	if err := DumpFS(os.DirFS(src), dest); err != nil {
		t.Fatalf("dump: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dest, "file.txt")); err != nil {
		t.Fatalf("file not copied: %v", err)
	}
}

func TestSetupLogWriters(t *testing.T) {
	tmp := t.TempDir()
	app := filepath.Join(tmp, "app.log")
	http := filepath.Join(tmp, "http.log")

	w1, w2, cleanup, err := SetupLogWriters("production", app, http)
	if err != nil {
		t.Fatalf("log writers: %v", err)
	}
	cleanup()
	if w1 == os.Stdout || w2 == os.Stdout {
		t.Fatalf("expected file writers")
	}
	if _, err := os.Stat(app); err != nil {
		t.Fatalf("app log not created")
	}
	if _, err := os.Stat(http); err != nil {
		t.Fatalf("http log not created")
	}

	_, _, _, err = SetupLogWriters("dev", app, http)
	if err != nil {
		t.Fatalf("dev writers: %v", err)
	}
}
