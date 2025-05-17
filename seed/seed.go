package seed

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/jmoiron/sqlx"
	"github.com/sqids/sqids-go"
)

const (
	seedPassword = "asdfasdfasdf"

	accAlphaName        = "Account Alpha"
	accAlphaUsers       = 15
	accAlphaCookbooks   = 25
	accAlphaapps        = 25
	accAlphaOwnerUser   = "alpha-owner"
	accAlphaOwnerEmail  = "alpha-owner@example.com"
	accAlphaAdmin2User  = "alpha-admin-2"
	accAlphaAdmin2Email = "alpha-admin-2@example.com"
	appAdminUser        = accAlphaOwnerUser

	accBravoName        = "Account Bravo"
	accBravoUsers       = 15
	accBravoCookbooks   = 25
	accBravoapps        = 25
	accBravoOwnerUser   = "bravo-owner"
	accBravoOwnerEmail  = "bravo-owner@example.com"
	accBravoAdmin2User  = "bravo-admin-2"
	accBravoAdmin2Email = "bravo-admin-2@example.com"
)

type appPermissionSpec struct {
	appIndex      int
	CanShared     bool
	CanIndividual bool
	CanAdmin      bool
	IsOwner       bool // Should generally be false unless assigning to the designated owner
}

type SeedUser struct {
	UsernameSuffix string
	EmailSuffix    string
	IsAccountAdmin bool // For account-level perms
	appPermSpecs   []appPermissionSpec
}

func SeedDatabaseIfDev(db *sqlx.DB) {
    lemcEnv := os.Getenv("LEMC_ENV")
    envLower := strings.ToLower(lemcEnv)
    isDev := envLower == "development" || envLower == "dev" || envLower == "test"

	if isDev {
		log.Println("🌱 Development environment detected. Seeding database...")
		totalUsers := accAlphaUsers + accBravoUsers
		totalCookbooks := accAlphaCookbooks + accBravoCookbooks
		totalapps := accAlphaapps + accBravoapps
		log.Printf("   Target: 2 accounts, %d users, %d cookbooks, %d apps", totalUsers, totalCookbooks, totalapps)

		accAlphaID, err := seedAccount(db, accAlphaName)
		if err != nil {
			log.Printf("🔥 Error seeding account '%s': %v. Halting seed.", accAlphaName, err)
			return
		}
		if accAlphaID == 0 {
			accAlphaID, err = getAccountIDByName(db, accAlphaName)
			if err != nil {
				log.Printf("🔥 Error fetching existing account ID for '%s': %v. Halting seed.", accAlphaName, err)
				return
			}
		}
		log.Printf("--- Seeding for Account Alpha (ID: %d) ---", accAlphaID)
		alphaStaticUsers := []SeedUser{
			{UsernameSuffix: "-viewer", EmailSuffix: "-viewer", IsAccountAdmin: false, appPermSpecs: []appPermissionSpec{
				{appIndex: 0, CanShared: true, CanIndividual: true, CanAdmin: false, IsOwner: false}, // View only on app 0
			}},
			{UsernameSuffix: "-editor", EmailSuffix: "-editor", IsAccountAdmin: false, appPermSpecs: []appPermissionSpec{
				{appIndex: 1, CanShared: true, CanIndividual: true, CanAdmin: true, IsOwner: false}, // Admin on app 2
			}},
			{UsernameSuffix: "-limited", EmailSuffix: "-limited", IsAccountAdmin: false, appPermSpecs: []appPermissionSpec{
				{appIndex: 0, CanShared: false, CanIndividual: true, CanAdmin: false, IsOwner: false}, // Individual only on app 0
			}},
		}
		seedDataForAccount(db, accAlphaID, accAlphaOwnerUser, accAlphaOwnerEmail, accAlphaAdmin2User, accAlphaAdmin2Email, "alpha", alphaStaticUsers, accAlphaCookbooks, accAlphaapps, true)

		accBravoID, err := seedAccount(db, accBravoName)
		if err != nil {
			log.Printf("🔥 Error seeding account '%s': %v. Halting seed.", accBravoName, err)
			return
		}
		if accBravoID == 0 {
			accBravoID, err = getAccountIDByName(db, accBravoName)
			if err != nil {
				log.Printf("🔥 Error fetching existing account ID for '%s': %v. Halting seed.", accBravoName, err)
				return
			}
		}
		log.Printf("--- Seeding for Account Bravo (ID: %d) ---", accBravoID)
		bravoStaticUsers := []SeedUser{
			{UsernameSuffix: "-main", EmailSuffix: "-main", IsAccountAdmin: false, appPermSpecs: []appPermissionSpec{
				{appIndex: 0, CanShared: true, CanIndividual: true, CanAdmin: true, IsOwner: false},  // Admin on app 0
				{appIndex: 1, CanShared: true, CanIndividual: true, CanAdmin: false, IsOwner: false}, // View on app 1
			}},
			{UsernameSuffix: "-extra", EmailSuffix: "-extra", IsAccountAdmin: false, appPermSpecs: []appPermissionSpec{
				{appIndex: 2, CanShared: true, CanIndividual: false, CanAdmin: false, IsOwner: false}, // Shared only on app 2
			}},
		}
		seedDataForAccount(db, accBravoID, accBravoOwnerUser, accBravoOwnerEmail, accBravoAdmin2User, accBravoAdmin2Email, "bravo", bravoStaticUsers, accBravoCookbooks, accBravoapps, false)

		log.Println("🌱 Database seeding complete.")

		log.Println("--- Seeded Admin User Credentials ---")
		log.Printf("  App Admin / Acc Alpha Owner: username='%s' password='%s'", accAlphaOwnerUser, seedPassword)
		log.Printf("  Acc Alpha Admin 2:         username='%s' password='%s'", accAlphaAdmin2User, seedPassword)
		log.Printf("  Acc Bravo Owner:           username='%s' password='%s'", accBravoOwnerUser, seedPassword)
		log.Printf("  Acc Bravo Admin 2:         username='%s' password='%s'", accBravoAdmin2User, seedPassword)
		log.Println("-------------------------------------")
	}
}

func seedDataForAccount(db *sqlx.DB, accountID int64, ownerUser, ownerEmail, admin2User, admin2Email, prefix string, staticUsers []SeedUser, numCookbooks, numapps int, designateOwnerAsAppAdmin bool) {
	log.Printf("--- Seeding data for account ID %d (%s) ---", accountID, prefix)
	userappPermSpecs := make(map[int64][]appPermissionSpec)

	ownerID, err := seedUser(db, accountID, ownerUser, ownerEmail, seedPassword, true, true, designateOwnerAsAppAdmin)
	if err != nil || ownerID == 0 {
		log.Printf("🔥 Failed to seed or find owner user '%s' for account %d: %v. Halting account seed.", ownerUser, accountID, err)
		return
	}

	admin2ID, err := seedUser(db, accountID, admin2User, admin2Email, seedPassword, false, true, false)
	if err != nil {
		log.Printf("🔥 Error seeding second admin user '%s' for account %d: %v", admin2User, accountID, err)
	} else {
		adminPerms := []appPermissionSpec{
			{appIndex: 0, CanShared: true, CanIndividual: true, CanAdmin: true, IsOwner: false},
			{appIndex: 1, CanShared: true, CanIndividual: true, CanAdmin: true, IsOwner: false},
		}
		userappPermSpecs[admin2ID] = adminPerms
		log.Printf("   👤 Admin user '%s' (ID: %d) seeded. Permissions specs stored.", admin2User, admin2ID)
	}

	log.Printf("--- Seeding %d statically defined users for account %d (%s) ---", len(staticUsers), accountID, prefix)
	for _, userData := range staticUsers {
		username := prefix + userData.UsernameSuffix
		email := prefix + userData.EmailSuffix + "@example.com"
		userID, err := seedUser(db, accountID, username, email, seedPassword, false, userData.IsAccountAdmin, false) // Not owner, not app admin
		if err != nil {
			log.Printf("🔥 Error seeding static user %s for account %d: %v", username, accountID, err)
			continue
		}
		if userID != 0 {
			userappPermSpecs[userID] = userData.appPermSpecs
			log.Printf("   👤 Static user '%s' (ID: %d) seeded. %d permission specs stored.", username, userID, len(userData.appPermSpecs))
		}
	}

	log.Printf("--- Seeding %d cookbooks for account %d (%s) ---", numCookbooks, accountID, prefix)
	cookbookIDs := make([]int64, 0, numCookbooks)
	for i := 0; i < numCookbooks; i++ {
		cookbookName := fmt.Sprintf("%s Cookbook %d", strings.Title(prefix), i+1)
		cookbookDesc := fmt.Sprintf("Description for %s cookbook #%d.", strings.Title(prefix), i+1)
		cookbookOwnerID := ownerID

		cookbookID, err := seedCookbook(db, accountID, cookbookOwnerID, cookbookName, cookbookDesc)
		if err != nil {
			log.Printf("🔥 Error seeding cookbook %s for account %d owner %d: %v", cookbookName, accountID, cookbookOwnerID, err)
		} else if cookbookID != 0 {
			cookbookIDs = append(cookbookIDs, cookbookID)
		} else {
			var existingID int64
			err := db.QueryRow("SELECT id FROM cookbooks WHERE name = ? AND account_id = ?", cookbookName, accountID).Scan(&existingID)
			if err != nil {
				log.Printf("🔥 Error fetching existing cookbook ID for %s (Account %d): %v", cookbookName, accountID, err)
			} else {
				log.Printf("   Found existing cookbook ID: %d for %s", existingID, cookbookName)
				cookbookIDs = append(cookbookIDs, existingID)
				errPerm := seedCookbookPermissions(db, accountID, cookbookOwnerID, existingID, true, true, true)
				if errPerm != nil {
					log.Printf("🔥 Error ensuring cookbook permissions for owner %d on existing cookbook %d: %v", cookbookOwnerID, existingID, errPerm)
				}
			}
		}
	}

	log.Printf("--- Seeding %d apps for account %d (%s) ---", numapps, accountID, prefix)
	type appInfo struct {
		ID         int64
		CookbookID int64
		Name       string
	}
	appInfos := make([]appInfo, 0, numapps)
	if len(cookbookIDs) == 0 {
		log.Printf("⚠️ Skipping app seeding for account %d because no cookbook IDs were successfully created or found.", accountID)
	} else {
		for i := 0; i < numapps; i++ {
			appName := fmt.Sprintf("%s app %d", strings.Title(prefix), i+1)
			appDesc := fmt.Sprintf("Description for %s app #%d.", strings.Title(prefix), i+1)
			linkedCookbookID := cookbookIDs[i%len(cookbookIDs)]
			appOwnerID := ownerID

			appID, err := seedapp(db, accountID, appOwnerID, linkedCookbookID, appName, appDesc)
			if err != nil {
				log.Printf("🔥 Error seeding app %s for account %d owner %d: %v", appName, accountID, appOwnerID, err)
			} else if appID != 0 {
				appInfos = append(appInfos, appInfo{ID: appID, CookbookID: linkedCookbookID, Name: appName})
			} else {
				var existingID int64
				err := db.QueryRow("SELECT id FROM apps WHERE name = ? AND account_id = ?", appName, accountID).Scan(&existingID)
				if err != nil {
					log.Printf("🔥 Error fetching existing app ID for %s (Account %d): %v", appName, accountID, err)
				} else {
					log.Printf("   Found existing app ID: %d for %s. Ensuring permissions...", existingID, appName)
					appInfos = append(appInfos, appInfo{ID: existingID, CookbookID: linkedCookbookID, Name: appName})
					errPerm := seedappPermissions(db, accountID, appOwnerID, existingID, linkedCookbookID, true, true, true, true)
					if errPerm != nil {
						log.Printf("🔥 Error ensuring app permissions for owner %d on existing app %d: %v", appOwnerID, existingID, errPerm)
					}
				}
			}
		}
	}

	log.Printf("--- Seeding specific app permissions for %d users on %d apps (Account %d - %s) ---", len(userappPermSpecs), len(appInfos), accountID, prefix)
	if len(appInfos) == 0 {
		log.Printf("   ⚠️ Skipping specific app permission seeding: No apps were created/found.")
	} else {
		for userID, specs := range userappPermSpecs {
			if userID == 0 {
				continue
			} // Skip if user seeding failed
			log.Printf("   Processing permissions for User ID: %d", userID)
			for _, spec := range specs {
				if spec.appIndex < 0 || spec.appIndex >= len(appInfos) {
					log.Printf("      ⚠️ Skipping spec: Invalid appIndex %d (Num apps: %d)", spec.appIndex, len(appInfos))
					continue
				}
				targetapp := appInfos[spec.appIndex]
				log.Printf("      Applying permissions to app '%s' (ID: %d, Index: %d)", targetapp.Name, targetapp.ID, spec.appIndex)

				err = seedappPermissions(db, accountID, userID, targetapp.ID, targetapp.CookbookID, spec.CanShared, spec.CanIndividual, spec.CanAdmin, spec.IsOwner)
				if err != nil {
					log.Printf("      🔥 Error seeding app permissions for User %d on app %d (%s): %v", userID, targetapp.ID, targetapp.Name, err)
				} else {
					log.Printf("      🔑 Permissions set for User %d on app %d (%s): Shared=%t, Indiv=%t, Admin=%t, Owner=%t", userID, targetapp.ID, targetapp.Name, spec.CanShared, spec.CanIndividual, spec.CanAdmin, spec.IsOwner)
				}
			}
		}
	}

	log.Printf("--- Ensuring owner (ID: %d) permissions on all %d apps ---", ownerID, len(appInfos))
	for _, app := range appInfos {
		errPerm := seedappPermissions(db, accountID, ownerID, app.ID, app.CookbookID, true, true, true, true) // Owner gets full perms
		if errPerm != nil {
			log.Printf("   🔥 Error ensuring owner permissions for owner %d on app %d (%s): %v", ownerID, app.ID, app.Name, errPerm)
		}
	}

	log.Printf("--- Finished seeding data for account ID %d (%s) ---", accountID, prefix)
}

func seedUser(db *sqlx.DB, accountID int64, username, email, password string, isAccountOwner, isAccountAdmin, isAppAdmin bool) (int64, error) {
	log.Printf("Seeding user: %s for account %d...", username, accountID)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("🔥 Error hashing password for user %s: %v", username, err)
		return 0, err
	}

	var userID int64
	userQuery := "INSERT OR IGNORE INTO users (username, email, hash) VALUES (?, ?, ?) RETURNING id"
	err = db.QueryRow(userQuery, username, email, string(hashedPassword)).Scan(&userID)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		err = nil
		err = db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
		if err != nil {
			log.Printf("🔥 Error fetching existing user ID for %s: %v", username, err)
			return 0, err
		}
		log.Printf("   👤 User '%s' already exists (ID: %d). Ensuring permissions...", username, userID)
	} else if err != nil {
		log.Printf("🔥 Error inserting user %s: %v", username, err)
		return 0, err
	} else {
		log.Printf("   👤 User '%s' created successfully (ID: %d)", username, userID)
	}

	accPermCanAdmin := isAccountOwner || isAccountAdmin
	accPermCanCreate := isAccountOwner || isAccountAdmin
	accPermCanView := true
	accPermIsOwner := isAccountOwner

	permQuery := `INSERT INTO permissions_accounts
		(user_id, account_id, can_administer, can_create_apps, can_view_apps, can_create_cookbooks, can_view_cookbooks, is_owner)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(account_id, user_id) DO UPDATE SET
		can_administer = excluded.can_administer,
		can_create_apps = excluded.can_create_apps,
		can_view_apps = excluded.can_view_apps,
		can_create_cookbooks = excluded.can_create_cookbooks,
		can_view_cookbooks = excluded.can_view_cookbooks,
		is_owner = excluded.is_owner`

	_, err = db.Exec(permQuery, userID, accountID, accPermCanAdmin, accPermCanCreate, accPermCanView, accPermCanCreate, accPermCanView, accPermIsOwner)
	if err != nil {
		log.Printf("🔥 Error setting/updating account permissions for user ID %d on account ID %d: %v", userID, accountID, err)
	} else {
		log.Printf("   🔑 Account permissions set/verified for user ID %d (Owner: %t, Admin: %t).", userID, accPermIsOwner, accPermCanAdmin)
	}

	if isAppAdmin {
		err = seedAppPermissions(db, userID, true, isAccountOwner)
		if err != nil {
			log.Printf("🔥 Error setting/updating app permissions for designated app admin user ID %d: %v", userID, err)
		}
	}

	return userID, nil
}

func seedAppPermissions(db *sqlx.DB, userID int64, canAdminister, isOwner bool) error {
	log.Printf("   Seeding app permissions for user ID: %d...", userID)
	permQuery := `INSERT INTO permissions_system
        (user_id, can_administer, is_owner)
        VALUES (?, ?, ?)
        ON CONFLICT(user_id) DO UPDATE SET
        can_administer = excluded.can_administer,
        is_owner = excluded.is_owner`
	_, err := db.Exec(permQuery, userID, canAdminister, isOwner)
	if err != nil {
		log.Printf("🔥 Error setting/updating app permissions for user ID %d: %v", userID, err)
		return err
	} else {
		log.Printf("   🔑 App permissions set/verified for user ID %d.", userID)
	}
	return nil
}

func seedAccount(db *sqlx.DB, accountName string) (int64, error) {
	log.Printf("   Seeding account: %s...", accountName)
	var accountID int64
	tempSquid := fmt.Sprintf("temp_%s", uuid.New().String())
	insertQuery := "INSERT INTO accounts (name, squid) VALUES (?, ?) RETURNING id"

	err := db.QueryRow(insertQuery, accountName, tempSquid).Scan(&accountID)

	if err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed: accounts.name") {
		log.Printf("   🏢 Account '%s' already exists. Fetching ID...", accountName)
		fetchQuery := "SELECT id FROM accounts WHERE name = ?"
		errFetch := db.QueryRow(fetchQuery, accountName).Scan(&accountID)
		if errFetch != nil {
			log.Printf("🔥 Error fetching existing account ID for %s: %v", accountName, errFetch)
			return 0, errFetch
		}
		log.Printf("   🏢 Found existing account ID: %d. Ensuring real squid...", accountID)
	} else if err != nil {
		log.Printf("🔥 Error inserting account %s (using temp squid %s): %v", accountName, tempSquid, err)
		return 0, err
	} else {
		log.Printf("   🏢 Account '%s' created successfully (ID: %d) with temp squid. Setting real squid...", accountName, accountID)
	}

	s, err := sqids.New(
		sqids.Options{
			MinLength: 4,
			Alphabet:  os.Getenv("LEMC_SQUID_ALPHABET"),
		})
	if err != nil {
		log.Printf("🔥 Error creating sqids generator: %v", err)
		return accountID, err
	}
	realSquid, err := s.Encode([]uint64{uint64(accountID)})
	if err != nil {
		log.Printf("🔥 Error encoding real squid for account ID %d: %v", accountID, err)
		return accountID, err
	}

	updateQuery := "UPDATE accounts SET squid = ? WHERE id = ?"
	result, err := db.Exec(updateQuery, realSquid, accountID)
	if err != nil {
		log.Printf("🔥 Error updating account %d with real squid '%s': %v", accountID, realSquid, err)
		return accountID, err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("   🦑 Real squid '%s' set for account ID %d.", realSquid, accountID)
	} else {
		var existingSquid sql.NullString
		checkQuery := "SELECT squid FROM accounts WHERE id = ?"
		db.QueryRow(checkQuery, accountID).Scan(&existingSquid)
		if existingSquid.Valid && existingSquid.String == realSquid {
			log.Printf("   🦑 Real squid '%s' was already correctly set for account ID %d.", realSquid, accountID)
		} else {
			log.Printf("   🦑 Real squid update for account ID %d did not affect rows. Current='%s', Target='%s'", accountID, existingSquid.String, realSquid)
		}
	}

	return accountID, nil
}

func getAccountIDByName(db *sqlx.DB, accountName string) (int64, error) {
	var accountID int64
	err := db.QueryRow("SELECT id FROM accounts WHERE name = ?", accountName).Scan(&accountID)
	if err != nil {
		return 0, err
	}
	return accountID, nil
}

func seedCookbook(db *sqlx.DB, accountID, ownerID int64, cookbookName, cookbookDesc string) (int64, error) {
	log.Printf("   Seeding cookbook: %s...", cookbookName)
	var cookbookID int64
	cbQuery := `INSERT INTO cookbooks (account_id, owner_id, name, description, uuid, api_key, yaml_shared, yaml_individual)
              VALUES (?, ?, ?, ?, ?, ?, ?, ?)
              RETURNING id`

	uuidVal := uuid.New().String()
	apiKey := uuid.New().String()
	defaultYamlShared := "shared_cookbook_key: shared_value"
	defaultYamlIndividual := "individual_cookbook_key: individual_value"

	err := db.QueryRow(cbQuery, accountID, ownerID, cookbookName, cookbookDesc, uuidVal, apiKey, defaultYamlShared, defaultYamlIndividual).Scan(&cookbookID)

	if err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed: cookbooks.name") {
		log.Printf("      📖 Cookbook '%s' already exists for this account. Fetching ID...", cookbookName)
		fetchQuery := "SELECT id FROM cookbooks WHERE name = ? AND account_id = ?"
		errFetch := db.QueryRow(fetchQuery, cookbookName, accountID).Scan(&cookbookID)
		if errFetch != nil {
			log.Printf("🔥 Error fetching existing cookbook ID for %s: %v", cookbookName, errFetch)
			return 0, errFetch
		}
		log.Printf("      📖 Found existing cookbook ID: %d.", cookbookID)
		return cookbookID, nil
	} else if err != nil {
		log.Printf("🔥 Error inserting cookbook %s: %v", cookbookName, err)
		return 0, err
	} else {
		log.Printf("      📖 Cookbook '%s' created successfully (ID: %d)", cookbookName, cookbookID)
		errPerm := seedCookbookPermissions(db, accountID, ownerID, cookbookID, true, true, true)
		if errPerm != nil {
			log.Printf("🔥 Error setting cookbook permissions for owner %d on cookbook %d: %v", ownerID, cookbookID, errPerm)
		} else {
			log.Printf("      🔑 Cookbook permissions set for owner ID %d on cookbook ID %d.", ownerID, cookbookID)
		}
		return cookbookID, nil
	}
}

func seedCookbookPermissions(db *sqlx.DB, accountID, userID, cookbookID int64, canView, canEdit, isOwner bool) error {
	permQuery := `INSERT INTO permissions_cookbooks
        (account_id, user_id, cookbook_id, can_view, can_edit, is_owner)
        VALUES (?, ?, ?, ?, ?, ?)
        ON CONFLICT(account_id, user_id, cookbook_id) DO UPDATE SET
        can_view = excluded.can_view,
        can_edit = excluded.can_edit,
        is_owner = excluded.is_owner`
	_, err := db.Exec(permQuery, accountID, userID, cookbookID, canView, canEdit, isOwner)
	return err
}

func seedapp(db *sqlx.DB, accountID, ownerID, cookbookID int64, appName, appDesc string) (int64, error) {
	log.Printf("   Seeding app: %s (Owner: %d, Cookbook: %d)...", appName, ownerID, cookbookID)
	crsQuery := `INSERT INTO apps (account_id, owner_id, cookbook_id, name, description, uuid, api_key, yaml_shared, yaml_individual)
               VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
               RETURNING id`

	uuidVal := uuid.New().String()
	apiKey := uuid.New().String()
	defaultYamlShared := "shared_app_key: shared_value"
	defaultYamlIndividual := "individual_app_key: individual_value"

	var appID int64
	err := db.QueryRow(crsQuery, accountID, ownerID, cookbookID, appName, appDesc, uuidVal, apiKey, defaultYamlShared, defaultYamlIndividual).Scan(&appID)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: apps.name") {
			log.Printf("         📚 app '%s' already exists for this account. Skipping creation.", appName)
			fetchQuery := "SELECT id FROM apps WHERE name = ? AND account_id = ?"
			errFetch := db.QueryRow(fetchQuery, appName, accountID).Scan(&appID)
			if errFetch != nil {
				log.Printf("🔥 Error fetching existing app ID for %s: %v", appName, errFetch)
				return 0, nil
			}
			log.Printf("         📚 Found existing app ID: %d. Will ensure permissions later...", appID)
			return appID, nil
		} else if strings.Contains(err.Error(), "UNIQUE constraint failed: apps.uuid") {
			log.Printf("🔥 UUID collision for app %s. This should be extremely rare.", appName)
		}
		log.Printf("🔥 Error inserting app %s: %v", appName, err)
		return 0, err
	} else {
		log.Printf("         📚 app '%s' created successfully (ID: %d). Will set owner permissions later...", appName, appID)
		return appID, nil
	}
}

func seedappPermissions(db *sqlx.DB, accountID, userID, appID, cookbookID int64, canShared, canIndividual, canAdmin, isOwner bool) error {
	permQuery := `INSERT INTO permissions_apps
        (account_id, user_id, app_id, cookbook_id, can_shared, can_individual, can_administer, is_owner, api_key)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
        ON CONFLICT(account_id, user_id, app_id) DO UPDATE SET
        can_shared = excluded.can_shared,
        can_individual = excluded.can_individual,
        can_administer = excluded.can_administer,
        is_owner = excluded.is_owner`
	apiKey := uuid.New().String()
	_, err := db.Exec(permQuery, accountID, userID, appID, cookbookID, canShared, canIndividual, canAdmin, isOwner, apiKey)
	return err
}
