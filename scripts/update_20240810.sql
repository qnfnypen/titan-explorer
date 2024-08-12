DROP TABLE IF EXISTS `edge_config`;
CREATE TABLE edge_config (
`node_id` VARCHAR(128) NOT NULL,
`config` VARCHAR(2048) NOT NULL,
`created_at` DATETIME(3) NOT NULL DEFAULT 0,
`updated_at` DATETIME(3) NOT NULL DEFAULT 0,
PRIMARY KEY (`node_id`)
) ENGINE=InnoDB COMMENT='Edge Config';  

alter table bugs add column `benefit_log` MEDIUMTEXT NOT NULL COMMENT 'edge benefit logs';