# Titan Explorer
Titan Explorer is a RESTful API service built with Go that provides a set of endpoints for accessing and manipulating data in the application.

NOTE: This project is still in development.

## Getting Started
To install and run the API service, follow these steps:

1. Clone the repository: `git clone https://github.com/gnasnik/titan-explorer.git`
2. Copy the `config.toml-example` file to `config.toml` and modify the configuration settings as needed.
3. Install dependencies: `go mod tidy`
4. Start the server: `go run main.go`

By default, the server runs on port 8080. You can change the port by setting the `ApiListen` variable.


## Issues
Feel free to submit issues and enhancement requests.

## License
Titan Explorer is released under the terms of both the MIT License and the Apache2.0.

See [MIT](LICENSE-MIT) and [Apache2.0](LICENSE-APACHE) for more information.

## titan-storage API变更说明
+ 新增获取调度器区域列表接口: `/api/v1/storage/get_area_id`
+ 取消数据同步接口: `/api/v1/storage/get_locateStorage`
+ 需要增加 `area_id` 请求参数的接口:
  - `/api/v1/storage/create_asset`,响应由原来的单节点，变为节点列表
  - `/api/v1/storage/delete_asset`
  - `/api/v1/storage/get_asset_info`
  - `/api/v1/storage/get_asset_list`
  - `/api/v1/storage/get_all_asset_list`
  - `/api/v1/storage/share_status_set`
  - `/api/v1/storage/get_asset_count`
  - `/api/v1/storage/get_asset_group_list`
  - `/api/v1/storage/move_asset_to_group`
