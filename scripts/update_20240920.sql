ALTER TABLE user_asset_group ADD COLUMN share_status tinyint DEFAULT '0' NOT NULL;
ALTER TABLE user_asset_group ADD COLUMN visit_count int NOT NULL DEFAULT '0';