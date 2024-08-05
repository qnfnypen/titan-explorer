package job

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gnasnik/titan-explorer/api"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/oprds"
	"github.com/robfig/cron/v3"
)

var (
	wg  = new(sync.WaitGroup)
	ctx = context.Background()
)

// SyncShedulersAsset 同步调度器文件
func SyncShedulersAsset() {
	c := cron.New(cron.WithLocation(time.Local))

	c.AddFunc("@every 10s", syncUserScheduler())
	c.AddFunc("@every 15s", syncUnLoginAsset())

	c.Start()
}

// syncUserScheduler 同步登陆后用户的调度器信息
func syncUserScheduler() func() {
	return func() {
		// 获取 schedulers
		payloads, err := oprds.GetClient().GetAllSchedulerInfos(ctx)
		if err != nil {
			log.Println(err)
			return
		}
		for _, v := range payloads {
			wg.Add(1)
			go func(v *oprds.Payload) {
				defer wg.Done()

				scli, err := api.GetSchedulerClient(ctx, v.AreaID)
				if err != nil {
					log.Println(fmt.Errorf("get client of scheduler error:%w", err))
					return
				}
				// 判断L1节点是否同步完成
				rs, err := scli.GetAssetRecord(ctx, v.CID)
				if err != nil {
					log.Println(fmt.Errorf("GetAssetRecord error:%w", err))
					return
				}
				if len(rs.ReplicaInfos) == 0 {
					return
				}
				unSyncAids, err := dao.GetUnSyncAreaIDs(ctx, v.UserID, v.Hash)
				if err != nil {
					log.Println(fmt.Errorf("GetUnSyncAreaIDs error:%w", err))
					return
				}
				aids, err := api.SyncShedulers(ctx, scli, "", v.CID, 0, unSyncAids)
				if err != nil {
					log.Println(fmt.Errorf("SyncShedulers error:%w", err))
					return
				}
				err = dao.UpdateUnSyncAreaIDs(ctx, v.UserID, v.Hash, aids)
				if err != nil {
					log.Println(fmt.Errorf("UpdateUnSyncAreaIDs error:%w", err))
					return
				}
				oprds.GetClient().DelSchedulerInfo(ctx, v)
			}(v)
		}
		wg.Wait()
	}
}

func syncUnLoginAsset() func() {
	return func() {
		// 获取 schedulers
		payloads, err := oprds.GetClient().GetAllAreaIDs(ctx)
		if err != nil {
			log.Println(err)
			return
		}
		for _, v := range payloads {
			wg.Add(1)
			go func(v *oprds.AreaIDPayload) {
				defer wg.Done()

				scli, err := api.GetSchedulerClient(ctx, v.AreaIDs[0])
				if err != nil {
					log.Println(fmt.Errorf("get client of scheduler error:%w", err))
					return
				}
				// 判断L1节点是否同步完成
				rs, err := scli.GetAssetRecord(ctx, v.CID)
				if err != nil {
					log.Println(fmt.Errorf("GetAssetRecord error:%w", err))
					return
				}
				if len(rs.ReplicaInfos) == 0 {
					return
				}
				aids, err := api.SyncAreaIDs(ctx, scli, "", v.CID, 0, v.AreaIDs[1:])
				if err != nil {
					log.Println(fmt.Errorf("SyncShedulers error:%w", err))
					return
				}
				aids = append(aids, v.AreaIDs[0])
				payload := oprds.UnLoginSyncArea{}
				for _, v := range aids {
					payload.List = append(payload.List, oprds.UnloginSyncAreaDetail{AreaID: v, IsSync: true})
				}
				oprds.GetClient().SetUnloginAssetInfo(ctx, v.Hash, &payload)
				oprds.GetClient().DelAreaIDs(ctx, v)
			}(v)
		}
		wg.Wait()
	}
}
