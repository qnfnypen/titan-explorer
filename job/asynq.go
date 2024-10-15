package job

import (
	"log"

	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/opasynq"
	"github.com/hibiken/asynq"
	"github.com/hibiken/asynqmon"
)

var MonitorHandler *asynqmon.HTTPHandler

// StartAsynqServer 启动 asynq 服务端
func StartAsynqServer() {
	go startExplorerServer()
	go startTenantServer()
}

func startExplorerServer() {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: config.Cfg.RedisAddr, Password: config.Cfg.RedisPassword},
		asynq.Config{Concurrency: 10},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(opasynq.TypeAssetGroupID, deleteAssetGroup)
	mux.HandleFunc(opasynq.TypeDeleteAssetOperation, deleteAsset)

	if err := srv.Run(mux); err != nil {
		log.Fatalf("Explorer server encountered an error: %v", err)
	}
}

func startTenantServer() {
	tenantSrv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: config.Cfg.RedisAddr, Password: config.Cfg.RedisPassword},
		asynq.Config{
			Concurrency:    10,
			RetryDelayFunc: retry,
		},
	)

	tenantMux := asynq.NewServeMux()
	tenantMux.HandleFunc(opasynq.TaskTypeAssetUploadedNotify, assetUploadNotify)
	tenantMux.HandleFunc(opasynq.TaskTypeAssetDeleteNotify, assetDeleteNotify)

	if err := tenantSrv.Run(tenantMux); err != nil {
		log.Fatalf("Tenant server encountered an error: %v", err)
	}
}
