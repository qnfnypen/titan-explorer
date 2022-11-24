DROP TABLE
    IF
    EXISTS `users`;
CREATE TABLE `users` (
`id` BIGINT ( 20 ) NOT NULL AUTO_INCREMENT,
`uuid` VARCHAR ( 255 ) NOT NULL DEFAULT '',
`username` VARCHAR ( 255 ) NOT NULL DEFAULT '',
`pass_hash` VARCHAR ( 255 ) NOT NULL DEFAULT '',
`user_email` VARCHAR ( 255 ) NOT NULL DEFAULT '',
`address` VARCHAR ( 255 ) NOT NULL DEFAULT '',
`role` TINYINT ( 4 ) NOT NULL DEFAULT 0,
`created_at` DATETIME ( 3 ) NOT NULL DEFAULT 0,
`updated_at` DATETIME ( 3 ) NOT NULL DEFAULT 0,
`deleted_at` DATETIME ( 3 ) NOT NULL DEFAULT 0,
PRIMARY KEY ( `id` )
) ENGINE = INNODB DEFAULT CHARSET = utf8mb4;
DROP TABLE
    IF
    EXISTS `login_log`;
CREATE TABLE `login_log` (
    `id` BIGINT ( 20 ) NOT NULL AUTO_INCREMENT,
    `login_username` VARCHAR ( 50 ) NOT NULL DEFAULT '',
    `ip_address` VARCHAR ( 50 ) NOT NULL DEFAULT '',
    `login_location` VARCHAR ( 255 ) NOT NULL DEFAULT '',
    `browser` VARCHAR ( 50 ) NOT NULL DEFAULT '',
    `os` VARCHAR ( 50 ) NOT NULL DEFAULT '',
    `status` TINYINT ( 4 ) NOT NULL DEFAULT 0,
    `msg` VARCHAR ( 255 ) NOT NULL DEFAULT '',
    `created_at` DATETIME ( 3 ) NOT NULL DEFAULT 0,
    PRIMARY KEY ( `id` ) USING BTREE
) ENGINE = INNODB CHARACTER
SET = utf8mb4;
DROP TABLE
    IF
    EXISTS `operation_log`;
CREATE TABLE `operation_log` (
        `id` BIGINT ( 20 ) UNSIGNED NOT NULL AUTO_INCREMENT,
        `title` VARCHAR ( 50 ) NOT NULL DEFAULT '',
        `business_type` INT ( 2 ) NOT NULL DEFAULT 0,
        `method` VARCHAR ( 100 ) NOT NULL DEFAULT '',
        `request_method` VARCHAR ( 10 ) NOT NULL DEFAULT '',
        `operator_type` INT ( 1 ) NOT NULL DEFAULT 0,
        `operator_username` VARCHAR ( 50 ) NOT NULL DEFAULT '',
        `operator_url` VARCHAR ( 500 ) NOT NULL DEFAULT '',
        `operator_ip` VARCHAR ( 50 ) NOT NULL DEFAULT '',
        `operator_location` VARCHAR ( 255 ) NOT NULL DEFAULT '',
        `operator_param` VARCHAR ( 2000 ) NOT NULL DEFAULT '',
        `json_result` VARCHAR ( 2000 ) NOT NULL DEFAULT '',
        `status` INT ( 1 ) NOT NULL DEFAULT 0,
        `error_msg` VARCHAR ( 2000 ) NOT NULL DEFAULT '',
        `created_at` DATETIME ( 6 ) NOT NULL DEFAULT 0,
        `updated_at` DATETIME ( 6 ) NOT NULL DEFAULT 0,
        PRIMARY KEY ( `id` ) USING BTREE
) ENGINE = INNODB CHARACTER
SET = utf8mb4;
DROP TABLE
    IF
    EXISTS `schedulers`;
CREATE TABLE `schedulers` (
     `id` BIGINT ( 20 ) NOT NULL AUTO_INCREMENT,
     `name` VARCHAR ( 255 ) NOT NULL DEFAULT '',
     `group` VARCHAR ( 255 ) NOT NULL DEFAULT '',
     `address` VARCHAR ( 255 ) NOT NULL DEFAULT '',
     `status` INT ( 1 ) NOT NULL DEFAULT 0,
     `token` VARCHAR ( 255 ) NOT NULL DEFAULT '',
     `created_at` DATETIME ( 3 ) NOT NULL DEFAULT 0,
     `updated_at` DATETIME ( 3 ) NOT NULL DEFAULT 0,
     `deleted_at` DATETIME ( 3 ) NOT NULL DEFAULT 0,
     PRIMARY KEY ( `id` )
) ENGINE = INNODB DEFAULT CHARSET = utf8mb4;
DROP TABLE
    IF
    EXISTS `device_info`;
CREATE TABLE `device_info` (
      `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
      `created_at` DATETIME ( 3 ) NOT NULL DEFAULT 0,
      `updated_at` DATETIME ( 3 ) NOT NULL DEFAULT 0,
      `deleted_at` DATETIME ( 3 ) NOT NULL DEFAULT 0,
      `device_id` VARCHAR ( 191 ) CHARACTER
          SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
      `secret` VARCHAR ( 191 ) CHARACTER
          SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
      `node_type` INT ( 2 ) NOT NULL DEFAULT 0,
      `device_name` CHAR ( 56 ) CHARACTER
          SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
      `user_id` VARCHAR ( 191 ) CHARACTER
          SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
      `sn_code` VARCHAR ( 191 ) CHARACTER
          SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
      `operator` VARCHAR ( 191 ) CHARACTER
          SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
      `network_type` VARCHAR ( 191 ) CHARACTER
          SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
      `today_income` DOUBLE NOT NULL DEFAULT 0,
      `yesterday_income` DOUBLE NOT NULL DEFAULT 0,
      `cumu_profit` DOUBLE NOT NULL DEFAULT 0,
      `system_version` VARCHAR ( 191 ) CHARACTER
          SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
      `product_type` VARCHAR ( 191 ) CHARACTER
          SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
      `network_info` VARCHAR ( 191 ) CHARACTER
          SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
      `external_ip` VARCHAR ( 191 ) CHARACTER
          SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
      `internal_ip` VARCHAR ( 191 ) CHARACTER
          SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
      `ip_location` VARCHAR ( 191 ) CHARACTER
          SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
      `mac_location` VARCHAR ( 191 ) CHARACTER
          SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
      `nat_type` VARCHAR ( 191 ) CHARACTER
          SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
      `upnp` VARCHAR ( 191 ) CHARACTER
          SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
      `pkg_loss_ratio` FLOAT ( 32 ) NOT NULL DEFAULT '0' COMMENT '',
      `nat_ratio` FLOAT ( 32 ) NOT NULL DEFAULT '0' COMMENT 'Nat',
      `latency` FLOAT ( 32 ) NOT NULL DEFAULT '0' COMMENT '',
      `cpu_usage` FLOAT ( 32 ) NOT NULL DEFAULT '0' COMMENT '',
      `memory_usage` FLOAT ( 32 ) NOT NULL DEFAULT '0' COMMENT '',
      `disk_usage` FLOAT ( 32 ) NOT NULL DEFAULT '0' COMMENT '',
      `work_status` CHAR ( 28 ) CHARACTER
          SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
      `device_status` CHAR ( 28 ) CHARACTER
          SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
      `disk_type` CHAR ( 28 ) CHARACTER
          SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
      `io_system` VARCHAR ( 191 ) CHARACTER
          SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
      `online_time` FLOAT ( 32 ) NOT NULL DEFAULT '0',
      `today_online_time` FLOAT ( 32 ) NOT NULL DEFAULT '0',
      `today_profit` FLOAT ( 32 ) NOT NULL DEFAULT '0',
      `seven_days_profit` FLOAT ( 32 ) NOT NULL DEFAULT '0',
      `month_profit` FLOAT ( 32 ) NOT NULL DEFAULT '0',
      `bandwidth_up` FLOAT ( 32 ) NOT NULL DEFAULT '0',
      `bandwidth_down` FLOAT ( 32 ) NOT NULL DEFAULT '0',
      PRIMARY KEY ( `id` ) USING BTREE,
      INDEX `idx_device_info_deleted_at` ( `deleted_at` ASC ) USING BTREE
) ENGINE = INNODB AUTO_INCREMENT = 1 CHARACTER
SET = utf8mb4;
DROP TABLE
    IF
    EXISTS `task_info`;
CREATE TABLE `task_info` (
    `id` BIGINT ( 20 ) NOT NULL AUTO_INCREMENT,
    `created_at` DATETIME ( 3 ) NOT NULL DEFAULT 0,
    `updated_at` DATETIME ( 3 ) NOT NULL DEFAULT 0,
    `deleted_at` DATETIME ( 3 ) NOT NULL DEFAULT 0,
    `user_id` VARCHAR ( 128 ) NOT NULL DEFAULT '',
    `miner_id` VARCHAR ( 128 ) NOT NULL DEFAULT '',
    `device_id` VARCHAR ( 128 ) NOT NULL DEFAULT '',
    `file_name` VARCHAR ( 128 ) NOT NULL DEFAULT '',
    `ip_address` VARCHAR ( 32 ) NOT NULL DEFAULT '',
    `cid` VARCHAR ( 128 ) NOT NULL DEFAULT '',
    `bandwidth_up` VARCHAR ( 32 ) NOT NULL DEFAULT '',
    `bandwidth_down` VARCHAR ( 32 ) NOT NULL DEFAULT '',
    `time_need` VARCHAR ( 32 ) NOT NULL DEFAULT '',
    `time` TIMESTAMP NOT NULL DEFAULT 0,
    `service_country` VARCHAR ( 56 ) NOT NULL DEFAULT '',
    `region` VARCHAR ( 56 ) NOT NULL DEFAULT '',
    `status` VARCHAR ( 56 ) NOT NULL DEFAULT '',
    `price` FLOAT ( 32 ) NOT NULL DEFAULT '0',
    `file_size` FLOAT ( 32 ) NOT NULL DEFAULT '0',
    `download_url` VARCHAR ( 256 ) NOT NULL DEFAULT '',
    PRIMARY KEY ( `id` )
) ENGINE = INNODB DEFAULT CHARSET = utf8;
DROP TABLE
    IF
    EXISTS `income_daily`;
CREATE TABLE `income_daily` (
       `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
       `created_at` TIMESTAMP NOT NULL DEFAULT 0,
       `updated_at` TIMESTAMP NOT NULL DEFAULT 0,
       `deleted_at` DATETIME ( 3 ) NOT NULL DEFAULT 0,
       `user_id` VARCHAR ( 128 ) NOT NULL DEFAULT '',
       `device_id` VARCHAR ( 128 ) NOT NULL DEFAULT '',
       `time` TIMESTAMP NOT NULL DEFAULT 0,
       `income` FLOAT ( 32 ) NOT NULL DEFAULT '0',
       `online_time` FLOAT ( 32 ) NOT NULL DEFAULT '0',
       `pkg_loss_ratio` FLOAT ( 32 ) NOT NULL DEFAULT '0',
       `latency` FLOAT ( 32 ) NOT NULL DEFAULT '0',
       `nat_ratio` FLOAT ( 32 ) NOT NULL DEFAULT '0',
       `disk_usage` FLOAT ( 32 ) NOT NULL DEFAULT '0',
       PRIMARY KEY ( `id` ) USING BTREE,
       INDEX `idx_income_daily_deleted_at` ( `deleted_at` ASC ) USING BTREE
) ENGINE = INNODB DEFAULT CHARSET = utf8;
DROP TABLE
    IF
    EXISTS `hour_daily`;
CREATE TABLE `hour_daily` (
     `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
     `created_at` TIMESTAMP NOT NULL DEFAULT 0,
     `updated_at` TIMESTAMP NOT NULL DEFAULT 0,
     `deleted_at` DATETIME ( 3 ) NOT NULL DEFAULT 0,
     `user_id` VARCHAR ( 128 ) NOT NULL DEFAULT '',
     `device_id` VARCHAR ( 128 ) NOT NULL DEFAULT '',
     `time` TIMESTAMP NOT NULL DEFAULT 0,
     `hour_income` FLOAT ( 32 ) NOT NULL DEFAULT '0',
     `online_time` FLOAT ( 32 ) NOT NULL DEFAULT '0',
     `pkg_loss_ratio` FLOAT ( 32 ) NOT NULL DEFAULT '0',
     `latency` FLOAT ( 32 ) NOT NULL DEFAULT '0',
     `nat_ratio` FLOAT ( 32 ) NOT NULL DEFAULT '0',
     `disk_usage` FLOAT ( 32 ) NOT NULL DEFAULT '0',
     PRIMARY KEY ( `id` ) USING BTREE,
     INDEX `idx_hour_daily_deleted_at` ( `deleted_at` ASC ) USING BTREE
) ENGINE = INNODB DEFAULT CHARSET = utf8;
DROP TABLE
    IF
    EXISTS `retrieval_info`;
CREATE TABLE `retrieval_info` (
         `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
         `created_at` DATETIME ( 3 ) NOT NULL DEFAULT 0,
         `updated_at` DATETIME ( 3 ) NOT NULL DEFAULT 0,
         `deleted_at` DATETIME ( 3 ) NOT NULL DEFAULT 0,
         `service_country` VARCHAR ( 191 ) CHARACTER
             SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
         `service_status` VARCHAR ( 191 ) CHARACTER
             SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
         `task_status` CHAR ( 56 ) CHARACTER
             SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
         `file_name` VARCHAR ( 191 ) CHARACTER
             SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
         `file_size` VARCHAR ( 191 ) CHARACTER
             SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
         `create_time` VARCHAR ( 191 ) CHARACTER
             SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
         `cid` VARCHAR ( 191 ) CHARACTER
             SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
         `price` DOUBLE NOT NULL DEFAULT 0,
         `miner_id` CHAR ( 50 ) CHARACTER
             SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
         `user_id` CHAR ( 56 ) CHARACTER
             SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
         PRIMARY KEY ( `id` ) USING BTREE,
         INDEX `idx_retrieval_info_deleted_at` ( `deleted_at` ASC ) USING BTREE
) ENGINE = INNODB AUTO_INCREMENT = 6 CHARACTER
SET = utf8mb4 COLLATE = utf8mb4_general_ci;