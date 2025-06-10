package partials

import (
	"bytes"
	"context"
	"testing"
)

func TestPrintWiki_InvalidBase64(t *testing.T) {
	var buf bytes.Buffer
	err := printWiki("invalid-base64").Render(context.Background(), &buf)
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
}
