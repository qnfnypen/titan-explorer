ALTER TABLE user_asset DROP COLUMN `short_pass`;

ALTER TABLE link ADD COLUMN `short_pass` CHAR(64) NOT NULL DEFAULT '';

ALTER TABLE link ADD COLUMN `expire_time` DATETIME;