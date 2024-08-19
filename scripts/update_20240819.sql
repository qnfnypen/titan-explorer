DROP TABLE IF EXISTS `area_map`;
CREATE TABLE `area_map` (
    `id` int(10) NOT NULL AUTO_INCREMENT,
    `area_en` varchar(255) NOT NULL DEFAULT '' COMMENT '区域英文名',
    `area_cn` varchar(255) NOT NULL DEFAULT '' COMMENT '区域中文名',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT '区域名称映射';