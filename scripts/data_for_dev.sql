
INSERT INTO users(uuid, username, pass_hash, role, avatar) values ("90f33b18-5a3f-4243-b341-6bf856beb682", "admin", "$2a$10$8C22JaaAMhW61FsYfnGS2.Y3fgan5dytkaD2mwUQGBL.e67vie0o2", 0, "https://lf1-xgcdn-tos.pstatp.com/obj/vcloud/vadmin/start.8e0e4855ee346a46ccff8ff3e24db27b.png");

INSERT INTO schedulers(
`uuid`, `area`, `address`, `status`, `created_at`, `updated_at`, `deleted_at`
) values ('1', 'CN_GD_SHENZHEN', 'http://127.0.0.1:3456/rpc/v0', 1, 0, 0, 0);
