-- +goose Up
-- +goose StatementBegin
CREATE TABLE app_history (
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    app_id INTEGER NOT NULL,
    yaml_shared TEXT,
    yaml_individual TEXT,
    FOREIGN KEY (app_id) REFERENCES apps(id)
);

CREATE TRIGGER IF NOT EXISTS trg_apps_history
AFTER UPDATE ON apps
FOR EACH ROW
WHEN OLD.yaml_shared != NEW.yaml_shared OR OLD.yaml_individual != NEW.yaml_individual
BEGIN
    INSERT INTO app_history (app_id, yaml_shared, yaml_individual)
    VALUES (OLD.id, OLD.yaml_shared, OLD.yaml_individual);
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_apps_history;
DROP TABLE IF EXISTS app_history;
-- +goose StatementEnd
