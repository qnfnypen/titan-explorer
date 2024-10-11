package job

import (
	"log"

	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/opasynq"
	"github.com/hibiken/asynq"
)

// StartAsynqServer 启动 asynq 服务端
func StartAsynqServer() {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: config.Cfg.RedisAddr, Password: config.Cfg.RedisPassword},
		asynq.Config{Concurrency: 10},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(opasynq.TypeAssetGroupID, deleteAssetGroup)
	mux.HandleFunc(opasynq.TaskTypeAssetUploadedNotify, assetUploadNotify)
	mux.HandleFunc(opasynq.TypeDeleteAssetOperation, deleteAsset)

	if err := srv.Run(mux); err != nil {
		log.Fatal(err)
	}
}
