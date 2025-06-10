package models

import (
	"fmt"
	"log"
	"time"

	"github.com/jaredfolkins/letemcook/db"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

func NewUser() *User {
	return &User{
		Account: &Account{},
		Permissions: &Permissions{
			PermSystem:           &PermSystem{},
			PermissionsAccounts:  make([]*PermAccount, 0),
			PermissionsApps:      make([]*PermApp, 0),
			PermissionsCookbooks: make([]*PermCookbook, 0),
		},
	}
}

func CreateUserWithAccountID(u *User, account_id int64, tx *sqlx.Tx) (int64, error) {
	var query string
	query = `insert into users(username, email, hash, heckle) values(:username, :email, :hash, :heckle) returning id`
	var userID int64
	row := tx.QueryRowx(query, u.Username, u.Email, u.Hash, u.Heckle)
	err := row.Scan(&userID)
	if err != nil {
		log.Println("🔥 Failed to create User: ", err)
		return 0, err
	}
	u.ID = userID

	params := map[string]interface{}{
		"user_id":    userID,
		"account_id": account_id,
	}
	query = `
		insert into permissions_accounts(user_id, account_id, can_view_apps) 
		values(:user_id, :account_id, true) 
		returning id
	`

	_, err = tx.NamedExec(query, params)
	if err != nil {
		log.Println("🔥 Failed to join User to Account: ", err)
		return 0, err
	}

	return userID, nil
}

func CreateSuperUserWithAccountID(u *User, account_id int64, tx *sqlx.Tx) error {
	// Create the user entry
	userQuery := `insert into users(username, email, hash, heckle) values(:username, :email, :hash, :heckle) returning id`
	resp, err := tx.NamedExec(userQuery, u)
	if err != nil {
		log.Println("🔥 Failed to create User for SuperUser: ", err)
		return err
	}

	userID, err := resp.LastInsertId()
	if err != nil {
		log.Println("🔥 Failed to get User LastInsertId for SuperUser: ", err)
		return err
	}
	u.ID = userID // Set the ID on the user object

	// Grant full account permissions
	accountPermsParams := map[string]interface{}{
		"user_id":              userID,
		"account_id":           account_id,
		"can_administer":       true,
		"can_create_apps":      true,
		"can_view_apps":        true,
		"can_create_cookbooks": true,
		"can_view_cookbooks":   true,
		"is_owner":             true, // Typically superuser might be owner
	}
	accountPermsQuery := `
		insert into permissions_accounts(
			user_id, account_id, can_administer, can_create_apps, can_view_apps, 
			can_create_cookbooks, can_view_cookbooks, is_owner
		) 
		values(
			:user_id, :account_id, :can_administer, :can_create_apps, :can_view_apps, 
			:can_create_cookbooks, :can_view_cookbooks, :is_owner
		) 
		returning id
	`
	_, err = tx.NamedExec(accountPermsQuery, accountPermsParams)
	if err != nil {
		log.Println("🔥 Failed to grant Account permissions for SuperUser: ", err)
		return err
	}

	// Grant full App permissions
	appPermsParams := map[string]interface{}{
		"user_id":        userID,
		"can_administer": true,
		"is_owner":       true, // Superuser implies App owner/admin
	}
	appPermsQuery := `
		insert into permissions_system(user_id, can_administer, is_owner) 
		values(:user_id, :can_administer, :is_owner) 
		returning id
	`
	_, err = tx.NamedExec(appPermsQuery, appPermsParams)
	if err != nil {
		log.Println("🔥 Failed to grant App permissions for SuperUser: ", err)
		return err
	}

	return nil
}

func ByUsernameAndSquid(username string, squid string) (*User, error) {
	account, err := AccountBySquid(squid)
	if err != nil {
		return nil, err
	}

	return ByUsernameAndAccountID(username, account.ID)
}

func ByUsernameAndAccountID(username string, account_id int64) (*User, error) {
	u := NewUser()
	query1 := `
		select
		users.id,
		users.username,
		users.email,
		users.hash,
		users.created,
		users.updated,
		users.is_disabled,
		users.is_deleted,
		users.heckle
		from users
		join permissions_accounts on permissions_accounts.user_id = users.id
		join accounts on permissions_accounts.account_id = accounts.id
		where users.username = $1
		and accounts.id = $2
	`

	err := db.Db().Get(u, query1, username, account_id)
	if err != nil {
		return nil, err
	}

	query2 := `
		select
			accounts.id,
			accounts.name,
			accounts.squid,
			accounts.created,
			accounts.updated,
			accounts.is_deleted
		from accounts
		where accounts.id = $1
	`

	err = db.Db().Get(u.Account, query2, account_id)
	if err != nil {
		return u, err
	}

	query_app := `select id, user_id, created, updated, can_administer, is_owner from permissions_system where user_id = $1`
	err = db.Db().Get(u.Permissions.PermSystem, query_app, u.ID)
	if err != nil {
		log.Printf("Info/Warning: fetching permissions_system for user %d: %v", u.ID, err)
	}

	pa := []*PermAccount{}
	query3 := `select id, user_id, account_id, created, updated, can_administer, can_create_apps, can_view_apps, can_create_cookbooks, can_view_cookbooks, is_owner from permissions_accounts where user_id = $1 and account_id = $2`
	err = db.Db().Select(&pa, query3, u.ID, u.Account.ID)
	if err != nil {
		log.Printf("Error fetching permissions_accounts for user %d, account %d: %v", u.ID, u.Account.ID, err)
		return u, err
	}

	pc := []*PermApp{}
	query4 := `select id, user_id, account_id, cookbook_id, app_id, created, updated, can_shared, can_individual, can_administer, is_owner from permissions_apps where user_id = $1 and account_id = $2`
	err = db.Db().Select(&pc, query4, u.ID, u.Account.ID)
	if err != nil {
		log.Printf("Error fetching permissions_apps for user %d, account %d: %v", u.ID, u.Account.ID, err)
		return u, err
	}

	pcook := []*PermCookbook{}
	query5 := `select id, user_id, account_id, cookbook_id, created, updated, can_view, can_edit, is_owner from permissions_cookbooks where user_id = $1 and account_id = $2`
	err = db.Db().Select(&pcook, query5, u.ID, u.Account.ID)
	if err != nil {
		log.Printf("Error fetching permissions_cookbooks for user %d, account %d: %v", u.ID, u.Account.ID, err)
		return u, err
	}

	u.Permissions.PermissionsAccounts = pa
	u.Permissions.PermissionsApps = pc
	u.Permissions.PermissionsCookbooks = pcook

	return u, nil
}

func UserByIDAndAccountID(user_id, account_id int64) (*User, error) {
	u := NewUser()
	query1 := `
		select
		users.id,
		users.username,
		users.email,
		users.hash,
		users.created,
		users.updated,
		users.is_disabled,
		users.is_deleted,
		users.heckle
		from users
		join permissions_accounts on permissions_accounts.user_id = users.id
		join accounts on permissions_accounts.account_id = accounts.id
		where users.id = $1
		and accounts.id = $2
	`

	err := db.Db().Get(u, query1, user_id, account_id)
	if err != nil {
		return nil, err
	}

	query2 := `
		select
			accounts.id,
			accounts.name,
			accounts.squid,
			accounts.created,
			accounts.updated,
			accounts.is_deleted
		from accounts
		where accounts.id = $1
	`

	err = db.Db().Get(u.Account, query2, account_id)
	if err != nil {
		return u, err
	}

	query_app := `select id, user_id, created, updated, can_administer, is_owner from permissions_system where user_id = $1`
	err = db.Db().Get(u.Permissions.PermSystem, query_app, u.ID)
	if err != nil {
		log.Printf("Info/Warning: fetching permissions_system for user %d: %v", u.ID, err)
	}

	pa := []*PermAccount{}
	query3 := `select id, user_id, account_id, created, updated, can_administer, can_create_apps, can_view_apps, can_create_cookbooks, can_view_cookbooks, is_owner from permissions_accounts where user_id = $1 and account_id = $2`
	err = db.Db().Select(&pa, query3, u.ID, u.Account.ID)
	if err != nil {
		log.Printf("Error fetching permissions_accounts for user %d, account %d: %v", u.ID, u.Account.ID, err)
		return u, err
	}

	pc := []*PermApp{}
	query4 := `select id, user_id, account_id, cookbook_id, app_id, created, updated, can_shared, can_individual, can_administer, is_owner from permissions_apps where user_id = $1 and account_id = $2`
	err = db.Db().Select(&pc, query4, u.ID, u.Account.ID)
	if err != nil {
		log.Printf("Error fetching permissions_apps for user %d, account %d: %v", u.ID, u.Account.ID, err)
		return u, err
	}

	pcook := []*PermCookbook{}
	query5 := `select id, user_id, account_id, cookbook_id, created, updated, can_view, can_edit, is_owner from permissions_cookbooks where user_id = $1 and account_id = $2`
	err = db.Db().Select(&pcook, query5, u.ID, u.Account.ID)
	if err != nil {
		log.Printf("Error fetching permissions_cookbooks for user %d, account %d: %v", u.ID, u.Account.ID, err)
		return u, err
	}

	u.Permissions.PermissionsAccounts = pa
	u.Permissions.PermissionsApps = pc
	u.Permissions.PermissionsCookbooks = pcook

	return u, nil
}

type UserContext struct {
	ActingAs   *User
	LoggedInAs *User
}

func (uc *UserContext) IsAuthenticated() bool {
	if uc == nil || uc.LoggedInAs == nil {
		return false
	}

	if uc.LoggedInAs.ID == 0 {
		return false
	}
	return true
}

func (uc *UserContext) Username() string {
	if uc != nil && uc.LoggedInAs != nil && uc.ActingAs != nil {
		return uc.ActingAs.Username
	}
	return "Username() Error!"
}

func NewUserCtxByUsernameAndAccountID(username string, account_id int64) (*UserContext, error) {
	u, err := ByUsernameAndAccountID(username, account_id)
	if err != nil {
		return nil, err
	}

	uctx := &UserContext{
		ActingAs:   u,
		LoggedInAs: u,
	}

	return uctx, nil
}

func CountUsersByAccountID(accountID int64) (int, error) {
	var count int
	query := `
		SELECT COUNT(users.id)
		FROM users
		JOIN permissions_accounts ON permissions_accounts.user_id = users.id
		WHERE permissions_accounts.account_id = $1
		AND users.is_deleted = false
		AND users.is_disabled = false
	`
	err := db.Db().Get(&count, query, accountID)
	if err != nil {
		log.Printf("Error counting users for account %d: %v", accountID, err)
		return 0, err
	}
	return count, nil
}

func GetUsersByAccountID(accountID int64, page, limit int) ([]User, error) {
	users := []User{}

	offset := (page - 1) * limit

	query := `
		SELECT
			users.id,
			users.username,
			users.email,
			users.hash,
			users.created,
			users.updated,
			users.is_disabled,
			users.is_deleted,
			users.heckle
		FROM users
		JOIN permissions_accounts ON permissions_accounts.user_id = users.id
		WHERE permissions_accounts.account_id = $1
		AND users.is_deleted = false
		AND users.is_disabled = false
		ORDER BY users.username ASC
		LIMIT $2 OFFSET $3
	`

	err := db.Db().Select(&users, query, accountID, limit, offset)
	if err != nil {
		log.Printf("Error fetching paginated users for account %d (page %d, limit %d): %v", accountID, page, limit, err)
		return nil, err
	}

	return users, nil
}

// CreateUser creates a new user with a hashed password and associates them
// with the given account ID, granting default permissions.
func CreateUser(username, email, password string, accountID int64) (*User, error) {
	// Hash the password
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("🔥 Failed to hash password for user %s: %v", username, err)
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	hash := string(hashBytes)

	// Use a transaction
	tx, err := db.Db().Beginx()
	if err != nil {
		log.Printf("🔥 Failed to begin transaction for user creation: %v", err)
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	// Defer rollback in case of error, commit on success
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // Re-panic after rollback
		} else if err != nil {
			tx.Rollback() // Rollback on error
		} else {
			err = tx.Commit() // Commit on success
			if err != nil {
				log.Printf("🔥 Failed to commit transaction for user creation: %v", err)
			}
		}
	}()

	// Insert the user - set default heckle to false here
	userQuery := `INSERT INTO users (username, email, hash, heckle) VALUES ($1, $2, $3, $4) RETURNING id`
	var userID int64
	err = tx.QueryRowx(userQuery, username, email, hash, false).Scan(&userID)
	if err != nil {
		log.Printf("🔥 Failed to create user %s: %v", username, err)
		// Error is handled by defer
		return nil, fmt.Errorf("failed to insert user: %w", err)
	}

	// Grant default account permissions (can view apps)
	permParams := map[string]interface{}{
		"user_id":    userID,
		"account_id": accountID,
	}
	permQuery := `INSERT INTO permissions_accounts (user_id, account_id, can_view_apps) VALUES (:user_id, :account_id, true)`
	_, err = tx.NamedExec(permQuery, permParams)
	if err != nil {
		log.Printf("🔥 Failed to grant default permissions to user %d in account %d: %v", userID, accountID, err)
		// Error is handled by defer
		return nil, fmt.Errorf("failed to grant permissions: %w", err)
	}

	// Return the newly created user (or at least the ID)
	newUser := &User{
		ID:       userID,
		Username: username,
		Email:    email,
		Heckle:   false, // Set default in struct as well
		// Hash is not typically returned unless needed
		Account: &Account{ID: accountID}, // Basic account info
		// Permissions are minimal, could be loaded separately if needed
	}

	log.Printf("✅ Successfully created user %d (%s) in account %d", userID, username, accountID)
	return newUser, nil // Error is nil if commit succeeds
}

// UpdateUserPassword updates a user's password hash in the database
func UpdateUserPassword(userID int64, hashedPassword string) error {
	query := `UPDATE users SET hash = ?, updated = datetime('now') WHERE id = ?`
	_, err := db.Db().Exec(query, hashedPassword, userID)
	if err != nil {
		log.Printf("Error updating password for user %d: %v", userID, err)
		return err
	}
	return nil
}

// UpdateUserHeckle updates a user's heckle setting in the database
func UpdateUserHeckle(userID int64, heckle bool) error {
	query := `UPDATE users SET heckle = ?, updated = datetime('now') WHERE id = ?`
	_, err := db.Db().Exec(query, heckle, userID)
	if err != nil {
		log.Printf("Error updating heckle setting for user %d: %v", userID, err)
		return err
	}
	log.Printf("Successfully updated heckle for user %d to %v", userID, heckle)
	return nil
}

type User struct {
	Created    time.Time `db:"created"`
	Updated    time.Time `db:"updated"`
	ID         int64     `db:"id"`
	IsDisabled bool      `db:"is_disabled"`
	IsDeleted  bool      `db:"is_deleted"`
	Heckle     bool      `db:"heckle"`
	Hash       string    `db:"hash"`
	Username   string    `db:"username"`
	Email      string    `db:"email"`
	Password   string    `db:"-"`

	Account     *Account     `db:"-"`
	Permissions *Permissions `db:"-"`
}

func SearchAllUsers(search string, limit int) ([]User, error) {
	users := []User{}
	query := `
        SELECT
            users.id,
            users.username,
            users.email,
            accounts.id AS "account.id",
            accounts.name AS "account.name",
            accounts.squid AS "account.squid"
        FROM users
        JOIN permissions_accounts ON permissions_accounts.user_id = users.id
        JOIN accounts ON permissions_accounts.account_id = accounts.id
        WHERE users.is_deleted = false
          AND users.is_disabled = false
          AND (users.username LIKE $1 OR users.email LIKE $1)
        GROUP BY users.id, accounts.id
        ORDER BY users.username
        LIMIT $2`
	search = "%" + search + "%"
	err := db.Db().Select(&users, query, search, limit)
	if err != nil {
		return nil, err
	}
	// Attach permission bundle for each user
	for i := range users {
		u := &users[i]
		full, err := UserByIDAndAccountID(u.ID, u.Account.ID)
		if err == nil {
			users[i] = *full
		}
	}
	return users, nil
}
