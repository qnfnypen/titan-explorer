ALTER TABLE user_asset DROP COLUMN `area_id`;
ALTER TABLE user_asset DROP COLUMN `is_sync`;
-- 删除现有主键
ALTER TABLE user_asset DROP PRIMARY KEY;
-- 添加新的主键
ALTER TABLE user_asset ADD PRIMARY KEY (`hash`, `user_id`);

DROP TABLE IF EXISTS `user_asset_area` (
`hash` VARCHAR(128) NOT NULL,
`user_id` VARCHAR(128) NOT NULL,
`area_id` varchar(128) NOT NULL,
`is_sync` BOOLEAN DEFAULT true,
PRIMARY KEY (`hash`,`user_id`,`area_id`),
) ENGINE=InnoDB COMMENT='titan存储服务的用户文件地区表';