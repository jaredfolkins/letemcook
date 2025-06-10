package views

import (
	"github.com/a-h/templ"
	"testing"
)

func TestJoinURLErrs(t *testing.T) {
	url, err := templ.JoinURLErrs("https://example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(url) != "https://example.com" {
		t.Fatalf("unexpected url: %s", url)
	}
}
