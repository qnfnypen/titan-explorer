
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