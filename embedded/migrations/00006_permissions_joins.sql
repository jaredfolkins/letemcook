-- +goose Up
-- +goose StatementBegin
CREATE TABLE permissions_system (
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    id INTEGER PRIMARY KEY AUTOINCREMENT, -- For SQLite
    user_id INTEGER NOT NULL references users(id),
    can_administer BOOLEAN NOT NULL DEFAULT FALSE,
    is_owner BOOLEAN NOT NULL DEFAULT FALSE,
    UNIQUE(user_id)
);

CREATE INDEX permissions_system_user_id_idx ON permissions_system (user_id);

CREATE TABLE permissions_accounts (
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    id INTEGER PRIMARY KEY AUTOINCREMENT, -- For SQLite
    user_id INTEGER NOT NULL references users(id),
    account_id INTEGER NOT NULL references accounts(id),
    can_administer BOOLEAN NOT NULL DEFAULT FALSE,
    can_create_apps BOOLEAN NOT NULL DEFAULT FALSE,
    can_view_apps BOOLEAN NOT NULL DEFAULT FALSE,
    can_create_cookbooks BOOLEAN NOT NULL DEFAULT FALSE,
    can_view_cookbooks BOOLEAN NOT NULL DEFAULT FALSE,
    is_owner BOOLEAN NOT NULL DEFAULT FALSE,
    UNIQUE(account_id, user_id)
);

CREATE INDEX permissions_accounts_account_id_idx ON permissions_accounts (account_id);
CREATE INDEX permissions_accounts_user_id_idx ON permissions_accounts (user_id);

CREATE TABLE permissions_apps (
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    id INTEGER PRIMARY KEY AUTOINCREMENT, -- For SQLite
    user_id INTEGER NOT NULL references users(id),
    account_id INTEGER NOT NULL references accounts(id),
    app_id INTEGER NOT NULL references apps(id),
    cookbook_id INTEGER NOT NULL references cookbooks(id),
    can_shared BOOLEAN NOT NULL DEFAULT FALSE,
    can_individual BOOLEAN NOT NULL DEFAULT FALSE,
    can_administer BOOLEAN NOT NULL DEFAULT FALSE,
    is_owner BOOLEAN NOT NULL DEFAULT FALSE,
    UNIQUE(account_id, user_id, app_id)
);

CREATE INDEX permissions_apps_account_id_idx ON permissions_apps (account_id);
CREATE INDEX permissions_apps_user_id_idx ON permissions_apps (user_id);

CREATE TABLE permissions_cookbooks (
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    id INTEGER PRIMARY KEY AUTOINCREMENT, -- For SQLite
    user_id INTEGER NOT NULL references users(id),
    account_id INTEGER NOT NULL references accounts(id),
    cookbook_id INTEGER NOT NULL references cookbooks(id),
    can_view BOOLEAN NOT NULL DEFAULT FALSE,
    can_edit BOOLEAN NOT NULL DEFAULT FALSE,
    is_owner BOOLEAN NOT NULL DEFAULT FALSE,
    UNIQUE(account_id, user_id, cookbook_id)
);

CREATE INDEX permissions_cookbooks_account_id_idx ON permissions_cookbooks (account_id);
CREATE INDEX permissions_cookbooks_user_id_idx ON permissions_cookbooks (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE permissions_accounts;
DROP TABLE permissions_apps;
DROP TABLE permissions_cookbooks;
DROP TABLE permissions_system;
-- +goose StatementEnd