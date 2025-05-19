//go:build test

package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jaredfolkins/letemcook/db"
	"github.com/jaredfolkins/letemcook/middleware"
	"github.com/jaredfolkins/letemcook/util"
	"github.com/labstack/echo/v4"
)

func newCtx(t *testing.T) middleware.LemcContext {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return middleware.NewCustomContext(c)
}

func TestRedirLoginHandler(t *testing.T) {
	teardown := db.SetupTestDB(t)
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
	teardown := db.SetupTestDB(t)
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
