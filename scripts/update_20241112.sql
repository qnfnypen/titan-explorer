ALTER TABLE user_asset ADD COLUMN `client_ip`  char(64) NOT NULL DEFAULT '' COMMENT 'user upload client ip';

ALTER TABLE assets ADD COLUMN `client_ip` char(64) NOT NULL DEFAULT '' COMMENT 'user upload client ip';

ALTER TABLE device_info ADD COLUMN `replica_count` bigint(20) NOT NULL DEFAULT '0' COMMENT '实时的文件数量';