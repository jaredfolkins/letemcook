package models

import "testing"

func TestAppAclsUsersOwnerFallback(t *testing.T) {
	// remove owner permission so function must fetch owner info separately
	if _, err := testDB.Exec(`DELETE FROM permissions_apps WHERE user_id=? AND app_id=?`, testUserID, testAppID); err != nil {
		t.Fatalf("cleanup: %v", err)
	}
	defer testDB.Exec(`INSERT INTO permissions_apps (user_id, account_id, app_id, cookbook_id, can_shared, can_individual, can_administer, is_owner, api_key) VALUES (?, ?, ?, ?, 1,1,1,1, 'permkey')`, testUserID, testAccountID, testAppID, testCookbookID)

	acls, err := AppAclsUsers(testAccountID, testAppID)
	if err != nil {
		t.Fatalf("AppAclsUsers returned error: %v", err)
	}
	if len(acls) == 0 || acls[0].UserID != testUserID || !acls[0].IsOwner {
		t.Fatalf("unexpected ACLs: %#v", acls)
	}
}
