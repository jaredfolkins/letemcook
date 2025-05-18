-- +goose Up

-- +goose StatementBegin

PRAGMA foreign_keys = ON;

-- tables
CREATE TABLE accounts (
   created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
   updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
   id INTEGER PRIMARY KEY AUTOINCREMENT, -- For SQLite
   squid VARCHAR(255) NOT NULL,
   name VARCHAR(255) NOT NULL,
   is_deleted BOOLEAN NOT NULL DEFAULT false,
   UNIQUE(name) ON CONFLICT REPLACE
);

-- indexes
CREATE UNIQUE INDEX accounts_name_idx ON accounts (name);

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP TABLE accounts;
-- +goose StatementEnd
