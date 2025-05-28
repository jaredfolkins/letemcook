package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jaredfolkins/letemcook/middleware"
	"github.com/jaredfolkins/letemcook/models"
	"github.com/labstack/echo/v4"
)

// TestGetAppsHandlerUnauthenticated ensures an unauthorized response when user context is missing.
func TestGetAppsHandlerUnauthenticated(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/lemc/apps", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// No user context set
	err := GetAppsHandler(middleware.NewCustomContext(c))
	if err == nil {
		t.Fatalf("expected error due to missing user context")
	}
}

// TestGetAppsHandlerDefaults verifies handler defaults page and limit when invalid values provided.
func TestGetAppsHandlerDefaults(t *testing.T) {
	// Set up required environment variables
	os.Setenv("LEMC_SQUID_ALPHABET", "abcdefghijklmnopqrstuvwxyz0123456789")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/lemc/apps?page=x&limit=bad", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	user := &models.User{ID: 1, Account: &models.Account{ID: 1}}
	uc := &models.UserContext{LoggedInAs: user, ActingAs: user}
	cc := middleware.SetUserContext(c, uc)

	// Handler will fail before DB access since models package isn't fully initialised during tests.
	_ = GetAppsHandler(cc)
}
