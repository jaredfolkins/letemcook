-- +goose Up
-- +goose StatementBegin
CREATE TABLE cookbooks (
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    id INTEGER PRIMARY KEY AUTOINCREMENT, -- For SQLite
    account_id INTEGER NOT NULL,
    owner_id INTEGER NOT NULL,
    uuid TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NULL,
    yaml_shared TEXT NULL,
    yaml_individual TEXT NULL,
    api_key TEXT NOT NULL,
    is_published BOOLEAN NOT NULL DEFAULT false,
    is_deleted BOOLEAN NOT NULL DEFAULT false,
    FOREIGN KEY (account_id) REFERENCES accounts(id),
    FOREIGN KEY (owner_id) REFERENCES users(id),
    UNIQUE(account_id, name)
);

CREATE INDEX cookbooks_description_idx ON cookbooks(description);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE cookbooks;
-- +goose StatementEnd