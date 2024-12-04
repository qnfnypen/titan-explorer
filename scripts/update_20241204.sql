ALTER TABLE `asset_transfer_log` ADD COLUMN `available_bandwidth`  bigint(20)  NOT NULL DEFAULT '0' COMMENT '总可用带宽';

ALTER TABLE `asset_transfer_log` ADD COLUMN `first_byte_time`  int(10)  NOT NULL DEFAULT '0' COMMENT '首字节到达时间';