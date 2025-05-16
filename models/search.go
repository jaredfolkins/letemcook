package models

import "github.com/jaredfolkins/letemcook/db"

func SearchForCookbookAclUsersNotAssigned(search string, accountID, cookbookID int64, limit int) ([]CookbookAcl, error) {
	var cba []CookbookAcl
	/*
		query := `
			select
			users.id as "user_id",
			users.username,
			users.email
			from users
			join permissions_accounts on permissions_accounts.user_id = users.id
			join accounts on permissions_accounts.account_id = accounts.id
			left join permissions_cookbooks on permissions_cookbooks.user_id = users.id
			join cookbooks on cookbooks.id = permissions_cookbooks.cookbook_id
			where permissions_cookbooks.account_id != $1
			and permissions_cookbooks.cookbook_id != $2
			and accounts.id = $1
			and (users.username LIKE $3 or users.email LIKE $3)
			order by users.username
			desc limit $4`
	*/

	query := `select
	users.id as "user_id",
		users.username,
		users.email
	from users
	join permissions_accounts on permissions_accounts.user_id = users.id
	join accounts on permissions_accounts.account_id = accounts.id
	left join permissions_cookbooks on permissions_cookbooks.user_id = users.id
	and permissions_cookbooks.cookbook_id = $1
	where accounts.id = $2
	and permissions_cookbooks.cookbook_id is null
	and (users.username like $3 or users.email like $3)
	order by users.username
	desc limit $4`

	search = "%" + search + "%"
	err := db.Db().Select(&cba, query, cookbookID, accountID, search, limit)
	if err != nil {
		return nil, err
	}
	return cba, nil
}

func SearchForCookbooks(search string, userID, accountID int64, limit int, is_published bool) ([]Cookbook, error) {
	var cba []Cookbook
	query := `
		select
		    cookbooks.id as "id",
		    cookbooks.uuid as "uuid",
			cookbooks.name as "name",
			cookbooks.description as "description"
		from 
			cookbooks
		join 
			permissions_cookbooks on permissions_cookbooks.cookbook_id = cookbooks.id
		join 
			accounts on permissions_cookbooks.account_id = accounts.id
		where
			permissions_cookbooks.user_id = $1
		and
			accounts.id = $2
		and 
			(cookbooks.name like $3 or cookbooks.description like $3)
		and
			cookbooks.is_deleted = false
		and
			cookbooks.is_published = $4
		order by 
			cookbooks.name
		desc limit 
			$5`

	search = "%" + search + "%"
	err := db.Db().Select(&cba, query, userID, accountID, search, is_published, limit)
	if err != nil {
		return nil, err
	}
	return cba, nil
}

func CookbookAclsUsers(accountID, cookbookID int64) ([]CookbookAcl, error) {
	var cba []CookbookAcl
	query := `
		select
		users.id as "user_id",
		users.username as "username",
		users.email as "email",
		permissions_cookbooks.can_view as "can_view",
		permissions_cookbooks.can_edit as "can_edit",
		permissions_cookbooks.is_owner as "is_owner"
		from users
		join permissions_cookbooks on permissions_cookbooks.user_id = users.id
		join accounts on permissions_cookbooks.account_id = accounts.id
		where accounts.id = $1
		and permissions_cookbooks.cookbook_id = $2
		and permissions_cookbooks.user_id IS NOT NULL
		order by users.username
	`

	err := db.Db().Select(&cba, query, accountID, cookbookID)
	if err != nil {
		return nil, err
	}
	return cba, nil
}

func SearchForappAclUsersNotAssigned(search string, accountID, appID int64, limit int) ([]AppAcl, error) {
	var ca []AppAcl // Use AppAcl slice
	query := `
        SELECT
            u.id as "user_id",
            u.username,
            u.email
        FROM users u
        JOIN permissions_accounts pa ON pa.user_id = u.id
        LEFT JOIN permissions_apps pc ON pc.user_id = u.id AND pc.app_id = $1
        WHERE pa.account_id = $2
          AND pc.app_id IS NULL -- Key condition: user not in App permissions
          AND (u.username LIKE $3 OR u.email LIKE $3)
        ORDER BY u.username
        LIMIT $4`

	search = "%" + search + "%"
	err := db.Db().Select(&ca, query, appID, accountID, search, limit)
	if err != nil {
		return nil, err
	}
	return ca, nil
}
