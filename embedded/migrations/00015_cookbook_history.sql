-- +goose Up
-- +goose StatementBegin
CREATE TABLE cookbook_history (
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    cookbook_id INTEGER NOT NULL,
    yaml_shared TEXT,
    yaml_individual TEXT,
    FOREIGN KEY (cookbook_id) REFERENCES cookbooks(id)
);

CREATE TRIGGER IF NOT EXISTS trg_cookbooks_history
AFTER UPDATE ON cookbooks
FOR EACH ROW
WHEN OLD.yaml_shared != NEW.yaml_shared OR OLD.yaml_individual != NEW.yaml_individual
BEGIN
    INSERT INTO cookbook_history (cookbook_id, yaml_shared, yaml_individual)
    VALUES (OLD.id, OLD.yaml_shared, OLD.yaml_individual);
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_cookbooks_history;
DROP TABLE IF EXISTS cookbook_history;
-- +goose StatementEnd
