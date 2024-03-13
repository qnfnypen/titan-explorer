create table sign_info(
`miner_id` varchar(255),
`address` varchar(255),
`date` bigint,
`signed_msg` varchar(1024),
primary key(miner_id)
);


create table epoch_info(
`id` bigint(20) NOT NULL AUTO_INCREMENT,
`name`  VARCHAR(128) NOT NULL DEFAULT '',
`token` VARCHAR(128) NOT NULL DEFAULT '',
`start_date` DATETIME(3) NOT NULL DEFAULT 0,
`end_date`  DATETIME(3) NOT NULL DEFAULT 0,
`created_at` DATETIME(3) NOT NULL DEFAULT 0,
`updated_at` DATETIME(3) NOT NULL DEFAULT 0,
)


ALTER  TABLE  device_info ADD COLUMN nat_type VARCHAR(128) NOT NULL DEFAULT '' AFTER retrieval_count;


 -- 0302

ALTER TABLE full_node_info ADD COLUMN  online_validator_count bigint(20) NOT NULL DEFAULT 0 AFTER validator_count;
ALTER TABLE full_node_info ADD COLUMN  online_candidate_count bigint(20) NOT NULL DEFAULT 0 AFTER candidate_count;
ALTER TABLE full_node_info ADD COLUMN  online_edge_count bigint(20) NOT NULL DEFAULT 0 AFTER edge_count;
ALTER TABLE full_node_info ADD COLUMN  memory bigint(20) NOT NULL DEFAULT 0 AFTER edge_count;
ALTER TABLE full_node_info ADD COLUMN  ip_count bigint(20) NOT NULL DEFAULT 0 AFTER memory;
ALTER TABLE full_node_info ADD COLUMN  cpu_cores bigint(20) NOT NULL DEFAULT 0 AFTER memory;

ALTER  TABLE  device_info ADD COLUMN income_incr FLOAT(32) NOT NULL DEFAULT 0 AFTER retrieval_count;



ALTER TABLE users MODIFY reward FLOAT(32) NOT NULL DEFAULT 0;
ALTER TABLE users MODIFY payout FLOAT(32) NOT NULL DEFAULT 0;
ALTER TABLE users MODIFY frozen_reward FLOAT(32) NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN referral_reward FLOAT(32) NOT NULL DEFAULT 0 AFTER reward;
ALTER TABLE users ADD COLUMN device_count bigint(20) NOT NULL DEFAULT 0 AFTER referral_reward;

ALTER  TABLE  device_info ADD COLUMN area_id VARCHAR(64) NOT NULL DEFAULT '';


ALTER TABLE device_info MODIFY COLUMN today_profit DECIMAL(14, 6) DEFAULT 0;
ALTER TABLE device_info MODIFY COLUMN yesterday_profit DECIMAL(14, 6) DEFAULT 0;
ALTER TABLE device_info MODIFY COLUMN seven_days_profit DECIMAL(14, 6) DEFAULT 0;
ALTER TABLE device_info MODIFY COLUMN month_profit DECIMAL(14, 6) DEFAULT 0;
ALTER TABLE device_info MODIFY COLUMN cumulative_profit DECIMAL(14, 6) DEFAULT 0;
ALTER TABLE device_info MODIFY COLUMN available_profit DECIMAL(14, 6) DEFAULT 0;

ALTER TABLE device_info_hour MODIFY COLUMN hour_income DECIMAL(14, 6) DEFAULT 0;
ALTER TABLE device_info_daily MODIFY COLUMN income DECIMAL(14, 6) DEFAULT 0;

ALTER TABLE users MODIFY COLUMN reward DECIMAL(14, 6) DEFAULT 0;
ALTER TABLE users MODIFY COLUMN payout DECIMAL(14, 6) DEFAULT 0;
ALTER TABLE users MODIFY COLUMN frozen_reward DECIMAL(14, 6) DEFAULT 0;
ALTER TABLE users MODIFY COLUMN referral_reward DECIMAL(14, 6) DEFAULT 0;


ALTER TABLE users ADD COLUMN closed_test_reward DECIMAL(14, 6) NOT NULL DEFAULT 0 AFTER reward;