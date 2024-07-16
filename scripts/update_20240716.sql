DROP TABLE IF EXISTS `asset_group`;
CREATE TABLE `asset_group` (
`id` INT UNSIGNED AUTO_INCREMENT,
`user_id` VARCHAR(128) NOT NULL,
`name` VARCHAR(32) DEFAULT '',
`parent` INT DEFAULT 0,
`created_time` DATETIME DEFAULT CURRENT_TIMESTAMP,
PRIMARY KEY (`id`),
KEY `idx_user_id` (`user_id`),
KEY `idx_parent` (`parent`)
) ENGINE=InnoDB COMMENT='titan存储服务的用户文件组表';