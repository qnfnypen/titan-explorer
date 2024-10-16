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

// operateSyncIPFSRecord 增加用户ipfs文件信息
func operateSyncIPFSRecord(ctx context.Context, t *asynq.Task) error {
	var payload opasynq.IPFSRecordPayload

	// 解析塞入的内容
	err := json.Unmarshal(t.Payload(), &payload)
	if err != nil {
		return err
	}

	// 获取其他的节点区域
	aids, err := api.GetOtherAreaIDs(payload.AreaID)
	if err != nil {
		log.Println(fmt.Errorf("SyncShedulers error:%v", err))
		return fmt.Errorf("get other areaids error:%w", err)
	}
	// 插入数据并修改相关信息
	err = dao.AddAssetAndUpdateSize(ctx, &payload.Info, aids, payload.AreaID)
	if err != nil {
		log.Println(fmt.Errorf("AddAssetAndUpdateSize error:%w", err))
		return fmt.Errorf("AddAssetAndUpdateSize error:%w", err)
	}

	scli, err := api.GetSchedulerClient(ctx, payload.AreaID)
	if err != nil {
		log.Println(fmt.Errorf("get client of scheduler error:%v", err))
		return fmt.Errorf("get client of scheduler error:%v", err)
	}
	unSyncAids, err := dao.GetUnSyncAreaIDs(ctx, payload.Info.UserID, payload.Info.Hash)
	if err != nil {
		log.Println(fmt.Errorf("SyncShedulers error:%v", err))
		return fmt.Errorf("GetUnSyncAreaIDs error:%v", err)
	}
	// 对于已经有的节点不再进行同步，并变更状态
	aids, err = SyncShedulers(ctx, scli, payload.Info.Cid, 0, payload.Info.UserID, unSyncAids)
	if err != nil {
		log.Println(fmt.Errorf("SyncShedulers error:%v", err))
		return fmt.Errorf("SyncShedulers error:%v", err)
	}
	err = dao.UpdateUnSyncAreaIDs(ctx, payload.Info.UserID, payload.Info.Hash, aids)
	if err != nil {
		log.Println(fmt.Errorf("UpdateUnSyncAreaIDs error:%w", err))
		return fmt.Errorf("UpdateUnSyncAreaIDs error:%w", err)
	}

	return nil
}
