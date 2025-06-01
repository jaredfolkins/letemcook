package handlers

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

// Simple mock templ component for testing
func mockComponent(content string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := w.Write([]byte(content))
		return err
	})
}

func TestHTML(t *testing.T) {
	// Setup echo context
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Test content
	testContent := "<h1>Test HTML Content</h1>"
	mockComp := mockComponent(testContent)

	// Call HTML function
	err := HTML(c, mockComp)

	// Assertions
	if err != nil {
		t.Fatalf("HTML function returned error: %v", err)
	}

	// Check that content type header is set correctly
	contentType := rec.Header().Get(echo.HeaderContentType)
	if contentType != echo.MIMETextHTML {
		t.Errorf("Expected content type %s, got %s", echo.MIMETextHTML, contentType)
	}

	// Check that the content was written to the response
	responseBody := rec.Body.String()
	if !strings.Contains(responseBody, testContent) {
		t.Errorf("Expected response body to contain %s, got %s", testContent, responseBody)
	}
}

func TestHTML_ComponentRenderError(t *testing.T) {
	// Setup echo context
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create a component that returns an error
	errorComponent := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		return io.ErrUnexpectedEOF
	})

	// Call HTML function
	err := HTML(c, errorComponent)

	// Should return the error from component.Render()
	if err != io.ErrUnexpectedEOF {
		t.Errorf("Expected error %v, got %v", io.ErrUnexpectedEOF, err)
	}

	// Content type should still be set even if render fails
	contentType := rec.Header().Get(echo.HeaderContentType)
	if contentType != echo.MIMETextHTML {
		t.Errorf("Expected content type %s, got %s", echo.MIMETextHTML, contentType)
	}
}
