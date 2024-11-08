DROP TABLE IF EXISTS `sync_ipfs_record`;
CREATE TABLE `sync_ipfs_record` (
    `id` int(10) NOT NULL AUTO_INCREMENT,
    `username` varchar(255) NOT NULL DEFAULT '' COMMENT '用户名',
    `name` varchar(255) NOT NULL DEFAULT '' COMMENT '文件名',
    `cid` varchar(255) NOT NULL DEFAULT '' COMMENT '文件cid',
    `group_id` int(10) NOT NULL DEFAULT '0' COMMENT '文件组id',
    `size` int(10) NOT NULL DEFAULT '0' COMMENT '文件大小',
    `area_id` varchar(255) NOT NULL DEFAULT '' COMMENT '首个节点区域',
    `timestamp` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '同步开始的时间戳',
    `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '状态 0-未成功 1-成功',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uq_uc` (`username`,`cid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT 'ipfs文件同步列表';

ALTER TABLE asset_transfer_log ADD COLUMN `area` varchar(64) NOT NULL DEFAULT '';
