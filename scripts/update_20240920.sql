ALTER TABLE user_asset_group ADD COLUMN share_status tinyint DEFAULT '0' NOT NULL;
ALTER TABLE user_asset_group ADD COLUMN visit_count int NOT NULL DEFAULT '0';

-- 增加用户文件映射表
DROP TABLE IF EXISTS `user_asset_map`;
CREATE TABLE `user_asset_map` (
  `id` BIGINT(20) NOT NULL AUTO_INCREMENT,
  `user_id` varchar(255) NOT NULL DEFAULT '' COMMENT '用户id',
  `asset_hash` varchar(255) NOT NULL DEFAULT '' COMMENT '文件hash',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_uh` (`user_id`,`asset_hash`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT '用户文件映射表';