alter table device_info_daily add column `external_ip` VARCHAR(28) NOT NULL DEFAULT '';

DROP TABLE IF EXISTS `temp_asset`;
CREATE TABLE temp_asset (
`hash` VARCHAR(128) NOT NULL,
`download_count` INT DEFAULT 0,
`share_count` INT DEFAULT 0,
PRIMARY KEY (`hash`)
) ENGINE=InnoDB COMMENT='titan存储服务临时上传的文件';