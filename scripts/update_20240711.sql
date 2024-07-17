alter table users add column `total_storage_size` BIGINT NOT NULL DEFAULT 0;
alter table users add column `used_storage_size` BIGINT NOT NULL DEFAULT 0;
alter table users add column `total_traffic` BIGINT NOT NULL DEFAULT 0;
alter table users add column `peak_bandwidth` INT NOT NULL DEFAULT 0;
alter table users add column `download_count` INT NOT NULL DEFAULT 0;
alter table users add column `enable_vip` BOOLEAN DEFAULT false;
alter table users add column `api_keys` BLOB;