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

-- Additional Users
INSERT INTO users (id, username, email, hash) VALUES
    (5, 'alpha-viewer', 'alpha-viewer@example.com', '$2b$12$Aj7VVYncuqKkZCUMVmgYCuDqV1Dsv2IaolWANsgEAU9t7sL/8Js8C'),
    (6, 'alpha-editor', 'alpha-editor@example.com', '$2b$12$Aj7VVYncuqKkZCUMVmgYCuDqV1Dsv2IaolWANsgEAU9t7sL/8Js8C'),
    (7, 'alpha-limited', 'alpha-limited@example.com', '$2b$12$Aj7VVYncuqKkZCUMVmgYCuDqV1Dsv2IaolWANsgEAU9t7sL/8Js8C'),
    (8, 'bravo-main', 'bravo-main@example.com', '$2b$12$Aj7VVYncuqKkZCUMVmgYCuDqV1Dsv2IaolWANsgEAU9t7sL/8Js8C'),
    (9, 'bravo-extra', 'bravo-extra@example.com', '$2b$12$Aj7VVYncuqKkZCUMVmgYCuDqV1Dsv2IaolWANsgEAU9t7sL/8Js8C');

-- Additional System Permissions for new users
INSERT INTO permissions_system (user_id, can_administer, is_owner) VALUES
    (5, 0, 0), -- alpha-viewer
    (6, 0, 0), -- alpha-editor
    (7, 0, 0), -- alpha-limited
    (8, 0, 0), -- bravo-main
    (9, 0, 0); -- bravo-extra

-- Additional Account Permissions for new users
INSERT INTO permissions_accounts (user_id, account_id, can_administer, can_create_apps, can_view_apps, can_create_cookbooks, can_view_cookbooks, is_owner) VALUES
    (5, 1, 0, 0, 1, 0, 1, 0), -- alpha-viewer on Account Alpha
    (6, 1, 0, 0, 1, 0, 1, 0), -- alpha-editor on Account Alpha
    (7, 1, 0, 0, 1, 0, 1, 0), -- alpha-limited on Account Alpha
    (8, 2, 0, 0, 1, 0, 1, 0), -- bravo-main on Account Bravo
    (9, 2, 0, 0, 1, 0, 1, 0); -- bravo-extra on Account Bravo

-- Additional Cookbooks
INSERT INTO cookbooks (id, account_id, owner_id, uuid, name, description, yaml_shared, yaml_individual, api_key) VALUES
    (3, 1, 1, lower(hex(randomblob(16))), 'Alpha Cookbook 2', 'Description for Alpha cookbook #2.', 'shared_cookbook_key: shared_value', 'individual_cookbook_key: individual_value', lower(hex(randomblob(16)))),
    (4, 2, 3, lower(hex(randomblob(16))), 'Bravo Cookbook 2', 'Description for Bravo cookbook #2.', 'shared_cookbook_key: shared_value', 'individual_cookbook_key: individual_value', lower(hex(randomblob(16))));

-- Permissions for new Cookbooks (owners)
INSERT INTO permissions_cookbooks (account_id, user_id, cookbook_id, can_view, can_edit, is_owner) VALUES
    (1, 1, 3, 1, 1, 1), -- alpha-owner on Alpha Cookbook 2
    (2, 3, 4, 1, 1, 1); -- bravo-owner on Bravo Cookbook 2

-- Additional Apps (to meet static user perm requirements)
INSERT INTO apps (id, account_id, owner_id, cookbook_id, uuid, name, description, yaml_shared, yaml_individual, api_key, is_active) VALUES
    (3, 1, 1, 3, lower(hex(randomblob(16))), 'Alpha App 2', 'Description for Alpha App #2.', 'shared_app_key: shared_value', 'individual_app_key: individual_value', lower(hex(randomblob(16))), 1),
    (4, 2, 3, 4, lower(hex(randomblob(16))), 'Bravo App 2', 'Description for Bravo App #2.', 'shared_app_key: shared_value', 'individual_app_key: individual_value', lower(hex(randomblob(16))), 1),
    (5, 2, 3, 4, lower(hex(randomblob(16))), 'Bravo App 3', 'Description for Bravo App #3.', 'shared_app_key: shared_value', 'individual_app_key: individual_value', lower(hex(randomblob(16))), 1);

-- Permissions for new Apps (owners)
INSERT INTO permissions_apps (account_id, user_id, app_id, cookbook_id, can_shared, can_individual, can_administer, is_owner, api_key) VALUES
    (1, 1, 3, 3, 1, 1, 1, 1, lower(hex(randomblob(16)))), -- alpha-owner on Alpha App 2
    (2, 3, 4, 4, 1, 1, 1, 1, lower(hex(randomblob(16)))), -- bravo-owner on Bravo App 2
    (2, 3, 5, 4, 1, 1, 1, 1, lower(hex(randomblob(16)))); -- bravo-owner on Bravo App 3

-- Permissions for new Apps (static users as per Go code appPermSpecs)
-- alpha-viewer (user 5) on Alpha App 1 (app 1, cookbook 1) - appIndex 0
INSERT INTO permissions_apps (account_id, user_id, app_id, cookbook_id, can_shared, can_individual, can_administer, is_owner, api_key) VALUES
    (1, 5, 1, 1, 1, 1, 0, 0, lower(hex(randomblob(16))));
-- alpha-editor (user 6) on Alpha App 2 (app 3, cookbook 3) - appIndex 1
INSERT INTO permissions_apps (account_id, user_id, app_id, cookbook_id, can_shared, can_individual, can_administer, is_owner, api_key) VALUES
    (1, 6, 3, 3, 1, 1, 1, 0, lower(hex(randomblob(16))));
-- alpha-limited (user 7) on Alpha App 1 (app 1, cookbook 1) - appIndex 0
INSERT INTO permissions_apps (account_id, user_id, app_id, cookbook_id, can_shared, can_individual, can_administer, is_owner, api_key) VALUES
    (1, 7, 1, 1, 0, 1, 0, 0, lower(hex(randomblob(16))));

-- bravo-main (user 8) on Bravo App 1 (app 2, cookbook 2) - appIndex 0
INSERT INTO permissions_apps (account_id, user_id, app_id, cookbook_id, can_shared, can_individual, can_administer, is_owner, api_key) VALUES
    (2, 8, 2, 2, 1, 1, 1, 0, lower(hex(randomblob(16))));
-- bravo-main (user 8) on Bravo App 2 (app 4, cookbook 4) - appIndex 1
INSERT INTO permissions_apps (account_id, user_id, app_id, cookbook_id, can_shared, can_individual, can_administer, is_owner, api_key) VALUES
    (2, 8, 4, 4, 1, 1, 0, 0, lower(hex(randomblob(16))));

-- bravo-extra (user 9) on Bravo App 3 (app 5, cookbook 4) - appIndex 2
INSERT INTO permissions_apps (account_id, user_id, app_id, cookbook_id, can_shared, can_individual, can_administer, is_owner, api_key) VALUES
    (2, 9, 5, 4, 1, 0, 0, 0, lower(hex(randomblob(16))));

-- Permissions for admin-2 users on new second apps
-- alpha-admin-2 (user 2) on Alpha App 2 (app 3, cookbook 3) - appIndex 1
INSERT INTO permissions_apps (account_id, user_id, app_id, cookbook_id, can_shared, can_individual, can_administer, is_owner, api_key) VALUES
    (1, 2, 3, 3, 1, 1, 1, 0, lower(hex(randomblob(16))));
-- bravo-admin-2 (user 4) on Bravo App 2 (app 4, cookbook 4) - appIndex 1
INSERT INTO permissions_apps (account_id, user_id, app_id, cookbook_id, can_shared, can_individual, can_administer, is_owner, api_key) VALUES
    (2, 4, 4, 4, 1, 1, 1, 0, lower(hex(randomblob(16))));

-- Additional 11 Cookbooks for Account Alpha (alpha-owner, user_id 1, account_id 1)
INSERT INTO cookbooks (id, account_id, owner_id, uuid, name, description, yaml_shared, yaml_individual, api_key) VALUES
    (5, 1, 1, lower(hex(randomblob(16))), 'Alpha Cookbook 3', 'Description for Alpha cookbook #3.', 'shared_cookbook_key: shared_value', 'individual_cookbook_key: individual_value', lower(hex(randomblob(16)))),
    (6, 1, 1, lower(hex(randomblob(16))), 'Alpha Cookbook 4', 'Description for Alpha cookbook #4.', 'shared_cookbook_key: shared_value', 'individual_cookbook_key: individual_value', lower(hex(randomblob(16)))),
    (7, 1, 1, lower(hex(randomblob(16))), 'Alpha Cookbook 5', 'Description for Alpha cookbook #5.', 'shared_cookbook_key: shared_value', 'individual_cookbook_key: individual_value', lower(hex(randomblob(16)))),
    (8, 1, 1, lower(hex(randomblob(16))), 'Alpha Cookbook 6', 'Description for Alpha cookbook #6.', 'shared_cookbook_key: shared_value', 'individual_cookbook_key: individual_value', lower(hex(randomblob(16)))),
    (9, 1, 1, lower(hex(randomblob(16))), 'Alpha Cookbook 7', 'Description for Alpha cookbook #7.', 'shared_cookbook_key: shared_value', 'individual_cookbook_key: individual_value', lower(hex(randomblob(16)))),
    (10, 1, 1, lower(hex(randomblob(16))), 'Alpha Cookbook 8', 'Description for Alpha cookbook #8.', 'shared_cookbook_key: shared_value', 'individual_cookbook_key: individual_value', lower(hex(randomblob(16)))),
    (11, 1, 1, lower(hex(randomblob(16))), 'Alpha Cookbook 9', 'Description for Alpha cookbook #9.', 'shared_cookbook_key: shared_value', 'individual_cookbook_key: individual_value', lower(hex(randomblob(16)))),
    (12, 1, 1, lower(hex(randomblob(16))), 'Alpha Cookbook 10', 'Description for Alpha cookbook #10.', 'shared_cookbook_key: shared_value', 'individual_cookbook_key: individual_value', lower(hex(randomblob(16)))),
    (13, 1, 1, lower(hex(randomblob(16))), 'Alpha Cookbook 11', 'Description for Alpha cookbook #11.', 'shared_cookbook_key: shared_value', 'individual_cookbook_key: individual_value', lower(hex(randomblob(16)))),
    (14, 1, 1, lower(hex(randomblob(16))), 'Alpha Cookbook 12', 'Description for Alpha cookbook #12.', 'shared_cookbook_key: shared_value', 'individual_cookbook_key: individual_value', lower(hex(randomblob(16)))),
    (15, 1, 1, lower(hex(randomblob(16))), 'Alpha Cookbook 13', 'Description for Alpha cookbook #13.', 'shared_cookbook_key: shared_value', 'individual_cookbook_key: individual_value', lower(hex(randomblob(16))));

INSERT INTO permissions_cookbooks (account_id, user_id, cookbook_id, can_view, can_edit, is_owner) VALUES
    (1, 1, 5, 1, 1, 1),
    (1, 1, 6, 1, 1, 1),
    (1, 1, 7, 1, 1, 1),
    (1, 1, 8, 1, 1, 1),
    (1, 1, 9, 1, 1, 1),
    (1, 1, 10, 1, 1, 1),
    (1, 1, 11, 1, 1, 1),
    (1, 1, 12, 1, 1, 1),
    (1, 1, 13, 1, 1, 1),
    (1, 1, 14, 1, 1, 1),
    (1, 1, 15, 1, 1, 1);

-- Additional 11 Apps for Account Alpha (linked to new cookbooks)
INSERT INTO apps (id, account_id, owner_id, cookbook_id, uuid, name, description, yaml_shared, yaml_individual, api_key, is_active) VALUES
    (6, 1, 1, 5, lower(hex(randomblob(16))), 'Alpha App 3', 'Description for Alpha App #3.', 'shared_app_key: shared_value', 'individual_app_key: individual_value', lower(hex(randomblob(16))), 1),
    (7, 1, 1, 6, lower(hex(randomblob(16))), 'Alpha App 4', 'Description for Alpha App #4.', 'shared_app_key: shared_value', 'individual_app_key: individual_value', lower(hex(randomblob(16))), 1),
    (8, 1, 1, 7, lower(hex(randomblob(16))), 'Alpha App 5', 'Description for Alpha App #5.', 'shared_app_key: shared_value', 'individual_app_key: individual_value', lower(hex(randomblob(16))), 1),
    (9, 1, 1, 8, lower(hex(randomblob(16))), 'Alpha App 6', 'Description for Alpha App #6.', 'shared_app_key: shared_value', 'individual_app_key: individual_value', lower(hex(randomblob(16))), 1),
    (10, 1, 1, 9, lower(hex(randomblob(16))), 'Alpha App 7', 'Description for Alpha App #7.', 'shared_app_key: shared_value', 'individual_app_key: individual_value', lower(hex(randomblob(16))), 1),
    (11, 1, 1, 10, lower(hex(randomblob(16))), 'Alpha App 8', 'Description for Alpha App #8.', 'shared_app_key: shared_value', 'individual_app_key: individual_value', lower(hex(randomblob(16))), 1),
    (12, 1, 1, 11, lower(hex(randomblob(16))), 'Alpha App 9', 'Description for Alpha App #9.', 'shared_app_key: shared_value', 'individual_app_key: individual_value', lower(hex(randomblob(16))), 1),
    (13, 1, 1, 12, lower(hex(randomblob(16))), 'Alpha App 10', 'Description for Alpha App #10.', 'shared_app_key: shared_value', 'individual_app_key: individual_value', lower(hex(randomblob(16))), 1),
    (14, 1, 1, 13, lower(hex(randomblob(16))), 'Alpha App 11', 'Description for Alpha App #11.', 'shared_app_key: shared_value', 'individual_app_key: individual_value', lower(hex(randomblob(16))), 1),
    (15, 1, 1, 14, lower(hex(randomblob(16))), 'Alpha App 12', 'Description for Alpha App #12.', 'shared_app_key: shared_value', 'individual_app_key: individual_value', lower(hex(randomblob(16))), 1),
    (16, 1, 1, 15, lower(hex(randomblob(16))), 'Alpha App 13', 'Description for Alpha App #13.', 'shared_app_key: shared_value', 'individual_app_key: individual_value', lower(hex(randomblob(16))), 1);

INSERT INTO permissions_apps (account_id, user_id, app_id, cookbook_id, can_shared, can_individual, can_administer, is_owner, api_key) VALUES
    (1, 1, 6, 5, 1, 1, 1, 1, lower(hex(randomblob(16)))),
    (1, 1, 7, 6, 1, 1, 1, 1, lower(hex(randomblob(16)))),
    (1, 1, 8, 7, 1, 1, 1, 1, lower(hex(randomblob(16)))),
    (1, 1, 9, 8, 1, 1, 1, 1, lower(hex(randomblob(16)))),
    (1, 1, 10, 9, 1, 1, 1, 1, lower(hex(randomblob(16)))),
    (1, 1, 11, 10, 1, 1, 1, 1, lower(hex(randomblob(16)))),
    (1, 1, 12, 11, 1, 1, 1, 1, lower(hex(randomblob(16)))),
    (1, 1, 13, 12, 1, 1, 1, 1, lower(hex(randomblob(16)))),
    (1, 1, 14, 13, 1, 1, 1, 1, lower(hex(randomblob(16)))),
    (1, 1, 15, 14, 1, 1, 1, 1, lower(hex(randomblob(16)))),
    (1, 1, 16, 15, 1, 1, 1, 1, lower(hex(randomblob(16))));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Delete additional 11 apps and their permissions for Account Alpha (Owner: user_id 1)
DELETE FROM permissions_apps WHERE app_id BETWEEN 6 AND 16 AND account_id = 1 AND user_id = 1;
DELETE FROM apps WHERE id BETWEEN 6 AND 16 AND account_id = 1;

-- Delete additional 11 cookbooks and their permissions for Account Alpha (Owner: user_id 1)
DELETE FROM permissions_cookbooks WHERE cookbook_id BETWEEN 5 AND 15 AND account_id = 1 AND user_id = 1;
DELETE FROM cookbooks WHERE id BETWEEN 5 AND 15 AND account_id = 1;

DELETE FROM permissions_apps WHERE user_id >= 5 OR app_id >=3;
DELETE FROM permissions_apps WHERE user_id = 2 AND app_id = 3; -- alpha-admin-2 on Alpha App 2
DELETE FROM permissions_apps WHERE user_id = 4 AND app_id = 4; -- bravo-admin-2 on Bravo App 2
DELETE FROM apps WHERE id >= 3;
DELETE FROM permissions_cookbooks WHERE cookbook_id >= 3;
DELETE FROM cookbooks WHERE id >= 3;
DELETE FROM permissions_accounts WHERE user_id >= 5;
DELETE FROM permissions_system WHERE user_id >= 5;
DELETE FROM users WHERE id >= 5;
DELETE FROM permissions_apps;
DELETE FROM accounts;
-- +goose StatementEnd
