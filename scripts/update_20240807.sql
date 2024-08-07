ALTER TABLE ads ADD COLUMN `is_text` BOOLEAN DEFAULT false; 

Update ads set is_text = true where ads_type = 2;