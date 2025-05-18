package models

import "testing"

func TestUserDetailViewTitle(t *testing.T) {
	u := User{Username: "tester"}
	v := UserDetailView{User: u}
	if v.Title() != "User Details: tester" {
		t.Fatalf("unexpected title: %s", v.Title())
	}
}

func TestGetUserIDsForSharedCookbook(t *testing.T) {
	// Without a database configured this just ensures the function can be called.
	if _, err := GetUserIDsForSharedCookbook("missing"); err == nil {
		t.Log("expected error or empty result when db is not initialized")
	}
}

func TestGetUserIDsForSharedApp(t *testing.T) {
	// Without a database configured this just ensures the function can be called.
	if _, err := GetUserIDsForSharedApp("missing"); err == nil {
		t.Log("expected error or empty result when db is not initialized")
	}
}
