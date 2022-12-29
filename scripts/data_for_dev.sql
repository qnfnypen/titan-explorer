
INSERT INTO users(uuid, username, pass_hash, role) values ("90f33b18-5a3f-4243-b341-6bf856beb682", "admin", "$2a$10$8C22JaaAMhW61FsYfnGS2.Y3fgan5dytkaD2mwUQGBL.e67vie0o2", 0);

INSERT INTO schedulers(
`name`, `group`, `address`, `status`, `created_at`, `updated_at`, `deleted_at`
) values ('default', 'default', 'http://127.0.0.1:3456/rpc/v0', 1, 0, 0, 0);
