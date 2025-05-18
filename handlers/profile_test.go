package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jaredfolkins/letemcook/db"
	"github.com/jaredfolkins/letemcook/embedded"
	"github.com/jaredfolkins/letemcook/middleware"
	"github.com/jaredfolkins/letemcook/models"
	"github.com/labstack/echo/v4"
	"github.com/pressly/goose/v3"
	"golang.org/x/crypto/bcrypt"
)

func setupProfileTestDB(t *testing.T) func() {
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

func createUser(t *testing.T) (*models.Account, *models.User) {
	dbc := db.Db()
	tx, err := dbc.Beginx()
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	acc, err := models.AccountCreate("Test", tx)
	if err != nil {
		t.Fatalf("acct: %v", err)
	}
	pw, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	u := models.NewUser()
	u.Username = "user1"
	u.Email = "u@example.com"
	u.Hash = string(pw)
	u.Heckle = false
	id, err := models.CreateUserWithAccountID(u, acc.ID, tx)
	if err != nil {
		t.Fatalf("user create: %v", err)
	}
	u.ID = id
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
	return acc, u
}

func newContext(t *testing.T, method, target string, user *models.User) LemcContext {
	e := echo.New()
	req := httptest.NewRequest(method, target, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	uc := &models.UserContext{ActingAs: user, LoggedInAs: user}
	return middleware.SetUserContext(c, uc)
}

func TestGetProfileHandler(t *testing.T) {
	teardown := setupProfileTestDB(t)
	defer teardown()
	_, user := createUser(t)
	ctx := newContext(t, http.MethodGet, "/lemc/profile?partial=true", user)
	if err := GetProfileHandler(ctx); err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if ctx.Response().Status != http.StatusOK {
		t.Fatalf("status %d", ctx.Response().Status)
	}
}

func TestPostChangePasswordHandlerMismatch(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/lemc/profile/password", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	uc := &models.UserContext{} // not authenticated
	c := middleware.SetUserContext(ctx, uc)
	if err := PostChangePasswordHandler(c); err == nil {
		t.Fatal("expected error")
	}
}

func TestPostToggleHeckleHandler(t *testing.T) {
	teardown := setupProfileTestDB(t)
	defer teardown()
	_, user := createUser(t)
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/lemc/profile/heckle", strings.NewReader("heckle_enabled=on"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	uc := &models.UserContext{ActingAs: user, LoggedInAs: user}
	c := middleware.SetUserContext(ctx, uc)
	if err := PostToggleHeckleHandler(c); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	u, err := models.UserByIDAndAccountID(user.ID, user.Account.ID)
	if err != nil {
		t.Fatalf("lookup user: %v", err)
	}
	if !u.Heckle {
		t.Error("heckle not updated")
	}
}
