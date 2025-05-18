package models

import (
	"encoding/base64"
	"reflect"
	"testing"
)

func TestExtractImgSrcs(t *testing.T) {
	html := `<div><img src="a.png"><p>x</p><img src="/img/b.jpg"></div>`
	got, err := extractImgSrcs(html)
	if err != nil {
		t.Fatalf("err %v", err)
	}
	expect := []string{"a.png", "/img/b.jpg"}
	if !reflect.DeepEqual(got, expect) {
		t.Fatalf("%v != %v", got, expect)
	}
}

func TestGetLastParameter(t *testing.T) {
	if got := getLastParameter("http://x/y/z"); got != "z" {
		t.Fatalf("got %s", got)
	}
}

func TestStoragePurgeUnusedFiles(t *testing.T) {
	wiki := `<img src="file1.png">`
	storage := Storage{
		Files: map[string]string{"file1.png": "a", "file2.png": "b"},
		Wikis: map[int]string{1: base64.StdEncoding.EncodeToString([]byte(wiki))},
	}
	if err := storage.PurgeUnusedFiles(); err != nil {
		t.Fatalf("purge: %v", err)
	}
	if _, ok := storage.Files["file1.png"]; !ok {
		t.Errorf("file1 removed")
	}
	if _, ok := storage.Files["file2.png"]; ok {
		t.Errorf("file2 not deleted")
	}
}

func TestRecipeUsernameOrAdmin(t *testing.T) {
	r := Recipe{Name: "abc"}
	if r.UsernameOrAdmin() != "abc" {
		t.Fatalf("got %s", r.UsernameOrAdmin())
	}
	r.IsShared = true
	if r.UsernameOrAdmin() != "shared" {
		t.Fatalf("shared got %s", r.UsernameOrAdmin())
	}
}

func TestNewYamlIndividual(t *testing.T) {
	y := NewYamlIndividual()
	if y == nil || y.Cookbook.Storage.Files == nil || len(y.Cookbook.Storage.Files) != 0 {
		t.Fatalf("unexpected %#v", y)
	}
}
