package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jaredfolkins/letemcook/models"
	"github.com/labstack/echo/v4"
)

func TestNewCustomContext(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create a custom context
	cc := NewCustomContext(c)

	// Assert
	if cc == nil {
		t.Error("Expected non-nil custom context")
	}

	if cc.UserContext() == nil {
		t.Error("Expected non-nil user context")
	}
}

func TestBeforeMiddleware(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Define a handler to use after middleware
	handler := func(c echo.Context) error {
		// Try to cast to LemcContext
		cc, ok := c.(LemcContext)
		if !ok {
			t.Error("Failed to cast to LemcContext")
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		// Verify userContext exists
		if cc.UserContext() == nil {
			t.Error("Expected non-nil UserContext")
		}

		return nil
	}

	// Apply middleware
	middleware := Before(handler)

	// Call the middleware
	err := middleware(c)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestSetUserContext(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create a test user context
	testUserCtx := &models.UserContext{}

	// Set the user context
	cc := SetUserContext(c, testUserCtx)

	// Assert
	if cc == nil {
		t.Error("Expected non-nil context after SetUserContext")
	}

	if cc.UserContext() != testUserCtx {
		t.Error("User context not set correctly")
	}
}
