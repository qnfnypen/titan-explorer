DROP TABLE IF EXISTS `bugs`;
CREATE TABLE `bugs` (
`id` BIGINT(20) NOT NULL AUTO_INCREMENT,
`code` VARCHAR(255) NOT NULL DEFAULT '',
`username`  VARCHAR(255) NOT NULL DEFAULT '',
`node_id` VARCHAR(255) NOT NULL DEFAULT '',
`email` VARCHAR(255) NOT NULL DEFAULT '',
`description` varchar(1024) not null COMMENT 'bug description',
`telegram_id` VARCHAR(255) NOT NULL DEFAULT '',
`feedback_type` tinyint(1) NOT NULL DEFAULT 0 COMMENT '1:意见 2:咨询 3:异常 4:其他',
`feedback` text not null COMMENT 'feedback content',
`pics` text NOT NULL COMMENT 'upload pictures url arr',
`log` MEDIUMTEXT NOT NULL COMMENT 'log',
`platform` tinyint(1) NOT NULL DEFAULT 0 COMMENT '1macos 2windows 3android 4ios',
`version` varchar(255) NOT NULL DEFAULT '',
`state` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '1waiting 2done',
`reward` int(8) NOT NULL DEFAULT 0,
`reward_type` CHAR(10) NOT NULL DEFAULT 'tnt2',
`operator` VARCHAR(255) NOT NULL DEFAULT '',
`created_at` DATETIME(3) NOT NULL DEFAULT 0,
`updated_at` DATETIME(3) NOT NULL DEFAULT 0,
PRIMARY KEY (`id`)
) ENGINE = INNODB CHARSET = utf8mb4;