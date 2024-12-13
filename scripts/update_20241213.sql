CREATE TABLE IF NOT EXISTS `asset_transfer_detail` (
    `trace_id` char(36) PRIMARY KEY NOT NULL,
    `node_id` varchar(128) NOT NULL DEFAULT '',
    `state` tinyint(1) NOT NULL DEFAULT 0 COMMENT '0:created 1:success 2:failed',
    `transfer_type` enum('upload', 'download') NOT NULL DEFAULT 'download', 
    `peek` int(10) NOT NULL DEFAULT 0,
    `elasped_time` int(10) NOT NULL DEFAULT 0,
    `size` int(10) NOT NULL DEFAULT 0,
    `errors` varchar(2048) NOT NULL DEFAULT '' COMMENT 'failed logs', 
    `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
    KEY `idx_trace_id` (`trace_id`) USING BTREE,
    KEY `idx_node_id` (`node_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT '传输详情表';