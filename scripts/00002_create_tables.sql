
DROP TABLE IF EXISTS `subscription`;
CREATE TABLE subscription (
`id` bigint(20) NOT NULL AUTO_INCREMENT,
`company`  VARCHAR(128) NOT NULL DEFAULT '',
`name` VARCHAR(128) NOT NULL DEFAULT '',
`email` VARCHAR(128) NOT NULL DEFAULT '',
`telegram` VARCHAR(128) NOT NULL DEFAULT '',
`wechat` VARCHAR(128) NOT NULL DEFAULT '',
`location` VARCHAR(128) NOT NULL DEFAULT '',
`storage` VARCHAR(128) NOT NULL DEFAULT '',
`calculation` VARCHAR(128) NOT NULL DEFAULT '',
`bandwidth` VARCHAR(128) NOT NULL DEFAULT '',
`join_testnet` int(3) not null default 0,
`idle_resource_percentages`  VARCHAR(128) NOT NULL DEFAULT '',
`subscribe` int(3) not null default 0,
`source`  VARCHAR(128) NOT NULL DEFAULT '',
`created_at` DATETIME(3) NOT NULL DEFAULT 0,
`updated_at` DATETIME(3) NOT NULL DEFAULT 0,
PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


DROP TABLE IF EXISTS `signature`;
CREATE TABLE signature (
`id` bigint(20) NOT NULL AUTO_INCREMENT,
`username`  VARCHAR(128) NOT NULL DEFAULT '',
`node_id` VARCHAR(128) NOT NULL DEFAULT '',
`area_id` VARCHAR(128) NOT NULL DEFAULT '',
`message` VARCHAR(128) NOT NULL DEFAULT '',
`hash` VARCHAR(128) NOT NULL DEFAULT '',
`signature` VARCHAR(256) NOT NULL DEFAULT '',
`created_at` DATETIME(3) NOT NULL DEFAULT 0,
`updated_at` DATETIME(3) NOT NULL DEFAULT 0,
PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `storage_user_info`;
CREATE TABLE storage_user_info (
`user_id` VARCHAR(128) NOT NULL,
`total_storage_size` BIGINT DEFAULT 0,
`used_storage_size` BIGINT DEFAULT 0,
`api_keys` BLOB,
`total_traffic` BIGINT DEFAULT 0,
`peak_bandwidth` INT DEFAULT 0,
`download_count` INT DEFAULT 0,
`enable_vip` BOOLEAN DEFAULT false,
`update_peak_time` DATETIME DEFAULT CURRENT_TIMESTAMP,
PRIMARY KEY (`user_id`)
) ENGINE=InnoDB COMMENT='titan存储服务的用户信息表';

DROP TABLE IF EXISTS `user_asset`;
CREATE TABLE user_asset (
`hash` VARCHAR(128) NOT NULL,
`user_id` VARCHAR(128) NOT NULL,
`area_id` varchar(128) NOT NULL,
`asset_name` VARCHAR(128) DEFAULT '' ,
`asset_type` VARCHAR(128) DEFAULT '' ,
`share_status` TINYINT DEFAULT 0,
`created_time` DATETIME DEFAULT CURRENT_TIMESTAMP,
`total_size` BIGINT DEFAULT 0,
`expiration` DATETIME DEFAULT CURRENT_TIMESTAMP,
`password` VARCHAR(128) DEFAULT '' ,		
`group_id` INT DEFAULT 0,
`is_sync` BOOLEAN DEFAULT true,
PRIMARY KEY (`hash`,`user_id`,`area_id`),
KEY `idx_user_id` (`user_id`),
KEY `idx_group_id` (`group_id`)
) ENGINE=InnoDB COMMENT='titan存储服务的用户文件表';

DROP TABLE IF EXISTS `user_asset_group`;
CREATE TABLE user_asset_group (
`id` INT UNSIGNED AUTO_INCREMENT,
`user_id` VARCHAR(128) NOT NULL,
`name` VARCHAR(32) DEFAULT '',
`parent` INT DEFAULT 0,
`created_time` DATETIME DEFAULT CURRENT_TIMESTAMP,
PRIMARY KEY (`id`),
KEY `idx_user_id` (`user_id`),
KEY `idx_parent` (`parent`)
) ENGINE=InnoDB COMMENT='titan存储服务的用户文件组表';

DROP TABLE IF EXISTS `asset_visit_count`;
CREATE TABLE asset_visit_count (
`hash` VARCHAR(128) NOT NULL,
`count` INT DEFAULT 0,
PRIMARY KEY (`hash`)
) ENGINE=InnoDB COMMENT='titan存储服务的文件被访问次数表';


