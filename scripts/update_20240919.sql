DROP TABLE IF EXISTS `asset_transfer_log`;
CREATE TABLE `asset_transfer_log` (
`trace_id` char(36) PRIMARY KEY NOT NULL,
`user_id` varchar(255) NOT NULL DEFAULT '',
`cid` VARCHAR(255) NOT NULL DEFAULT '',
`hash` VARCHAR(128) NOT NULL DEFAULT '',
`node_id` varchar(128) NOT NULL DEFAULT '',
`rate` int(10) NOT NULL DEFAULT 0,
`cost_ms` int(10) NOT NULL DEFAULT 0,
`total_size` int(10) NOT NULL DEFAULT 0,
`state` tinyint(1) NOT NULL DEFAULT 0 COMMENT '0:created 1:success 2:failed',
`transfer_type` enum('upload', 'download') NOT NULL DEFAULT 'download', 
`log` varchar(512) NOT NULL DEFAULT '',
`created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
KEY `idx_user_id_hash` (`user_id`, `hash`),
KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT 'file transfer metrics';
