-- +goose Up
-- +goose StatementBegin
CREATE TABLE apps (
   created  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
   updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
   id INTEGER PRIMARY KEY AUTOINCREMENT, -- For SQLite
   account_id INTEGER NOT NULL,
   owner_id INTEGER NOT NULL,
   cookbook_id integer not null,
   uuid TEXT NOT NULL,
   name TEXT NOT NULL,
   description TEXT NULL,
   yaml_shared TEXT NULL,
   yaml_individual TEXT NULL,
   api_key TEXT NOT NULL,
   is_active BOOLEAN NOT NULL DEFAULT false,
   is_deleted BOOLEAN NOT NULL DEFAULT false,
   is_assigned_by_default boolean not null default false,
   FOREIGN KEY (account_id) REFERENCES accounts(id),
   FOREIGN KEY (owner_id) REFERENCES users(id),
   UNIQUE(account_id, name)
);



CREATE INDEX apps_description_idx ON apps(description);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE apps;
-- +goose StatementEnd
