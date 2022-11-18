DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `uuid` longtext,
  `username` longtext,
  `pass_hash` longtext,
  `user_email` longtext,
  `address` longtext,
  `role` tinyint(4) NOT NULL DEFAULT 0,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


DROP TABLE IF EXISTS `login_log`;
CREATE TABLE `login_log`  (
`id` bigint(20) NOT NULL AUTO_INCREMENT,
`login_username` varchar(50) NULL DEFAULT '',
`ipaddr` varchar(50)  NULL DEFAULT '',
`login_location` varchar(255)  NULL DEFAULT '',
`browser` varchar(50)  NULL DEFAULT '',
`os` varchar(50)  NULL DEFAULT '',
`status` tinyint(4) NULL DEFAULT 0,
`msg` varchar(255)  NULL DEFAULT '',
`created_at` datetime(3) DEFAULT NULL,
PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4;

DROP TABLE IF EXISTS `operation_log`;
CREATE TABLE `operation_log`  (
 `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
 `title` varchar(50)  NULL DEFAULT '',
 `business_type` int(2) NULL DEFAULT 0 ,
 `method` varchar(100)  NULL DEFAULT '',
 `request_method` varchar(10)  NULL DEFAULT '',
 `operator_type` int(1) NULL DEFAULT 0,
 `operator_username` varchar(50)  NULL DEFAULT '',
 `operator_url` varchar(500)  NULL DEFAULT '',
 `operator_ip` varchar(50)  NULL DEFAULT '',
 `operator_location` varchar(255)  NULL DEFAULT '',
 `operator_param` text  NULL,
 `json_result` text  NULL,
 `status` int(1) NULL DEFAULT 0,
 `error_msg` varchar(2000)  NULL DEFAULT '',
 `created_at` datetime(6) DEFAULT NULL,
 `updated_at` datetime(6) DEFAULT NULL,
 PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4;


DROP TABLE IF EXISTS `schedulers`;
CREATE TABLE `schedulers` (
`id` bigint(20) NOT NULL AUTO_INCREMENT,
`name` longtext,
`group` longtext,
`address` longtext,
`status` int(1) NULL DEFAULT 0,
`created_at` datetime(3) DEFAULT NULL,
`updated_at` datetime(3) DEFAULT NULL,
`deleted_at` datetime(3) DEFAULT NULL,
PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


DROP TABLE IF EXISTS `device_info`;
CREATE TABLE `device_info`  (
    `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
    `created_at` datetime(3) NULL DEFAULT NULL,
    `updated_at` datetime(3) NULL DEFAULT NULL,
    `deleted_at` datetime(3) NULL DEFAULT NULL,
    `device_id` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
    `secret` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
    `node_type` int(2)  NULL DEFAULT 0,
    `device_name` char(56) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
    `user_id` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
    `sn_code` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
    `operator` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
    `network_type` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
    `today_income` double NULL DEFAULT NULL,
    `yesterday_income` double NULL DEFAULT NULL,
    `cumu_profit` double NULL DEFAULT NULL,
    `system_version` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
    `product_type` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
    `network_info` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
    `external_ip` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
    `internal_ip` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
    `ip_location` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
    `mac_location` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
    `nat_type` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
    `upnp` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
    `pkg_loss_ratio` float(32) NOT NULL DEFAULT '0' COMMENT '',
    `nat_ratio` float(32) NOT NULL DEFAULT '0' COMMENT 'Nat',
    `latency` float(32) NOT NULL DEFAULT '0' COMMENT '',
    `cpu_usage` float(32) NOT NULL DEFAULT '0' COMMENT '',
    `memory_usage` float(32) NOT NULL DEFAULT '0' COMMENT '',
    `disk_usage` float(32) NOT NULL DEFAULT '0' COMMENT '',
    `work_status` char(28) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
    `device_status` char(28) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
    `disk_type` char(28) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
    `io_system` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
    `today_online_time` float(32) NOT NULL DEFAULT '0',
    `today_profit` float(32) NOT NULL DEFAULT '0' ,
    `seven_days_profit` float(32) NOT NULL DEFAULT '0',
    `month_profit` float(32) NOT NULL DEFAULT '0',
    `bandwidth_up` float(32) NOT NULL DEFAULT '0' ,
    `bandwidth_down` float(32) NOT NULL DEFAULT '0',
    PRIMARY KEY (`id`) USING BTREE,
    INDEX `idx_device_info_deleted_at`(`deleted_at` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4;


DROP TABLE IF EXISTS `task_info`;
CREATE TABLE `task_info` (
`id` bigint(20) NOT NULL AUTO_INCREMENT ,
`created_at` datetime(3) NULL DEFAULT NULL,
`updated_at` datetime(3) NULL DEFAULT NULL,
`deleted_at` datetime(3) NULL DEFAULT NULL,
`user_id` varchar(128) NOT NULL DEFAULT '',
`miner_id` varchar(128) NOT NULL DEFAULT '' ,
`device_id` varchar(128) NOT NULL DEFAULT '' ,
`file_name` varchar(128) NOT NULL DEFAULT '',
`ip_address` varchar(32) NOT NULL DEFAULT '' ,
`cid` varchar(128) NOT NULL DEFAULT '' ,
`bandwidth_up` varchar(32) NOT NULL DEFAULT '',
`bandwidth_down` varchar(32) NOT NULL DEFAULT '',
`time_need` varchar(32) NOT NULL DEFAULT '',
`time` timestamp  NULL DEFAULT NULL,
`service_country` varchar(56) NOT NULL DEFAULT '',
`region` varchar(56) NOT NULL DEFAULT '',
`status` varchar(56) NOT NULL DEFAULT '',
`price` float(32) NOT NULL DEFAULT '0',
`file_size` float(32) NOT NULL DEFAULT '0',
`download_url` varchar(256) NOT NULL DEFAULT '',
PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


DROP TABLE IF EXISTS `income_daily`;
CREATE TABLE `income_daily`  (
   `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
   `created_at` timestamp NULL DEFAULT NULL ,
   `updated_at` timestamp NULL DEFAULT NULL ,
   `deleted_at` datetime(3) NULL DEFAULT NULL,
   `user_id` varchar(128) NOT NULL DEFAULT '' ,
   `device_id` varchar(128) NOT NULL DEFAULT '',
   `time` timestamp  NULL DEFAULT NULL ,
   `income` float(32) NOT NULL DEFAULT '0' ,
   `online_time` float(32) NOT NULL DEFAULT '0' ,
   `pkg_loss_ratio` float(32) NOT NULL DEFAULT '0',
   `latency` float(32) NOT NULL DEFAULT '0' ,
   `nat_ratio` float(32) NOT NULL DEFAULT '0' ,
   `disk_usage` float(32) NOT NULL DEFAULT '0' ,
   PRIMARY KEY (`id`) USING BTREE,
   INDEX `idx_income_daily_deleted_at`(`deleted_at` ASC) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

DROP TABLE IF EXISTS `hour_daily`;
CREATE TABLE `hour_daily`  (
 `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
 `created_at` timestamp NULL DEFAULT NULL,
 `updated_at` timestamp NULL DEFAULT NULL,
 `deleted_at` datetime(3) NULL DEFAULT NULL,
 `user_id` varchar(128) NOT NULL DEFAULT '' ,
 `device_id` varchar(128) NOT NULL DEFAULT '' ,
 `time` timestamp NULL DEFAULT NULL ,
 `hour_income` float(32) NOT NULL DEFAULT '0' ,
 `online_time` float(32) NOT NULL DEFAULT '0' ,
 `pkg_loss_ratio` float(32) NOT NULL DEFAULT '0' ,
 `latency` float(32) NOT NULL DEFAULT '0' ,
 `nat_ratio` float(32) NOT NULL DEFAULT '0',
 `disk_usage` float(32) NOT NULL DEFAULT '0',
 PRIMARY KEY (`id`) USING BTREE,
 INDEX `idx_hour_daily_deleted_at`(`deleted_at` ASC) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


DROP TABLE IF EXISTS `retrieval_info`;
CREATE TABLE `retrieval_info`  (
`id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
`created_at` datetime(3) NULL DEFAULT NULL,
`updated_at` datetime(3) NULL DEFAULT NULL,
`deleted_at` datetime(3) NULL DEFAULT NULL,
`service_country` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
`service_status` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
`task_status` char(56) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
`file_name` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
`file_size` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
`create_time` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
`cid` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
`price` double NULL DEFAULT NULL,
`miner_id` char(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
`user_id` char(56) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
PRIMARY KEY (`id`) USING BTREE,
INDEX `idx_retrieval_info_deleted_at`(`deleted_at` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 6 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci;
