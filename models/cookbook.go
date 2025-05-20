package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jaredfolkins/letemcook/db"
	"github.com/jmoiron/sqlx"
	"gopkg.in/yaml.v3"
)

type Cookbook struct {
	Created        time.Time `db:"created" json:"created"`
	Updated        time.Time `db:"updated" json:"updated"`
	ID             int64     `db:"id" json:"id"`
	AccountID      int64     `db:"account_id" json:"account_id"`
	OwnerID        int64     `db:"owner_id" json:"owner_id"`
	UUID           string    `db:"uuid" json:"uuid"`
	Name           string    `db:"name" json:"name"`
	Description    string    `db:"description" json:"description"`
	YamlShared     string    `db:"yaml_shared" json:"yaml_shared"`
	YamlIndividual string    `db:"yaml_individual" json:"yaml_individual"`
	ApiKey         string    `db:"api_key" json:"api_key"`
	IsPublished    bool      `db:"is_published" json:"is_published"`
	IsDeleted      bool      `db:"is_deleted" json:"is_deleted"`

	HtmlName        string `db:"-" json:"-"`
	HtmlDescription string `db:"-" json:"-"`

	ThumbnailTimestamp string `db:"-" json:"-"`

	UserPerms *PermCookbook `db:"userperms" json:"-"`
}

func CountCookbooks(uid int64) (int, error) {
	var count int
	query := `
	SELECT
		COUNT(DISTINCT c.id) -- Use DISTINCT
	FROM
		cookbooks c
	LEFT JOIN -- Use LEFT JOIN for specific permissions
		permissions_cookbooks pc ON c.id = pc.cookbook_id AND pc.user_id = ?
	JOIN -- Use JOIN for account permissions (must exist)
		permissions_accounts pa ON c.account_id = pa.account_id AND pa.user_id = ?
	WHERE
		c.is_deleted = false
		AND (
			-- Check if specific permissions exist and grant access
			(pc.user_id IS NOT NULL AND (pc.can_view = true OR pc.can_edit = true OR pc.is_owner = true))
		)
	`
	err := db.Db().Get(&count, query, uid, uid)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func Cookbooks(uid int64, page, limit int) ([]Cookbook, error) {
	var cbs []Cookbook
	offset := (page - 1) * limit

	// Select cookbook details (c.*) and user-specific permissions (pc.*)
	// Use aliases prefixed with "userperms." to match the struct tag and enable automatic scanning.
	query := `
	SELECT
		c.*,
		pc.id AS "userperms.id",
		pc.user_id AS "userperms.user_id",
		pc.account_id AS "userperms.account_id",
		pc.cookbook_id AS "userperms.cookbook_id",
		pc.created AS "userperms.created",
		pc.updated AS "userperms.updated",
		pc.can_view AS "userperms.can_view",
		pc.can_edit AS "userperms.can_edit",
		pc.is_owner AS "userperms.is_owner"
	FROM
		cookbooks c
	LEFT JOIN
		permissions_cookbooks pc ON c.id = pc.cookbook_id AND pc.user_id = ?
	WHERE
		c.account_id = (SELECT account_id FROM permissions_accounts WHERE user_id = ? LIMIT 1) -- Infer account ID from user
		AND c.is_deleted = false
		AND pc.user_id IS NOT NULL -- Ensure user has *some* specific permission record for the cookbook
	ORDER BY
		c.updated DESC
	LIMIT ?
	OFFSET ?
	`
	// Note: The WHERE clause assumes a user belongs to only one account for this list context.
	// If a user can see cookbooks from multiple accounts they belong to, this logic needs refinement.

	// Use sqlx.Select which handles StructScan automatically for slices
	err := db.Db().Select(&cbs, query, uid, uid, limit, offset)
	if err != nil {
		// sql.ErrNoRows is not an error in this context, just means no cookbooks match.
		if errors.Is(err, sql.ErrNoRows) {
			return cbs, nil // Return empty slice
		}
		log.Printf("Error querying cookbooks with permissions for user %d: %v", uid, err)
		return nil, err
	}

	// Thumbnail Timestamp calculation needs to happen after fetching
	for i, cb := range cbs {
		yamlDefault := YamlDefault{}
		// Use YamlIndividual for thumbnail info as it might be user-specific
		if err := yaml.Unmarshal([]byte(cb.YamlIndividual), &yamlDefault); err != nil {
			// Log error but don't fail the whole request? Or maybe use YamlShared as fallback?
			log.Printf("Error unmarshalling YamlIndividual for cookbook %s (ID %d) thumbnail: %v", cb.Name, cb.ID, err)
			cbs[i].ThumbnailTimestamp = "error"
			continue
		}
		if len(yamlDefault.Cookbook.Storage.Thumbnail.B64) == 0 {
			cbs[i].ThumbnailTimestamp = "0000000000"
		} else {
			cbs[i].ThumbnailTimestamp = yamlDefault.Cookbook.Storage.Thumbnail.Timestamp
		}
	}

	return cbs, nil
}

var default_cookbook = `
cookbook:
    environment:
        public:
            - USER_DEFINED_PUBLIC_ENV_VAR=somesillypublicvar
        private:
            - USER_DEFINED_PRIVATE_ENV_VAR=somesillyprivatevarthatyouwantsecret
    pages:
        - page: 1
          name: Hello World Page
          recipes:
            - recipe: hello world
              description: basic hello world lemc example
              form: []
              steps:
                - step: 1
                  image: docker.io/jfolkins/lemc-helloworld:latest
                  do: now
                  timeout: 10.minutes
`

func (c *Cookbook) Create(tx *sqlx.Tx) error {
	uuidWithTime, err := uuid.NewV7()
	if err != nil {
		return err
	}
	apiKeyWithTime, err := uuid.NewV7()
	if err != nil {
		return err
	}

	c.UUID = uuidWithTime.String()
	c.ApiKey = fmt.Sprintf("api-%s", apiKeyWithTime.String())

	if len(c.YamlShared) == 0 {
		c.YamlShared = default_cookbook
	}

	if len(c.YamlIndividual) == 0 {
		c.YamlIndividual = default_cookbook
	}

	query := `
		insert into cookbooks(
			account_id, 
			owner_id, 
			uuid, 
			name, 
			description, 
			yaml_shared, 
			yaml_individual, 
			api_key)
		values(
			:account_id, 
			:owner_id, 
			:uuid, 
			:name, 
			:description, 
			:yaml_shared, 
			:yaml_individual, 
			:api_key)
		returning id`

	r, err := tx.NamedExec(query, c)
	if err != nil {
		log.Println("🔥 Failed to join User to Account: ", err)
		return err
	}

	c.ID, err = r.LastInsertId()
	if err != nil {
		return err
	}

	query = `
		insert into permissions_cookbooks(
			user_id, 
			account_id, 
			cookbook_id, 
			can_view, 
			can_edit, 
			is_owner)
		values(
			$1, 
			$2, 	
			$3, 	
			true, 	
			true, 
			true)`

	_, err = tx.Exec(query, c.OwnerID, c.AccountID, c.ID)
	if err != nil {
		log.Println("🔥 Failed to create cookbook: ", err)
		return err
	}
	return nil
}

func (c *Cookbook) Update() error {
	tx, err := db.Db().Beginx()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Fetch existing YAML to detect changes and for history
	prior := struct {
		ID             int64  `db:"id"`
		YamlShared     string `db:"yaml_shared"`
		YamlIndividual string `db:"yaml_individual"`
	}{}

	q := `SELECT id, yaml_shared, yaml_individual FROM cookbooks WHERE account_id = ? AND uuid = ?`
	if err = tx.Get(&prior, q, c.AccountID, c.UUID); err != nil {
		return err
	}

	query := `
                update
                        cookbooks
                set
                        account_id = :account_id,
                        uuid = :uuid,
                        name = :name,
                        description = :description,
                        yaml_shared = :yaml_shared,
                        yaml_individual = :yaml_individual,
                        is_published = :is_published,
                        is_deleted = :is_deleted
                where
                    account_id = :account_id
                and
                        uuid = :uuid`

	if _, err = tx.NamedExec(query, c); err != nil {
		return err
	}

	if prior.YamlShared != c.YamlShared || prior.YamlIndividual != c.YamlIndividual {
		if _, err = tx.Exec(`INSERT INTO cookbook_history (cookbook_id, yaml_shared, yaml_individual) VALUES (?, ?, ?)`, prior.ID, prior.YamlShared, prior.YamlIndividual); err != nil {
			return err
		}
	}

	err = tx.Commit()
	return err
}

func (c *Cookbook) ByName(name string) error {
	q := `
		select 
			id, created, updated, account_id, owner_id, uuid, name, description, yaml_shared, yaml_individual, api_key, is_published, is_deleted 
		from 
			cookbooks 
		where 
			name = $1`
	err := db.Db().Get(c, q, name)

	if err != nil {
		return err
	}

	return nil
}

func (c *Cookbook) ByUUID(uuid string) error {
	q1 := `
		select 
			id, created, updated, account_id, owner_id, uuid, name, description, yaml_shared, yaml_individual, api_key, is_published, is_deleted 
		from 
			cookbooks 
		where 
			uuid = $1`
	err := db.Db().Get(c, q1, uuid)

	if err != nil {
		return err
	}

	return nil
}

func (c *Cookbook) ByUUIDAndAccountID(uuid string, aid int64) error {
	q := `
		select 
			id, created, updated, account_id, owner_id, uuid, name, description, yaml_shared, yaml_individual, api_key, is_published, is_deleted 
		from 
			cookbooks 
		where 
			uuid = $1
		and
			account_id = $2`

	err := db.Db().Get(c, q, uuid, aid)

	if err != nil {
		return err
	}

	return nil
}

func CookbookByIDAndAccountID(id int64, accountID int64) (*Cookbook, error) {
	cookbook := &Cookbook{}
	query := `SELECT id, created, updated, account_id, owner_id, uuid, name, description, yaml_shared, yaml_individual, api_key, is_published, is_deleted FROM cookbooks WHERE id = $1 AND account_id = $2`
	err := db.Db().Get(cookbook, query, id, accountID)
	if err != nil {
		return nil, err
	}
	return cookbook, nil
}

func (pc *PermCookbook) UpsertCookbookPermissions(tx *sqlx.Tx) error {

	query := `
		insert into permissions_cookbooks(
			user_id, 
			account_id, 
			cookbook_id, 
			can_view, 
			can_edit, 
			is_owner)
          values(
			:user_id, 
			:account_id, 
			:cookbook_id, 
			:can_view, 
			:can_edit, 
			:is_owner)
          on conflict (
			user_id, 
			account_id, 
			cookbook_id) 
          do update set 
			can_view = excluded.can_view, 
            can_edit = excluded.can_edit, 
            is_owner = excluded.is_owner`

	_, err := tx.NamedExec(query, pc)
	if err != nil {
		log.Println("🔥 Failed to create cookbook: ", err)
		return err
	}
	return nil
}

func (pc *PermCookbook) UpdateCookbookPermissions(tx *sqlx.Tx) error {

	query := `
		update permissions_cookbooks
		set 
			can_view = :can_view, 
			can_edit = :can_edit, 
			is_owner = :is_owner
        where 
			user_id = :user_id 
		and 
			account_id = :account_id 
		and 
			cookbook_id = :cookbook_id`

	_, err := tx.NamedExec(query, pc)
	if err != nil {
		log.Println("🔥 Failed to create cookbook: ", err)
		return err
	}
	return nil
}

func (pc *PermCookbook) DeleteCookbookPermissions(tx *sqlx.Tx) error {

	query := `
		delete 
		from 
			permissions_cookbooks 
		where
			user_id = :user_id 
		and 
			account_id = :account_id 
		and 
			cookbook_id = :cookbook_id`

	_, err := tx.NamedExec(query, pc)
	if err != nil {
		log.Println("🔥 Failed to create cookbook: ", err)
		return err
	}
	return nil
}

func (pc *PermCookbook) CookbookPermissions(user_id, account_id, cookbook_id int64) error {
	query := `select created, updated, id, user_id, account_id, cookbook_id, can_view, can_edit, is_owner from permissions_cookbooks where user_id = $1 and account_id = $2 and cookbook_id = $3`
	err := db.Db().Get(pc, query, user_id, account_id, cookbook_id)
	if err != nil {
		return err
	}

	return nil
}

func AllCookbooks() ([]Cookbook, error) {
	cbs := []Cookbook{}
	query := `SELECT id, created, updated, account_id, owner_id, uuid, name, description, yaml_shared, yaml_individual, api_key, is_published, is_deleted FROM cookbooks WHERE is_deleted = false`
	err := db.Db().Select(&cbs, query)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []Cookbook{}, nil
		}
		return nil, err
	}
	return cbs, nil
}
