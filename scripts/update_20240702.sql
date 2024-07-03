

alter table device_info add column `yesterday_online_time` float NOT NULL DEFAULT 0 after `today_online_time`;
alter table device_info add column `online_incentive_profit` DECIMAL(14, 6) NOT NULL DEFAULT 0;
alter table users add column `online_incentive_reward` DECIMAL(14, 6) NOT NULL DEFAULT 0;


ALTER TABLE `device_info` ADD INDEX `idx_yesterday_online_node_type`(`yesterday_online_time`, `node_type`) USING BTREE;


CREATE TABLE `device_online_incentive` (
`device_id` VARCHAR(128) NOT NULL DEFAULT '',
`user_id` VARCHAR(128) NOT NULL DEFAULT '',
`reward` DECIMAL(14, 6) NOT NULL DEFAULT 0,
`online_time` double NOT NULL DEFAULT '0',
`date` DATETIME(3) NOT NULL DEFAULT 0,
`created_at` DATETIME(3) NOT NULL DEFAULT 0,
KEY `idx_user_id` (`user_id`) USING BTREE,
UNIQUE KEY `uniq_device_date` (`device_id`, `date`) USING BTREE
) ENGINE = INNODB CHARSET = utf8mb4;

