-- +goose Up
-- +goose StatementBegin
CREATE TABLE configurations (
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    id INTEGER PRIMARY KEY AUTOINCREMENT, -- For SQLite
    account_id INTEGER NOT NULL,
    theme INTEGER TEXT NOT NULL,
    enable_register boolean not null default false,
    enable_heckle boolean not null default false,
    enable_api boolean not null default false,
    account_api_key TEXT NOT NULL,
    FOREIGN KEY (account_id) REFERENCES accounts (id),
    UNIQUE(account_id)
);

CREATE INDEX configurations_id_idx ON configurations(id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE configurations;
-- +goose StatementEnd
