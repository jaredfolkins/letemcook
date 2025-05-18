package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jaredfolkins/letemcook/db"
	"github.com/jaredfolkins/letemcook/embedded"
	"github.com/jaredfolkins/letemcook/middleware"
	"github.com/jaredfolkins/letemcook/util"
	"github.com/labstack/echo/v4"
	"github.com/pressly/goose/v3"
)

func setupTestDB(t *testing.T) func() {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("LEMC_DATA", tmp)
	t.Setenv("LEMC_ENV", "test")
	t.Setenv("LEMC_SQUID_ALPHABET", "abcdefghijklmnopqrstuvwxyz0123456789")

	mfs, err := embedded.GetMigrationsFS()
	if err != nil {
		t.Fatalf("migrations fs: %v", err)
	}
	goose.SetBaseFS(mfs)
	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatalf("set dialect: %v", err)
	}
	dbc := db.Db()
	if err := goose.Up(dbc.DB, "."); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return func() { dbc.Close() }
}

func newCtx(t *testing.T) middleware.LemcContext {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return middleware.NewCustomContext(c)
}

func TestRedirLoginHandler(t *testing.T) {
	teardown := setupTestDB(t)
	defer teardown()
	ctx := newCtx(t)
	if err := redirLoginHandler(ctx); err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if ctx.Response().Status != http.StatusTemporaryRedirect {
		t.Fatalf("status %d", ctx.Response().Status)
	}
	squid, name, _ := util.SquidAndNameByAccountID(1)
	expected := fmt.Sprintf("/lemc/login?squid=%s&account=%s", squid, name)
	if loc := ctx.Response().Header().Get("Location"); loc != expected {
		t.Errorf("location %s != %s", loc, expected)
	}
}

func TestRedirRegisterHandler(t *testing.T) {
	teardown := setupTestDB(t)
	defer teardown()
	ctx := newCtx(t)
	if err := redirRegisterHandler(ctx); err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if ctx.Response().Status != http.StatusTemporaryRedirect {
		t.Fatalf("status %d", ctx.Response().Status)
	}
	squid, name, _ := util.SquidAndNameByAccountID(1)
	expected := fmt.Sprintf("/lemc/register?squid=%s&account=%s", squid, name)
	if loc := ctx.Response().Header().Get("Location"); loc != expected {
		t.Errorf("location %s != %s", loc, expected)
	}
}
