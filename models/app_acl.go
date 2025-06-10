package models

import (
	"database/sql"
	"log"
	"sort"
	"strings"

	"github.com/jaredfolkins/letemcook/db"
)

type AppAcl struct {
	UserID        int64  `db:"user_id"`
	Username      string `db:"username"`
	Email         string `db:"email"`
	CanShared     bool   `db:"can_shared"`
	CanIndividual bool   `db:"can_individual"`
	CanAdmin      bool   `db:"can_administer"`
	IsOwner       bool   `db:"is_owner"`
}

func AppAclsUsers(accountID int64, appID int64) ([]AppAcl, error) {
	var ownerID int64
	ownerQuery := "SELECT owner_id FROM Apps WHERE id = ? AND account_id = ?"
	err := db.Db().Get(&ownerID, ownerQuery, appID, accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("AppAclsUsers: App not found or doesn't belong to account (appID: %d, AccountID: %d)", appID, accountID)
			return nil, err // Or return a more specific "not found" error
		}
		log.Printf("AppAclsUsers: Error fetching App owner (appID: %d, AccountID: %d): %v", appID, accountID, err)
		return nil, err
	}

	query := `
		SELECT
			u.id AS user_id,
			u.username,
			u.email,
			cup.can_shared,
			cup.can_individual,
			cup.can_administer,
			cup.is_owner
		FROM
			users u
		JOIN
			permissions_apps cup ON u.id = cup.user_id
		WHERE
			cup.account_id = ?
			AND cup.app_id = ?
	`

	var acls []AppAcl
	err = db.Db().Select(&acls, query, accountID, appID)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("AppAclsUsers: Error fetching App permissions (appID: %d, AccountID: %d): %v", appID, accountID, err)
		return nil, err
	}

	aclMap := make(map[int64]AppAcl)
	isOwnerInList := false

	for _, acl := range acls {
		acl.IsOwner = (acl.UserID == ownerID)
		if acl.IsOwner {
			acl.CanIndividual = true
			acl.CanShared = true
			acl.CanAdmin = true
			isOwnerInList = true
		}
		aclMap[acl.UserID] = acl
	}

	if !isOwnerInList && ownerID != 0 {
		var ownerAcl AppAcl
		ownerInfoQuery := `SELECT u.id AS user_id, u.username, u.email
                        FROM users u
                        JOIN permissions_accounts pa ON pa.user_id = u.id
                        WHERE u.id = ? AND pa.account_id = ?`
		err = db.Db().Get(&ownerAcl, ownerInfoQuery, ownerID, accountID)
		if err != nil {
			log.Printf("AppAclsUsers: Error fetching owner details (UserID: %d, AccountID: %d): %v", ownerID, accountID, err)
		} else {
			ownerAcl.CanIndividual = true
			ownerAcl.CanShared = true
			ownerAcl.CanAdmin = true
			ownerAcl.IsOwner = true
			aclMap[ownerAcl.UserID] = ownerAcl
		}
	}

	finalAcls := make([]AppAcl, 0, len(aclMap))

	for _, acl := range aclMap {
		if acl.IsOwner {
			finalAcls = append(finalAcls, acl)
		}
	}

	nonOwners := make([]AppAcl, 0)
	for _, acl := range aclMap {
		if !acl.IsOwner {
			nonOwners = append(nonOwners, acl)
		}
	}

	sort.Slice(nonOwners, func(i, j int) bool {
		return strings.ToLower(nonOwners[i].Username) < strings.ToLower(nonOwners[j].Username)
	})

	finalAcls = append(finalAcls, nonOwners...)

	return finalAcls, nil
}
