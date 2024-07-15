alter table users add column `total_storage_size` BIGINT NOT NULL DEFAULT 0;
alter table users add column `used_storage_size` BIGINT NOT NULL DEFAULT 0;
alter table users add column `total_traffic` BIGINT NOT NULL DEFAULT 0;
alter table users add column `peak_bandwidth` INT NOT NULL DEFAULT 0;
alter table users add column `download_count` INT NOT NULL DEFAULT 0;
alter table users add column `enable_vip` BOOLEAN DEFAULT false,
alter table users add column `api_keys` BLOB,

alter table assets add column `group_id` int NOT NULL DEFAULT 0;
alter table assets add column `area_id` varchar(255) NOT NULL DEFAULT 0;
alter table assets add column `share_status` TINYINT NOT NULL DEFAULT 0;
