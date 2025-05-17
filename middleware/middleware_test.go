package middleware

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/jaredfolkins/letemcook/models"
    "github.com/labstack/echo-contrib/session"
    "github.com/gorilla/sessions"
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
    // Create the session middleware with an in-memory cookie store
    sessMiddleware := session.Middleware(sessions.NewCookieStore([]byte("secret")))
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

    // Apply our middleware after the session middleware so session data is available
    handlerWithMiddleware := sessMiddleware(Before(handler))

    // Call the composed middleware chain
    err := handlerWithMiddleware(c)

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
