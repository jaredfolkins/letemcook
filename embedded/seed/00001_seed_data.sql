-- +goose Up
-- +goose StatementBegin
INSERT INTO accounts (id, squid, name) VALUES
    (1, 'xkQN', 'Account Alpha'),
    (2, 'Xijg', 'Account Bravo');

INSERT INTO users (id, username, email, hash) VALUES
    (1, 'alpha-owner', 'alpha-owner@example.com', '$2b$12$Aj7VVYncuqKkZCUMVmgYCuDqV1Dsv2IaolWANsgEAU9t7sL/8Js8C'),
    (2, 'alpha-admin-2', 'alpha-admin-2@example.com', '$2b$12$Aj7VVYncuqKkZCUMVmgYCuDqV1Dsv2IaolWANsgEAU9t7sL/8Js8C'),
    (3, 'bravo-owner', 'bravo-owner@example.com', '$2b$12$Aj7VVYncuqKkZCUMVmgYCuDqV1Dsv2IaolWANsgEAU9t7sL/8Js8C'),
    (4, 'bravo-admin-2', 'bravo-admin-2@example.com', '$2b$12$Aj7VVYncuqKkZCUMVmgYCuDqV1Dsv2IaolWANsgEAU9t7sL/8Js8C');

INSERT INTO permissions_system (user_id, can_administer, is_owner) VALUES
    (1, 1, 1),
    (2, 1, 0),
    (3, 1, 1),
    (4, 1, 0);

INSERT INTO permissions_accounts (user_id, account_id, can_administer, can_create_apps, can_view_apps, can_create_cookbooks, can_view_cookbooks, is_owner) VALUES
    (1, 1, 1, 1, 1, 1, 1, 1),
    (2, 1, 1, 1, 1, 1, 1, 0),
    (3, 2, 1, 1, 1, 1, 1, 1),
    (4, 2, 1, 1, 1, 1, 1, 0);

INSERT INTO cookbooks (id, account_id, owner_id, uuid, name, description, yaml_shared, yaml_individual, api_key) VALUES
    (1, 1, 1, lower(hex(randomblob(16))), 'Alpha Cookbook 1', 'Description for Alpha cookbook #1.', 'shared_cookbook_key: shared_value', 'individual_cookbook_key: individual_value', lower(hex(randomblob(16)))),
    (2, 2, 3, lower(hex(randomblob(16))), 'Bravo Cookbook 1', 'Description for Bravo cookbook #1.', 'shared_cookbook_key: shared_value', 'individual_cookbook_key: individual_value', lower(hex(randomblob(16))));

INSERT INTO permissions_cookbooks (account_id, user_id, cookbook_id, can_view, can_edit, is_owner) VALUES
    (1, 1, 1, 1, 1, 1),
    (1, 2, 1, 1, 1, 0),
    (2, 3, 2, 1, 1, 1),
    (2, 4, 2, 1, 1, 0);

INSERT INTO apps (id, account_id, owner_id, cookbook_id, uuid, name, description, yaml_shared, yaml_individual, api_key, is_active) VALUES
    (1, 1, 1, 1, lower(hex(randomblob(16))), 'Alpha App 1', 'Description for Alpha App #1.', 'shared_app_key: shared_value', 'individual_app_key: individual_value', lower(hex(randomblob(16))), 1),
    (2, 2, 3, 2, lower(hex(randomblob(16))), 'Bravo App 1', 'Description for Bravo App #1.', 'shared_app_key: shared_value', 'individual_app_key: individual_value', lower(hex(randomblob(16))), 1);

INSERT INTO permissions_apps (account_id, user_id, app_id, cookbook_id, can_shared, can_individual, can_administer, is_owner, api_key) VALUES
    (1, 1, 1, 1, 1, 1, 1, 1, lower(hex(randomblob(16)))),
    (1, 2, 1, 1, 1, 1, 1, 0, lower(hex(randomblob(16)))),
    (2, 3, 2, 2, 1, 1, 1, 1, lower(hex(randomblob(16)))),
    (2, 4, 2, 2, 1, 1, 1, 0, lower(hex(randomblob(16))));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM permissions_apps;
DELETE FROM apps;
DELETE FROM permissions_cookbooks;
DELETE FROM cookbooks;
DELETE FROM permissions_accounts;
DELETE FROM permissions_system;
DELETE FROM users;
DELETE FROM accounts;
-- +goose StatementEnd
