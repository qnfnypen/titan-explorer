
ALTER TABLE kol_level_up_record RENAME TO kol_change_log;

drop table kol_level_conf;

CREATE TABLE `kol_level_config` (
 `id` bigint(20) NOT NULL AUTO_INCREMENT,
 `level` int(4) NOT NULL DEFAULT 0,
 `commission_percent` double NOT NULL DEFAULT 0,
 `parent_commission_percent` double NOT NULL DEFAULT 0,
 `device_threshold` int(11) NOT NULL DEFAULT 0,
 `status` int(1) NOT NULL DEFAULT 0,
 `created_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
 `updated_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
 PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8mb4;


CREATE TABLE `user_reward_detail` (
`user_id` VARCHAR(128) NOT NULL DEFAULT '',
`from_user_id` VARCHAR(128) NOT NULL DEFAULT '',
`reward` DECIMAL(14, 6) NOT NULL DEFAULT 0,
`relationship` int(4) NOT NULL DEFAULT 0,
`created_at` DATETIME(3) NOT NULL DEFAULT 0,
`updated_at` DATETIME(3) NOT NULL DEFAULT 0,
KEY `idx_user_id` (`user_id`) USING BTREE,
UNIQUE KEY `uniq_user_fu` (`user_id`, `from_user_id`) USING BTREE
) ENGINE = INNODB CHARSET = utf8mb4;


CREATE TABLE `user_reward_log` (
`user_id` VARCHAR(128) NOT NULL DEFAULT '',
`from_user_id` VARCHAR(128) NOT NULL DEFAULT '',
`reward` DECIMAL(14, 6) NOT NULL DEFAULT 0,
`relationship` int(4) NOT NULL DEFAULT 0,
`created_at` DATETIME(3) NOT NULL DEFAULT 0,
`updated_at` DATETIME(3) NOT NULL DEFAULT 0,
KEY `idx_user_id` (`user_id`) USING BTREE,
UNIQUE KEY `uniq_user_fu` (`user_id`, `from_user_id`) USING BTREE
) ENGINE = INNODB CHARSET = utf8mb4;


ALTER TABLE users drop column referral_code;
ALTER TABLE users drop column payout;
ALTER TABLE users drop column frozen_reward;
ALTER TABLE users drop column device_online_count;
ALTER TABLE users drop column referrer_commission_reward;
ALTER TABLE users drop column from_kol_bonus_reward;
ALTER TABLE users ADD COLUMN cassini_reward DECIMAL(20, 6) NOT NULL DEFAULT 0 AFTER herschel_referral_reward;
ALTER TABLE users ADD COLUMN cassini_referral_reward DECIMAL(20, 6) NOT NULL DEFAULT 0 AFTER cassini_reward;