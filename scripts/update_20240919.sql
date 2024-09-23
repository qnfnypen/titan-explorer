DROP TABLE IF EXISTS `asset_transfer_logs`;
CREATE TABLE `asset_transfer_logs` (
`trace_id` char(36) PRIMARY KEY NOT NULL,
`user_id` varchar(255) NOT NULL DEFAULT '',
`cid` VARCHAR(255) NOT NULL DEFAULT '',
`hash` VARCHAR(128) NOT NULL DEFAULT '',
`rate` int(10) NOT NULL DEFAULT 0,
`cost_ms` int(10) NOT NULL DEFAULT 0,
`total_size` int(10) NOT NULL DEFAULT 0,
`succeed` tinyint(1) NOT NULL DEFAULT 1,
`transfer_type` enum('upload', 'download') NOT NULL DEFAULT 'download', 
`created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
KEY `idx_user_id_hash` (`user_id`, `hash`),
KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT 'file transfer metrics';
