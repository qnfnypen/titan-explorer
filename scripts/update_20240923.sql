DROP TABLE IF EXISTS `tenants`;
CREATE TABLE `tenants` (
`tenant_id` char(36) PRIMARY KEY NOT NULL,
`name` varchar(255) NOT NULL DEFAULT '',
`api_key` VARCHAR(255) NOT NULL DEFAULT '',            
`state` ENUM('active', 'inactive') NOT NULL DEFAULT 'active',
`upload_notify_url` varchar(255) NOT NULL DEFAULT '',
`created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT 'tenants info';

ALTER TABLE users ADD COLUMN tenant_id varchar(36) NOT NULL DEFAULT '';