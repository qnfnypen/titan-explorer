/*
 Navicat Premium Data Transfer

 Source Server         : huygens
 Source Server Type    : MySQL
 Source Server Version : 50743
 Source Host           : rm-j6c91ajsgtef7181g.mysql.rds.aliyuncs.com:3306
 Source Schema         : titan_explorer

 Target Server Type    : MySQL
 Target Server Version : 50743
 File Encoding         : 65001

 Date: 18/03/2024 10:16:39
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for application
-- ----------------------------
DROP TABLE IF EXISTS `application`;
CREATE TABLE `application` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` varchar(128) NOT NULL DEFAULT '',
  `email` varchar(128) NOT NULL DEFAULT '',
  `area_id` varchar(64) NOT NULL DEFAULT '',
  `ip_country` varchar(128) NOT NULL DEFAULT '',
  `ip_city` varchar(128) NOT NULL DEFAULT '',
  `public_key` varchar(2048) NOT NULL DEFAULT '',
  `node_type` tinyint(4) NOT NULL DEFAULT '0',
  `num` tinyint(4) NOT NULL DEFAULT '0',
  `amount` int(20) NOT NULL DEFAULT '0',
  `upstream_bandwidth` double NOT NULL DEFAULT '0',
  `disk_space` double NOT NULL DEFAULT '0',
  `ip` varchar(128) NOT NULL DEFAULT '',
  `status` tinyint(4) NOT NULL DEFAULT '0',
  `created_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `updated_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Table structure for application_result
-- ----------------------------
DROP TABLE IF EXISTS `application_result`;
CREATE TABLE `application_result` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `application_id` bigint(20) NOT NULL DEFAULT '0',
  `user_id` varchar(128) NOT NULL DEFAULT '',
  `device_id` varchar(128) NOT NULL DEFAULT '',
  `node_type` tinyint(4) NOT NULL DEFAULT '0',
  `secret` varchar(256) NOT NULL DEFAULT '0',
  `created_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `updated_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Table structure for assets
-- ----------------------------
DROP TABLE IF EXISTS `assets`;
CREATE TABLE `assets` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `node_id` varchar(255) NOT NULL DEFAULT '',
  `event` bigint(20) NOT NULL DEFAULT '0',
  `cid` varchar(255) NOT NULL DEFAULT '',
  `hash` varchar(255) NOT NULL DEFAULT '',
  `total_size` bigint(20) NOT NULL DEFAULT '0',
  `path` varchar(255) NOT NULL DEFAULT '',
  `end_time` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `expiration` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `user_id` varchar(255) NOT NULL DEFAULT '',
  `type` varchar(255) NOT NULL DEFAULT '',
  `name` varchar(255) NOT NULL DEFAULT '',
  `created_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `updated_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `deleted_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `project_id` bigint(20) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_cid` (`cid`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=96550 DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Table structure for assets_bakup
-- ----------------------------
DROP TABLE IF EXISTS `assets_bakup`;
CREATE TABLE `assets_bakup` (
  `id` bigint(20) NOT NULL DEFAULT '0',
  `node_id` varchar(255) CHARACTER SET utf8mb4 NOT NULL DEFAULT '',
  `event` bigint(20) NOT NULL DEFAULT '0',
  `cid` varchar(255) CHARACTER SET utf8mb4 NOT NULL DEFAULT '',
  `hash` varchar(255) CHARACTER SET utf8mb4 NOT NULL DEFAULT '',
  `total_size` bigint(20) NOT NULL DEFAULT '0',
  `path` varchar(255) CHARACTER SET utf8mb4 NOT NULL DEFAULT '',
  `end_time` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `expiration` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `user_id` varchar(255) CHARACTER SET utf8mb4 NOT NULL DEFAULT '',
  `type` varchar(255) CHARACTER SET utf8mb4 NOT NULL DEFAULT '',
  `name` varchar(255) CHARACTER SET utf8mb4 NOT NULL DEFAULT '',
  `created_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `updated_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `deleted_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `project_id` bigint(20) NOT NULL DEFAULT '0'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

-- ----------------------------
-- Table structure for cache_event
-- ----------------------------
DROP TABLE IF EXISTS `cache_event`;
CREATE TABLE `cache_event` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `device_id` varchar(128) NOT NULL DEFAULT '',
  `carfile_cid` varchar(128) NOT NULL DEFAULT '',
  `block_size` double NOT NULL DEFAULT '0',
  `blocks` bigint(20) NOT NULL DEFAULT '0',
  `replicaInfos` int(20) NOT NULL DEFAULT '0',
  `time` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `status` tinyint(4) NOT NULL DEFAULT '0',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_device_id_car_time` (`device_id`,`carfile_cid`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Table structure for device_info
-- ----------------------------
DROP TABLE IF EXISTS `device_info`;
CREATE TABLE `device_info` (
  `created_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `updated_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `deleted_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `bound_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `device_id` varchar(191) NOT NULL DEFAULT '',
  `node_type` int(11) NOT NULL DEFAULT '0',
  `device_rank` int(11) NOT NULL DEFAULT '0',
  `device_name` char(56) NOT NULL DEFAULT '',
  `user_id` varchar(191) NOT NULL DEFAULT '',
  `system_version` varchar(191) NOT NULL DEFAULT '',
  `network_info` varchar(191) NOT NULL DEFAULT '',
  `external_ip` varchar(28) NOT NULL DEFAULT '',
  `internal_ip` varchar(28) NOT NULL DEFAULT '',
  `ip_location` varchar(191) NOT NULL DEFAULT '',
  `ip_country` varchar(191) NOT NULL DEFAULT '',
  `ip_province` varchar(191) NOT NULL DEFAULT '',
  `ip_city` varchar(191) NOT NULL DEFAULT '',
  `latitude` double NOT NULL DEFAULT '0',
  `longitude` double NOT NULL DEFAULT '0',
  `mac_location` varchar(191) NOT NULL DEFAULT '',
  `cpu_usage` double NOT NULL DEFAULT '0',
  `cpu_cores` int(11) NOT NULL DEFAULT '0',
  `memory_usage` double NOT NULL DEFAULT '0',
  `memory` double NOT NULL DEFAULT '0',
  `disk_usage` double NOT NULL DEFAULT '0',
  `disk_space` double NOT NULL DEFAULT '0',
  `bind_status` char(28) NOT NULL DEFAULT '',
  `device_status` char(28) NOT NULL DEFAULT '',
  `device_status_code` bigint(20) NOT NULL DEFAULT '0',
  `active_status` int(11) NOT NULL DEFAULT '0',
  `disk_type` char(28) NOT NULL DEFAULT '',
  `io_system` varchar(191) NOT NULL DEFAULT '',
  `online_time` double NOT NULL DEFAULT '0',
  `today_online_time` double NOT NULL DEFAULT '0',
  `today_profit` decimal(14,6) DEFAULT '0.000000',
  `yesterday_profit` decimal(14,6) DEFAULT '0.000000',
  `seven_days_profit` decimal(14,6) DEFAULT '0.000000',
  `month_profit` decimal(14,6) DEFAULT '0.000000',
  `cumulative_profit` decimal(14,6) DEFAULT '0.000000',
  `available_profit` decimal(14,6) DEFAULT '0.000000',
  `bandwidth_up` double NOT NULL DEFAULT '0',
  `bandwidth_down` double NOT NULL DEFAULT '0',
  `download_traffic` double NOT NULL DEFAULT '0',
  `upload_traffic` double NOT NULL DEFAULT '0',
  `cache_count` bigint(20) NOT NULL DEFAULT '0',
  `retrieval_count` bigint(20) NOT NULL DEFAULT '0',
  `income_incr` double NOT NULL DEFAULT '0',
  `nat_type` varchar(128) NOT NULL DEFAULT '',
  `area_id` varchar(64) NOT NULL DEFAULT '',
  PRIMARY KEY (`device_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Table structure for device_info_daily
-- ----------------------------
DROP TABLE IF EXISTS `device_info_daily`;
CREATE TABLE `device_info_daily` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  `updated_at` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  `deleted_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `user_id` varchar(128) NOT NULL DEFAULT '',
  `device_id` varchar(128) NOT NULL DEFAULT '',
  `time` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  `income` decimal(14,6) DEFAULT '0.000000',
  `online_time` double NOT NULL DEFAULT '0',
  `pkg_loss_ratio` double NOT NULL DEFAULT '0',
  `latency` double NOT NULL DEFAULT '0',
  `nat_ratio` double NOT NULL DEFAULT '0',
  `disk_usage` double NOT NULL DEFAULT '0',
  `disk_space` double NOT NULL DEFAULT '0',
  `bandwidth_up` double NOT NULL DEFAULT '0',
  `bandwidth_down` double NOT NULL DEFAULT '0',
  `upstream_traffic` double NOT NULL DEFAULT '0',
  `downstream_traffic` double NOT NULL DEFAULT '0',
  `retrieval_count` bigint(20) NOT NULL DEFAULT '0',
  `block_count` bigint(20) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `idx_device_id_time` (`device_id`,`time`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=23509726 DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for device_info_hour
-- ----------------------------
DROP TABLE IF EXISTS `device_info_hour`;
CREATE TABLE `device_info_hour` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  `updated_at` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  `deleted_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `user_id` varchar(128) NOT NULL DEFAULT '',
  `device_id` varchar(128) NOT NULL DEFAULT '',
  `time` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  `hour_income` decimal(14,6) DEFAULT '0.000000',
  `online_time` double NOT NULL DEFAULT '0',
  `pkg_loss_ratio` double NOT NULL DEFAULT '0',
  `latency` double NOT NULL DEFAULT '0',
  `nat_ratio` double NOT NULL DEFAULT '0',
  `disk_usage` double NOT NULL DEFAULT '0',
  `disk_space` double NOT NULL DEFAULT '0',
  `bandwidth_up` double NOT NULL DEFAULT '0',
  `bandwidth_down` double NOT NULL DEFAULT '0',
  `upstream_traffic` double NOT NULL DEFAULT '0',
  `downstream_traffic` double NOT NULL DEFAULT '0',
  `retrieval_count` bigint(20) NOT NULL DEFAULT '0',
  `block_count` bigint(20) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`,`device_id`) USING BTREE,
  UNIQUE KEY `uniq_device_id_time` (`device_id`,`time`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=24652166 DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for fil_storage
-- ----------------------------
DROP TABLE IF EXISTS `fil_storage`;
CREATE TABLE `fil_storage` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `provider` varchar(64) NOT NULL DEFAULT '',
  `sector_num` varchar(64) NOT NULL DEFAULT '',
  `cost` double NOT NULL DEFAULT '0',
  `message_cid` varchar(255) NOT NULL DEFAULT '',
  `piece_cid` varchar(255) NOT NULL DEFAULT '',
  `payload_cid` varchar(255) NOT NULL DEFAULT '',
  `deal_id` varchar(255) NOT NULL DEFAULT '',
  `path` varchar(255) NOT NULL DEFAULT '',
  `f_index` int(20) NOT NULL DEFAULT '0',
  `piece_size` double NOT NULL DEFAULT '0',
  `gas` double NOT NULL DEFAULT '0',
  `pledge` double NOT NULL DEFAULT '0',
  `start_height` bigint(20) NOT NULL DEFAULT '0',
  `end_height` bigint(20) NOT NULL DEFAULT '0',
  `start_time` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `end_time` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `created_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `updated_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `deleted_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_provider_path` (`provider`,`path`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Table structure for full_node_info
-- ----------------------------
DROP TABLE IF EXISTS `full_node_info`;
CREATE TABLE `full_node_info` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `total_node_count` int(20) NOT NULL DEFAULT '0',
  `online_node_count` int(20) NOT NULL DEFAULT '0',
  `t_upstream_file_count` int(20) NOT NULL DEFAULT '0',
  `t_average_replica` double NOT NULL DEFAULT '0',
  `t_node_online_ratio` double NOT NULL DEFAULT '0',
  `f_backups_from_titan` double NOT NULL DEFAULT '0',
  `validator_count` int(20) NOT NULL DEFAULT '0',
  `online_validator_count` bigint(20) NOT NULL DEFAULT '0',
  `candidate_count` int(20) NOT NULL DEFAULT '0',
  `online_candidate_count` bigint(20) NOT NULL DEFAULT '0',
  `edge_count` int(20) NOT NULL DEFAULT '0',
  `memory` bigint(20) NOT NULL DEFAULT '0',
  `cpu_cores` bigint(20) NOT NULL DEFAULT '0',
  `ip_count` bigint(20) NOT NULL DEFAULT '0',
  `online_edge_count` bigint(20) NOT NULL DEFAULT '0',
  `total_storage` double NOT NULL DEFAULT '0',
  `storage_used` double NOT NULL DEFAULT '0',
  `total_upstream_bandwidth` double NOT NULL DEFAULT '0',
  `total_downstream_bandwidth` double NOT NULL DEFAULT '0',
  `total_carfile` bigint(20) NOT NULL DEFAULT '0',
  `total_carfile_size` double NOT NULL DEFAULT '0',
  `retrieval_count` bigint(20) NOT NULL DEFAULT '0',
  `next_election_time` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  `fvm_order_count` int(20) NOT NULL DEFAULT '0',
  `f_node_count` int(20) NOT NULL DEFAULT '0',
  `f_high` int(20) NOT NULL DEFAULT '0',
  `t_next_election_high` int(20) NOT NULL DEFAULT '0',
  `time` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  `created_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `updated_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_time` (`time`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=5440 DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Table structure for link
-- ----------------------------
DROP TABLE IF EXISTS `link`;
CREATE TABLE `link` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `username` varchar(255) NOT NULL DEFAULT '',
  `user_id` varchar(255) NOT NULL DEFAULT '',
  `cid` varchar(255) NOT NULL DEFAULT '',
  `long_link` varchar(1024) NOT NULL DEFAULT '',
  `short_link` varchar(255) NOT NULL DEFAULT '',
  `created_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `updated_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `deleted_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Table structure for location
-- ----------------------------
DROP TABLE IF EXISTS `location`;
CREATE TABLE `location` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `ip` varchar(28) NOT NULL DEFAULT '',
  `continent` varchar(28) NOT NULL DEFAULT '',
  `country` varchar(128) NOT NULL DEFAULT '',
  `province` varchar(128) NOT NULL DEFAULT '',
  `city` varchar(128) NOT NULL DEFAULT '',
  `longitude` varchar(28) NOT NULL DEFAULT '',
  `area_code` varchar(28) NOT NULL DEFAULT '',
  `latitude` varchar(28) NOT NULL DEFAULT '',
  `isp` varchar(256) NOT NULL DEFAULT '',
  `zip_code` varchar(28) NOT NULL DEFAULT '',
  `elevation` varchar(28) NOT NULL DEFAULT '',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_uuid` (`ip`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Table structure for location_cn
-- ----------------------------
DROP TABLE IF EXISTS `location_cn`;
CREATE TABLE `location_cn` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `ip` varchar(28) NOT NULL DEFAULT '',
  `continent` varchar(28) NOT NULL DEFAULT '',
  `country` varchar(128) NOT NULL DEFAULT '',
  `province` varchar(128) NOT NULL DEFAULT '',
  `city` varchar(128) NOT NULL DEFAULT '',
  `longitude` varchar(28) NOT NULL DEFAULT '',
  `area_code` varchar(28) NOT NULL DEFAULT '',
  `latitude` varchar(28) NOT NULL DEFAULT '',
  `isp` varchar(256) NOT NULL DEFAULT '',
  `zip_code` varchar(28) NOT NULL DEFAULT '',
  `elevation` varchar(28) NOT NULL DEFAULT '',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_uuid` (`ip`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=42497 DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Table structure for location_en
-- ----------------------------
DROP TABLE IF EXISTS `location_en`;
CREATE TABLE `location_en` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `ip` varchar(28) NOT NULL DEFAULT '',
  `continent` varchar(28) NOT NULL DEFAULT '',
  `country` varchar(128) NOT NULL DEFAULT '',
  `province` varchar(128) NOT NULL DEFAULT '',
  `city` varchar(128) NOT NULL DEFAULT '',
  `longitude` varchar(28) NOT NULL DEFAULT '',
  `area_code` varchar(28) NOT NULL DEFAULT '',
  `latitude` varchar(28) NOT NULL DEFAULT '',
  `isp` varchar(256) NOT NULL DEFAULT '',
  `zip_code` varchar(28) NOT NULL DEFAULT '',
  `elevation` varchar(28) NOT NULL DEFAULT '',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_uuid` (`ip`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=42492 DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Table structure for login_log
-- ----------------------------
DROP TABLE IF EXISTS `login_log`;
CREATE TABLE `login_log` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `login_username` varchar(50) NOT NULL DEFAULT '',
  `ip_address` varchar(50) NOT NULL DEFAULT '',
  `login_location` varchar(255) NOT NULL DEFAULT '',
  `browser` varchar(50) NOT NULL DEFAULT '',
  `os` varchar(50) NOT NULL DEFAULT '',
  `status` tinyint(4) NOT NULL DEFAULT '0',
  `msg` varchar(255) NOT NULL DEFAULT '',
  `created_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=28957 DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Table structure for operation_log
-- ----------------------------
DROP TABLE IF EXISTS `operation_log`;
CREATE TABLE `operation_log` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `title` varchar(50) NOT NULL DEFAULT '',
  `business_type` int(2) NOT NULL DEFAULT '0',
  `method` varchar(100) NOT NULL DEFAULT '',
  `request_method` varchar(10) NOT NULL DEFAULT '',
  `operator_type` int(1) NOT NULL DEFAULT '0',
  `operator_username` varchar(50) NOT NULL DEFAULT '',
  `operator_url` varchar(500) NOT NULL DEFAULT '',
  `operator_ip` varchar(50) NOT NULL DEFAULT '',
  `operator_location` varchar(255) NOT NULL DEFAULT '',
  `operator_param` varchar(2000) NOT NULL DEFAULT '',
  `json_result` varchar(2000) NOT NULL DEFAULT '',
  `status` int(1) NOT NULL DEFAULT '0',
  `error_msg` varchar(2000) NOT NULL DEFAULT '',
  `created_at` datetime(6) NOT NULL DEFAULT '0000-00-00 00:00:00.000000',
  `updated_at` datetime(6) NOT NULL DEFAULT '0000-00-00 00:00:00.000000',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Table structure for projects
-- ----------------------------
DROP TABLE IF EXISTS `projects`;
CREATE TABLE `projects` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL DEFAULT '',
  `created_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `updated_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Table structure for retrieval_event
-- ----------------------------
DROP TABLE IF EXISTS `retrieval_event`;
CREATE TABLE `retrieval_event` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `device_id` varchar(128) NOT NULL DEFAULT '',
  `token_id` varchar(128) NOT NULL DEFAULT '',
  `client_id` varchar(128) NOT NULL DEFAULT '',
  `blocks` bigint(20) NOT NULL DEFAULT '0',
  `time` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `carfile_cid` varchar(128) NOT NULL DEFAULT '',
  `block_size` double NOT NULL DEFAULT '0',
  `status` tinyint(4) NOT NULL DEFAULT '0',
  `upstream_bandwidth` double NOT NULL DEFAULT '0',
  `start_time` int(20) NOT NULL DEFAULT '0',
  `end_time` int(20) NOT NULL DEFAULT '0',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_device_id_time` (`token_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Table structure for reward_statement
-- ----------------------------
DROP TABLE IF EXISTS `reward_statement`;
CREATE TABLE `reward_statement` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `username` varchar(64) NOT NULL DEFAULT '',
  `from_user` varchar(255) NOT NULL DEFAULT '',
  `amount` bigint(20) NOT NULL DEFAULT '0',
  `event` varchar(64) NOT NULL DEFAULT '',
  `status` int(1) DEFAULT '0',
  `device_id` varchar(64) NOT NULL DEFAULT '',
  `created_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `updated_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=116 DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Table structure for schedulers
-- ----------------------------
DROP TABLE IF EXISTS `schedulers`;
CREATE TABLE `schedulers` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `uuid` varchar(255) NOT NULL DEFAULT '',
  `area` varchar(255) NOT NULL DEFAULT '',
  `address` varchar(255) NOT NULL DEFAULT '',
  `status` int(1) NOT NULL DEFAULT '0',
  `token` varchar(255) NOT NULL DEFAULT '',
  `created_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `updated_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `deleted_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Table structure for sign_info
-- ----------------------------
DROP TABLE IF EXISTS `sign_info`;
CREATE TABLE `sign_info` (
  `miner_id` varchar(255) COLLATE utf8mb4_bin NOT NULL,
  `address` varchar(255) COLLATE utf8mb4_bin DEFAULT NULL,
  `date` bigint(20) DEFAULT NULL,
  `signed_msg` varchar(1024) COLLATE utf8mb4_bin DEFAULT NULL,
  PRIMARY KEY (`miner_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

-- ----------------------------
-- Table structure for signature
-- ----------------------------
DROP TABLE IF EXISTS `signature`;
CREATE TABLE `signature` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `username` varchar(128) NOT NULL DEFAULT '',
  `node_id` varchar(128) NOT NULL DEFAULT '',
  `area_id` varchar(128) NOT NULL DEFAULT '',
  `message` varchar(128) NOT NULL DEFAULT '',
  `hash` varchar(128) NOT NULL DEFAULT '',
  `signature` varchar(128) NOT NULL DEFAULT '',
  `created_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `updated_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  PRIMARY KEY (`id`),
  KEY `idx_hash` (`hash`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=74387 DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Table structure for storage_hour
-- ----------------------------
DROP TABLE IF EXISTS `storage_hour`;
CREATE TABLE `storage_hour` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  `updated_at` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  `deleted_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `user_id` varchar(128) NOT NULL DEFAULT '',
  `time` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  `total_storage_size` bigint(32) NOT NULL DEFAULT '0',
  `used_storage_size` bigint(32) NOT NULL DEFAULT '0',
  `total_bandwidth` bigint(32) NOT NULL DEFAULT '0',
  `peak_bandwidth` bigint(32) NOT NULL DEFAULT '0',
  `download_count` int(32) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uniq_user_id_time` (`user_id`,`time`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for storage_provider
-- ----------------------------
DROP TABLE IF EXISTS `storage_provider`;
CREATE TABLE `storage_provider` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `provider_id` varchar(255) NOT NULL DEFAULT '',
  `ip` varchar(255) NOT NULL DEFAULT '',
  `location` varchar(255) NOT NULL DEFAULT '',
  `retrievable` int(1) NOT NULL DEFAULT '0',
  `created_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `updated_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_provider` (`provider_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Table structure for storage_stats
-- ----------------------------
DROP TABLE IF EXISTS `storage_stats`;
CREATE TABLE `storage_stats` (
  `id` bigint(32) NOT NULL AUTO_INCREMENT,
  `project_id` int(20) DEFAULT '0',
  `project_name` varchar(255) NOT NULL DEFAULT '',
  `total_size` bigint(32) NOT NULL DEFAULT '0',
  `user_count` bigint(20) NOT NULL DEFAULT '0',
  `provider_count` bigint(20) NOT NULL DEFAULT '0',
  `storage_change_24h` bigint(32) NOT NULL DEFAULT '0',
  `storage_change_percentage_24h` double NOT NULL DEFAULT '0',
  `time` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `expiration` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `locations` varchar(255) NOT NULL DEFAULT '',
  `created_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `updated_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `gas` double NOT NULL DEFAULT '0',
  `pledge` double NOT NULL DEFAULT '0',
  `s_rank` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Table structure for subscription
-- ----------------------------
DROP TABLE IF EXISTS `subscription`;
CREATE TABLE `subscription` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `company` varchar(128) NOT NULL DEFAULT '',
  `name` varchar(128) NOT NULL DEFAULT '',
  `email` varchar(128) NOT NULL DEFAULT '',
  `telegram` varchar(128) NOT NULL DEFAULT '',
  `wechat` varchar(128) NOT NULL DEFAULT '',
  `location` varchar(128) NOT NULL DEFAULT '',
  `storage` varchar(128) NOT NULL DEFAULT '',
  `calculation` varchar(128) NOT NULL DEFAULT '',
  `bandwidth` varchar(128) NOT NULL DEFAULT '',
  `join_testnet` int(3) NOT NULL DEFAULT '0',
  `idle_resource_percentages` varchar(128) NOT NULL DEFAULT '',
  `subscribe` int(3) NOT NULL DEFAULT '0',
  `source` varchar(128) NOT NULL DEFAULT '',
  `created_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `updated_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=12 DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Table structure for system_info
-- ----------------------------
DROP TABLE IF EXISTS `system_info`;
CREATE TABLE `system_info` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `scheduler_uuid` varchar(128) NOT NULL DEFAULT '',
  `car_file_count` bigint(20) NOT NULL DEFAULT '0',
  `download_count` bigint(20) NOT NULL DEFAULT '0',
  `next_election_time` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_uuid` (`scheduler_uuid`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=34050 DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Table structure for user_secret
-- ----------------------------
DROP TABLE IF EXISTS `user_secret`;
CREATE TABLE `user_secret` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` varchar(255) NOT NULL DEFAULT '',
  `app_key` varchar(255) NOT NULL DEFAULT '',
  `app_secret` varchar(255) NOT NULL DEFAULT '',
  `status` int(1) DEFAULT '0',
  `created_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `updated_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `deleted_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Table structure for users
-- ----------------------------
DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `uuid` varchar(255) NOT NULL DEFAULT '',
  `avatar` varchar(255) NOT NULL DEFAULT '',
  `username` varchar(255) NOT NULL DEFAULT '',
  `pass_hash` varchar(255) NOT NULL DEFAULT '',
  `user_email` varchar(255) NOT NULL DEFAULT '',
  `wallet_address` varchar(255) NOT NULL DEFAULT '',
  `role` tinyint(4) NOT NULL DEFAULT '0',
  `allocate_storage` int(1) NOT NULL DEFAULT '0',
  `created_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `updated_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `deleted_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `project_id` int(20) NOT NULL DEFAULT '0',
  `referral_code` varchar(64) NOT NULL DEFAULT '',
  `referrer` varchar(64) NOT NULL DEFAULT '',
  `reward` decimal(14,6) DEFAULT '0.000000',
  `closed_test_reward` decimal(14,6) NOT NULL DEFAULT '0.000000',
  `referral_reward` decimal(14,6) DEFAULT '0.000000',
  `device_count` bigint(20) NOT NULL DEFAULT '0',
  `payout` decimal(14,6) DEFAULT '0.000000',
  `frozen_reward` decimal(14,6) DEFAULT '0.000000',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_username` (`username`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=2058 DEFAULT CHARSET=utf8mb4;

-- ----------------------------
-- Table structure for validation_event
-- ----------------------------
DROP TABLE IF EXISTS `validation_event`;
CREATE TABLE `validation_event` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `device_id` varchar(128) NOT NULL DEFAULT '',
  `validator_id` varchar(128) NOT NULL DEFAULT '',
  `blocks` bigint(20) NOT NULL DEFAULT '0',
  `status` tinyint(4) NOT NULL DEFAULT '0',
  `time` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
  `duration` bigint(20) NOT NULL DEFAULT '0',
  `upstream_traffic` double NOT NULL DEFAULT '0',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_device_id_time` (`device_id`,`time`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

SET FOREIGN_KEY_CHECKS = 1;
