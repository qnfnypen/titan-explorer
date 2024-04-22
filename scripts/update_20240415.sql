ALTER  TABLE  device_info ADD COLUMN cpu_info VARCHAR(128) NOT NULL DEFAULT '' AFTER cpu_cores;
ALTER  TABLE  device_info ADD COLUMN last_seen DATETIME(3) NOT NULL DEFAULT 0;

ALTER  TABLE  device_info ADD COLUMN is_mobile INT(1) NOT NULL DEFAULT 0 AFTER area_id;


ALTER TABLE users ADD COLUMN huygens_reward DECIMAL(14, 6) NOT NULL DEFAULT 0 AFTER closed_test_reward;
ALTER TABLE users ADD COLUMN huygens_referral_reward DECIMAL(14, 6) NOT NULL DEFAULT 0 AFTER huygens_reward;

ALTER TABLE users ADD COLUMN herschel_reward DECIMAL(14, 6) NOT NULL DEFAULT 0 AFTER huygens_referral_reward;
ALTER TABLE users ADD COLUMN herschel_referral_reward DECIMAL(14, 6) NOT NULL DEFAULT 0 AFTER herschel_reward;

ALTER TABLE device_info_daily ADD COLUMN is_mobile INT(1) NOT NULL DEFAULT 0;