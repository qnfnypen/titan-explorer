ALTER TABLE asset_visit_count ADD COLUMN user_id varchar(255) DEFAULT '' NULL;

-- 删除现有主键
ALTER TABLE asset_visit_count DROP PRIMARY KEY;
-- 添加新的主键
ALTER TABLE asset_visit_count ADD PRIMARY KEY (`hash`, `user_id`);