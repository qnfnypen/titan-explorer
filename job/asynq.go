package job

import (
	"context"
	"encoding/json"
	"log"

	"github.com/gnasnik/titan-explorer/api"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/opasynq"
	"github.com/hibiken/asynq"
)

// StartAsynqServer 启动 asynq 服务端
func StartAsynqServer() {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: config.Cfg.RedisAddr, Password: config.Cfg.RedisPassword},
		// asynq.RedisClientOpt{Addr: "127.0.0.1:6379"},
		asynq.Config{Concurrency: 10},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(opasynq.TypeAssetGroupID, deleteAssetGroup)

	if err := srv.Run(mux); err != nil {
		log.Fatal(err)
	}
}

// deleteAssetGroup 删除文件组id
func deleteAssetGroup(ctx context.Context, t *asynq.Task) error {
	var payload opasynq.AssetGroupPayload

	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	log.Printf("get payload group_id:%d-----\n", payload.GroupID)
	pids := payload.GroupID

	for {
		if len(pids) == 0 {
			return nil
		}
		gids, err := dao.GetUserGroupByParent(ctx, payload.UserID, pids)
		if err != nil {
			log.Printf("GetUserGroupByParent error:%v-----\n", err)
			opasynq.DefaultCli.EnqueueAssetGroupID(ctx, opasynq.AssetGroupPayload{UserID: payload.UserID, GroupID: pids})
			return nil
		}
		deleteAssetByGroup(ctx, payload.UserID, gids)
		pids = gids
	}
}

// deleteAssetByGroup 通过文件组id删除文件组内的文件
func deleteAssetByGroup(ctx context.Context, uid string, gids []int64) error {
	// 删除指定文件组下的文件内容
	err := dao.DeleteUserGroupAsset(ctx, uid, gids)
	if err != nil {
		log.Printf("DeleteUserGroupAsset error:%v-----\n", err)
		opasynq.DefaultCli.EnqueueAssetGroupID(ctx, opasynq.AssetGroupPayload{UserID: uid, GroupID: gids})
		return nil
	}

	// 获取唯一存在的文件的区域映射表
	maps, err := dao.GetOnlyAssetsByUIDAndGroupID(ctx, uid, gids)
	if err != nil {
		log.Printf("GetOnlyAssetsByUIDAndGroupID error:%v-----\n", err)
		return nil
	}
	if len(maps) == 0 {
		return nil
	}

	// 调用调度器删除文件
	for k, v := range maps {
		scli, err := api.GetSchedulerClient(context.Background(), k)
		if err != nil {
			continue
		}
		scli.RemoveAssetRecords(context.Background(), v)
	}
	return nil
}
