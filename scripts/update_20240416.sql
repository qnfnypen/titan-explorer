
DROP TABLE IF EXISTS `referral_code`;
CREATE TABLE referral_code (
`id` bigint(20) NOT NULL AUTO_INCREMENT,
`user_id`  VARCHAR(128) NOT NULL DEFAULT '',
`code` VARCHAR(128) NOT NULL DEFAULT '',
`created_at` DATETIME(3) NOT NULL DEFAULT 0,
`updated_at` DATETIME(3) NOT NULL DEFAULT 0,
PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


DROP TABLE IF EXISTS `kol`;
CREATE TABLE `kol` (
 `id` BIGINT(20) NOT NULL AUTO_INCREMENT,
 `user_id`  VARCHAR(128) NOT NULL DEFAULT '',
 `level` INT(4) NOT NULL DEFAULT 0,
 `comment` VARCHAR(128) NOT NULL DEFAULT '',, ,
 `status` INT(1) NOT NULL DEFAULT 0,
 `created_at` DATETIME(3) NOT NULL DEFAULT 0,
 `updated_at` DATETIME(3) NOT NULL DEFAULT 0,
 PRIMARY KEY (`id`)
) ENGINE = INNODB CHARSET = utf8mb4;


DROP TABLE IF EXISTS `kol_level_conf`;
CREATE TABLE `kol_level_conf` (
`id` BIGINT(20) NOT NULL AUTO_INCREMENT,
`level` INT(4) NOT NULL DEFAULT 0,
`parent_commission_percent` INT(8) NOT NULL DEFAULT 0,
`children_bonus_percent` INT(8) NOT NULL DEFAULT 0,
`user_threshold` INT NOT NULL DEFAULT 0,
`device_threshold` INT NOT NULL DEFAULT 0,
`status` INT(1) NOT NULL DEFAULT 0,
`created_at` DATETIME(3) NOT NULL DEFAULT 0,
`updated_at` DATETIME(3) NOT NULL DEFAULT 0,
PRIMARY KEY (`id`)
) ENGINE = INNODB CHARSET = utf8mb4;


DROP TABLE IF EXISTS `data_collection`;
CREATE TABLE `data_collection` (
 `id` BIGINT(20) NOT NULL AUTO_INCREMENT,
 `event` INT(4) NOT NULL DEFAULT 0,
 `value` VARCHAR(128) NOT NULL DEFAULT '',
 `os` VARCHAR(128) NOT NULL DEFAULT '',
 `url` VARCHAR(128) NOT NULL DEFAULT '',
 `ip` VARCHAR(128) NOT NULL DEFAULT '',
 `created_at` DATETIME(3) NOT NULL DEFAULT 0,
 PRIMARY KEY (`id`)
) ENGINE = INNODB CHARSET = utf8mb4;


DROP TABLE IF EXISTS `user_reward_daily`;
CREATE TABLE `user_reward_daily` (
`user_id` VARCHAR(128) NOT NULL DEFAULT '',
`cumulative_reward` DECIMAL(14, 6) NOT NULL DEFAULT 0,
`reward` DECIMAL(14, 6) NOT NULL DEFAULT 0,
`app_reward` DECIMAL(14, 6) NOT NULL DEFAULT 0,
`cli_reward` DECIMAL(14, 6) NOT NULL DEFAULT 0,
`device_online_count` INT(20) NOT NULL DEFAULT 0,
`total_device_count` INT(20) NOT NULL DEFAULT 0,
`kol_bonus` DECIMAL(14, 6) NOT NULL DEFAULT 0,
`referral_reward` DECIMAL(14, 6) NOT NULL DEFAULT 0,
`referrer_user_id` VARCHAR(128) NOT NULL DEFAULT '',
`is_kol` INT(4) NOT NULL DEFAULT 0,
`is_referrer_kol` INT(4) NOT NULL DEFAULT 0,
`referrer_reward` DECIMAL(14, 6) NOT NULL DEFAULT 0,
`commission_percent`  INT(8) NOT NULL DEFAULT 0,
`kol_bonus_percent` INT(8) NOT NULL DEFAULT 0,
`time` DATETIME(3) NOT NULL DEFAULT 0,
`created_at` DATETIME(3) NOT NULL DEFAULT 0,
`updated_at` DATETIME(3) NOT NULL DEFAULT 0,
UNIQUE KEY `uniq_user_time` (`user_id`,`time`) USING BTREE,
KEY `idx_user_id` (`user_id`) USING BTREE,
KEY `idx_referrer_user_id` (`referrer_user_id`) USING BTREE
) ENGINE = INNODB CHARSET = utf8mb4;


CREATE TABLE `kol_level_up_record` (
`id` BIGINT(20) NOT NULL AUTO_INCREMENT,
`user_id` VARCHAR(128) NOT NULL DEFAULT '',
`before_level` int(4) NOT NULL DEFAULT 0,
`after_level` int(4) NOT NULL DEFAULT 0,
`referral_users_count` INT(20) NOT NULL DEFAULT 0,
`referral_nodes_count` INT(20) NOT NULL DEFAULT 0,
`created_at` DATETIME(3) NOT NULL DEFAULT 0,
PRIMARY KEY (`id`)
) ENGINE = INNODB CHARSET = utf8mb4;


insert into referral_code(user_id, code, created_at)  select username as user_id, referral_code as code, created_at from users ;vfb

alter table users add column `referrer_commission_reward` DECIMAL(14, 6) NOT NULL DEFAULT 0 after `device_count`;
alter table users add column `from_kol_bonus_reward` DECIMAL(14, 6) NOT NULL DEFAULT 0 after `referrer_commission_reward`;
alter table users add column `device_online_count` INT(20) NOT NULL DEFAULT 0 after `device_count`;

alter table device_info rename column `is_mobile` to `app_type`;
-- alter table device_info_daily drop column is_mobile;


alter table user_reward_daily rename column `mobile_reward` to `app_reward`;
alter table user_reward_daily rename column `pc_reward` to `cli_reward`;