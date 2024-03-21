

ALTER  TABLE  device_info ADD COLUMN titan_disk_space FLOAT(32) NOT NULL DEFAULT 0 AFTER disk_space;
ALTER  TABLE  device_info ADD COLUMN titan_disk_usage FLOAT(32) NOT NULL DEFAULT 0 AFTER titan_disk_space;

ALTER TABLE full_node_info ADD COLUMN  titan_disk_space FLOAT(32) NOT NULL DEFAULT 0 AFTER storage_used;
ALTER TABLE full_node_info ADD COLUMN  titan_disk_usage FLOAT(32) NOT NULL DEFAULT 0 AFTER titan_disk_space;