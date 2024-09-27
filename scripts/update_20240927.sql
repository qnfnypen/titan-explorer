ALTER TABLE device_info_daily ADD COLUMN `penalty_profit` double NOT NULL DEFAULT '0';
ALTER TABLE device_info ADD COLUMN `penalty_profit` double NOT NULL DEFAULT '0';