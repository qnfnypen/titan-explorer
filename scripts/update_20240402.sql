DROP TABLE IF EXISTS `app_version`;
CREATE TABLE app_version(
`id` bigint(20) NOT NULL AUTO_INCREMENT,
`version`  VARCHAR(128) NOT NULL DEFAULT '',
`min_version` VARCHAR(128) NOT NULL DEFAULT '',
`description` TEXT NOT NULL,
`url` VARCHAR(128) NOT NULL DEFAULT '',
`platform`  VARCHAR(64) NOT NULL DEFAULT '',
`size` bigint(20) NOT NULL DEFAULT 0,
`lang` VARCHAR(64) NOT NULL DEFAULT '',
`created_at` DATETIME(3) NOT NULL DEFAULT 0,
`updated_at` DATETIME(3) NOT NULL DEFAULT 0,
PRIMARY KEY (`id`)
)ENGINE = INNODB CHARSET = utf8mb4;


ALTER TABLE `titan_explorer`.`device_info`
    ADD INDEX `idx_user_id`(`user_id`) USING BTREE;