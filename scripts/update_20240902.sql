DROP TABLE IF EXISTS `user_l1_reward`;
CREATE TABLE `user_l1_reward` (
`user_id` varchar(255) NOT NULL DEFAULT '0',
`reward` int(10) NOT NULL DEFAULT '0',
`updated_at` datetime(3) NOT NULL DEFAULT '0000-00-00 00:00:00.000',
PRIMARY KEY (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT 'l1 reward';

ALTER TABLE user_asset ADD COLUMN cid varchar(255) DEFAULT '' NOT NULL;
