ALTER TABLE fil_storage ADD COLUMN gas FLOAT(32) NOT NULL DEFAULT 0;
ALTER TABLE fil_storage ADD COLUMN pledge FLOAT(32) NOT NULL DEFAULT 0;


ALTER TABLE assets ADD COLUMN user_id VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE assets ADD COLUMN name VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE assets ADD COLUMN type VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE assets ADD COLUMN project_id BIGINT(20) NOT NULL DEFAULT 0;

ALTER TABLE users ADD COLUMN project_id BIGINT(20) NOT NULL DEFAULT 0;


ALTER TABLE storage_stats ADD COLUMN locations VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE storage_stats ADD COLUMN s_rank BIGINT(20) NOT NULL DEFAULT 0;

ALTER  TABLE  device_info ADD COLUMN device_status_code BIGINT(20) NOT NULL DEFAULT 0 AFTER device_status;