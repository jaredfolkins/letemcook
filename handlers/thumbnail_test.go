//go:build test

package handlers

import (
	"encoding/base64"
	"testing"

	"github.com/jaredfolkins/letemcook/models"
)

func TestDecodeThumbnailData_Empty(t *testing.T) {
	data, err := decodeThumbnailData(models.YamlDefault{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data != nil {
		t.Fatalf("expected nil data for empty input")
	}
}

func TestDecodeThumbnailData_Invalid(t *testing.T) {
	y := models.YamlDefault{}
	y.Cookbook.Storage.Thumbnail.B64 = "invalid"
	if _, err := decodeThumbnailData(y); err == nil {
		t.Fatalf("expected error for invalid base64")
	}
}

func TestDecodeThumbnailData_Valid(t *testing.T) {
	y := models.YamlDefault{}
	orig := []byte("hello")
	y.Cookbook.Storage.Thumbnail.B64 = base64.StdEncoding.EncodeToString(orig)
	data, err := decodeThumbnailData(y)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != string(orig) {
		t.Fatalf("expected %q got %q", orig, data)
	}
}
