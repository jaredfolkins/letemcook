package models

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

var (
	testDB           *sqlx.DB
	testUserID       int64
	testAccountID    int64
	testCookbookID   int64
	testAppID        int64
	testCookbookUUID string
	testAppUUID      string
)

func seedPermissionTestData() error {
	res, err := testDB.Exec("INSERT INTO accounts (squid, name) VALUES (?, ?)", "acc", "Test Account")
	if err != nil {
		return err
	}
	testAccountID, _ = res.LastInsertId()

	hash, _ := bcrypt.GenerateFromPassword([]byte("pw"), bcrypt.DefaultCost)
	res, err = testDB.Exec("INSERT INTO users (username, email, hash) VALUES (?, ?, ?)", "user", "user@example.com", string(hash))
	if err != nil {
		return err
	}
	testUserID, _ = res.LastInsertId()

	if _, err = testDB.Exec("INSERT INTO permissions_system (user_id, can_administer, is_owner) VALUES (?,1,1)", testUserID); err != nil {
		return err
	}

	if _, err = testDB.Exec(`INSERT INTO permissions_accounts (user_id, account_id, can_administer, can_create_apps, can_view_apps, can_create_cookbooks, can_view_cookbooks, is_owner) VALUES (?, ?, 1,1,1,1,1,1)`, testUserID, testAccountID); err != nil {
		return err
	}

	testCookbookUUID = "cbuuid"
	res, err = testDB.Exec(`INSERT INTO cookbooks (account_id, owner_id, uuid, name, description, yaml_shared, yaml_individual, api_key) VALUES (?, ?, ?, 'cb', '', '', '', 'cbkey')`, testAccountID, testUserID, testCookbookUUID)
	if err != nil {
		return err
	}
	testCookbookID, _ = res.LastInsertId()

	if _, err = testDB.Exec(`INSERT INTO permissions_cookbooks (user_id, account_id, cookbook_id, can_view, can_edit, is_owner) VALUES (?, ?, ?, 1,1,1)`, testUserID, testAccountID, testCookbookID); err != nil {
		return err
	}

	testAppUUID = "appuuid"
	res, err = testDB.Exec(`INSERT INTO apps (account_id, owner_id, cookbook_id, uuid, name, description, yaml_shared, yaml_individual, api_key) VALUES (?, ?, ?, ?, 'app', '', '', '', 'appkey')`, testAccountID, testUserID, testCookbookID, testAppUUID)
	if err != nil {
		return err
	}
	testAppID, _ = res.LastInsertId()

	if _, err = testDB.Exec(`INSERT INTO permissions_apps (user_id, account_id, app_id, cookbook_id, can_shared, can_individual, can_administer, is_owner, api_key) VALUES (?, ?, ?, ?, 1,1,1,1, 'permkey')`, testUserID, testAccountID, testAppID, testCookbookID); err != nil {
		return err
	}
	return nil
}

func TestHasSystemPermission(t *testing.T) {
	ok, err := HasSystemPermission(testUserID, CanAdministerSystem)
	if err != nil {
		t.Fatalf("HasSystemPermission returned error: %v", err)
	}
	if !ok {
		t.Errorf("expected permission to be granted")
	}
}

func TestHasAccountPermission(t *testing.T) {
	ok, err := HasAccountPermission(testUserID, testAccountID, CanCreateApp)
	if err != nil {
		t.Fatalf("HasAccountPermission returned error: %v", err)
	}
	if !ok {
		t.Errorf("expected permission to be granted")
	}
}

func TestHasCookbookPermission(t *testing.T) {
	ok, err := HasCookbookPermission(testUserID, testAccountID, testCookbookUUID, CanEditCookbook)
	if err != nil {
		t.Fatalf("HasCookbookPermission returned error: %v", err)
	}
	if !ok {
		t.Errorf("expected permission to be granted")
	}
}

func TestHasAppPermission(t *testing.T) {
	ok, err := HasAppPermission(testUserID, testAccountID, testAppUUID, CanSharedApp)
	if err != nil {
		t.Fatalf("HasAppPermission returned error: %v", err)
	}
	if !ok {
		t.Errorf("expected shared permission")
	}
}

func TestToggleAppPermission(t *testing.T) {
	perm, err := AppPermissionsByUserAccountAndApp(testUserID, testAccountID, testAppID)
	if err != nil {
		t.Fatalf("initial fetch error: %v", err)
	}
	if !perm.CanIndividual {
		t.Fatalf("expected CanIndividual true before toggle")
	}
	if err := ToggleAppPermission(testUserID, testAccountID, testAppID, ToggleAppIndividual); err != nil {
		t.Fatalf("toggle error: %v", err)
	}
	perm2, err := AppPermissionsByUserAccountAndApp(testUserID, testAccountID, testAppID)
	if err != nil {
		t.Fatalf("post toggle fetch error: %v", err)
	}
	if perm2.CanIndividual {
		t.Errorf("expected CanIndividual to be toggled off")
	}
}
