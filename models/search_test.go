package models

import (
	"testing"

	"github.com/jaredfolkins/letemcook/db"
	"golang.org/x/crypto/bcrypt"
)

func createAccountAndUsers(t *testing.T) (*Account, *User, *User) {
	dbc := db.Db()
	tx, err := dbc.Beginx()
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	acc, err := AccountCreate("Test", tx)
	if err != nil {
		t.Fatalf("acct: %v", err)
	}
	pw, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	u1 := NewUser()
	u1.Username = "user1"
	u1.Email = "user1@example.com"
	u1.Hash = string(pw)
	u1.Heckle = false
	id1, err := CreateUserWithAccountID(u1, acc.ID, tx)
	if err != nil {
		t.Fatalf("user1: %v", err)
	}
	u1.ID = id1
	u2 := NewUser()
	u2.Username = "user2"
	u2.Email = "user2@example.com"
	u2.Hash = string(pw)
	u2.Heckle = false
	id2, err := CreateUserWithAccountID(u2, acc.ID, tx)
	if err != nil {
		t.Fatalf("user2: %v", err)
	}
	u2.ID = id2
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
	return acc, u1, u2
}

func createCookbookAndApp(t *testing.T, acc *Account, owner *User) (*Cookbook, *App) {
	dbc := db.Db()
	tx, err := dbc.Beginx()
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	cb := &Cookbook{AccountID: acc.ID, OwnerID: owner.ID, Name: "cb1", Description: "d", YamlShared: "y", YamlIndividual: "y"}
	if err := cb.Create(tx); err != nil {
		t.Fatalf("cb create: %v", err)
	}
	app := &App{AccountID: acc.ID, OwnerID: owner.ID, CookbookID: cb.ID, Name: "app1", Description: "d", YAMLShared: "y", YAMLIndividual: "y", IsMcpEnabled: true, IsActive: true}
	if err := app.Create(tx); err != nil {
		t.Fatalf("app create: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
	return cb, app
}

func TestSearchFunctions(t *testing.T) {
	teardown := db.SetupTestDB(t)
	defer teardown()

	acc, u1, u2 := createAccountAndUsers(t)
	cb, app := createCookbookAndApp(t, acc, u1)

	// user1 already assigned via Create; assign app to user1
	dbc := db.Db()
	if _, err := dbc.Exec(`INSERT INTO permissions_apps (user_id, account_id, app_id, cookbook_id, can_shared) VALUES (?, ?, ?, ?, true)`, u1.ID, acc.ID, app.ID, cb.ID); err != nil {
		t.Fatalf("perm app: %v", err)
	}

	// assign cookbook to user1 only
	if _, err := dbc.Exec(`INSERT INTO permissions_cookbooks (user_id, account_id, cookbook_id, can_view) VALUES (?, ?, ?, true)`, u2.ID, acc.ID, cb.ID+1); err == nil {
		// ensure user2 not assigned to cb
	}

	// Search for users not assigned to cookbook
	res, err := SearchForCookbookAclUsersNotAssigned("user", acc.ID, cb.ID, 10)
	if err != nil {
		t.Fatalf("search acl not assigned: %v", err)
	}
	if len(res) != 1 || res[0].UserID != u2.ID {
		t.Fatalf("unexpected result: %#v", res)
	}

	// Search for cookbooks
	cbs, err := SearchForCookbooks("cb", u1.ID, acc.ID, 10, false)
	if err != nil {
		t.Fatalf("search cookbooks: %v", err)
	}
	if len(cbs) == 0 {
		t.Fatalf("expected cookbook result")
	}

	// Cookbook ACL users
	if _, err := dbc.Exec(`INSERT INTO permissions_cookbooks (user_id, account_id, cookbook_id, can_view) VALUES (?, ?, ?, true)`, u2.ID, acc.ID, cb.ID); err != nil {
		t.Fatalf("perm cb: %v", err)
	}
	users, err := CookbookAclsUsers(acc.ID, cb.ID)
	if err != nil {
		t.Fatalf("cookbook acls users: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}

	// Search for app acl users not assigned
	res2, err := SearchForappAclUsersNotAssigned("user", acc.ID, app.ID, 10)
	if err != nil {
		t.Fatalf("search app acl: %v", err)
	}
	if len(res2) != 1 || res2[0].UserID != u2.ID {
		t.Fatalf("unexpected result: %#v", res2)
	}
}
