ALTER TABLE device_info ADD COLUMN `operation` tinyint(1) NOT NULL DEFAULT '0' COMMENT '1-已退出';
ALTER TABLE device_info ADD COLUMN `deactive_time` int(10) NOT NULL DEFAULT '0' COMMENT '可以取消退出的期限';