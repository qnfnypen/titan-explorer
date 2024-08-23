DROP TABLE IF EXISTS `area_map`;
CREATE TABLE `area_map` (
    `id` int(10) NOT NULL AUTO_INCREMENT,
    `area_en` varchar(255) NOT NULL DEFAULT '' COMMENT '区域英文名',
    `area_cn` varchar(255) NOT NULL DEFAULT '' COMMENT '区域中文名',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT '区域名称映射';

DROP TABLE IF EXISTS `asset_storage_hour`;
CREATE TABLE `asset_storage_hour` (
    `id` int(10) NOT NULL AUTO_INCREMENT,
    `hash` varchar(255) NOT NULL DEFAULT '0' COMMENT '文件hash',
    `total_traffic` int(10) NOT NULL DEFAULT '0' COMMENT '流量带宽',
    `peak_bandwidth` int(10) NOT NULL DEFAULT '0' COMMENT '带宽峰值',
    `download_count` int(10) NOT NULL DEFAULT '0' COMMENT '访问量',
    `timestamp` int(10) NOT NULL DEFAULT '0' COMMENT '截至小时时间戳',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT '文件小时周期存储数据';