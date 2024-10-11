package job

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/gnasnik/titan-explorer/api"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/opasynq"
	"github.com/hibiken/asynq"
)

// deleteAssetGroup 删除文件组id
func deleteAssetGroup(ctx context.Context, t *asynq.Task) error {
	var payload opasynq.AssetGroupPayload

	err := json.Unmarshal(t.Payload(), &payload)
	if err != nil {
		return err
	}

	log.Printf("get payload group_id:%d-----\n", payload.GroupID)
	pids := payload.GroupID

	for {
		if len(pids) == 0 {
			return nil
		}
		// 删除当前目录下的文件
		deleteAssetByGroup(ctx, payload.UserID, pids)
		// 获取当前目录下的次级目录，并将其赋值到当前目录
		pids, err = dao.GetUserGroupByParent(ctx, payload.UserID, pids)
		if err != nil {
			log.Printf("GetUserGroupByParent error:%v-----\n", err)
			opasynq.DefaultCli.EnqueueAssetGroupID(ctx, opasynq.AssetGroupPayload{UserID: payload.UserID, GroupID: pids})
			return nil
		}
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

// deleteAsset 删除调度器文件
func deleteAsset(ctx context.Context, t *asynq.Task) error {
	var payload opasynq.DeleteAssetPayload

	// 解析塞入的内容
	err := json.Unmarshal(t.Payload(), &payload)
	if err != nil {
		return err
	}
	// 获取调度器客户端
	scli, err := api.GetSchedulerClient(ctx, payload.AreaID)
	if err != nil {
		return fmt.Errorf("get scheduler client error:%w", err)
	}
	err = scli.RemoveAssetRecord(ctx, payload.CID)
	if err != nil {
		return fmt.Errorf("remove asset record error:%w", err)
	}

	return nil
}
