-- +goose Up

-- +goose StatementBegin

-- tables
CREATE TABLE users (
   created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
   updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
   id INTEGER PRIMARY KEY AUTOINCREMENT, -- For SQLite
   username VARCHAR(255) NOT NULL UNIQUE,
   email VARCHAR(255) NULL UNIQUE,
   hash VARCHAR(255) NOT NULL,
   is_disabled BOOLEAN NOT NULL DEFAULT false,
   is_deleted BOOLEAN NOT NULL DEFAULT false
);

-- indexes
CREATE UNIQUE INDEX users_username_idx ON users (username);



-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
-- +goose StatementEnd
