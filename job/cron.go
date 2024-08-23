package job

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gnasnik/titan-explorer/api"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/core/oprds"
	"github.com/robfig/cron/v3"
)

var (
	ctx = context.Background()
)

// SyncShedulersAsset 同步调度器文件
func SyncShedulersAsset() {
	c := cron.New(cron.WithLocation(time.Local))

	c.AddFunc("@every 10s", syncUserScheduler())
	c.AddFunc("@every 15s", syncUnLoginAsset())
	c.AddFunc("@hourly", syncDashboard())

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
		wg := new(sync.WaitGroup)
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
		wg := new(sync.WaitGroup)
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

func syncDashboard() func() {
	return func() {
		var (
			areaIDs       []string
			wg            = new(sync.WaitGroup)
			trafficMaps   = new(sync.Map)
			bandwidthMaps = new(sync.Map)
		)

		// 获取当前时间
		now := time.Now()
		pendTime := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, now.Location())
		startTime := pendTime.Add(-1 * time.Hour)
		endTime := pendTime.Add(-1 * time.Second)

		_, maps, err := api.GetAndStoreAreaIDs()
		if err != nil {
			log.Println(err)
			return
		}
		for _, v := range maps {
			areaIDs = append(areaIDs, v...)
		}

		for _, v := range areaIDs {
			wg.Add(1)
			go func(v string) {
				defer wg.Done()

				scli, err := api.GetSchedulerClient(ctx, v)
				if err != nil {
					log.Println(fmt.Errorf("get client of scheduler error:%w", err))
					return
				}
				infos, err := scli.GetDownloadResultsFromAssets(ctx, nil, startTime, endTime)
				if err != nil {
					log.Println(err)
					return
				}
				// 取出每个hash的最大值
				for _, v := range infos {
					storeTfOrBw(trafficMaps, v.Hash, v.TotalTraffic)
					storeTfOrBw(bandwidthMaps, v.Hash, v.PeakBandwidth)
				}
			}(v)
		}
		wg.Wait()

		err = storeAssetHourStorages(trafficMaps, bandwidthMaps, pendTime)
		if err != nil {
			log.Println(err)
		}
	}
}

func storeTfOrBw(maps *sync.Map, key string, value int64) {
	if oldValue, ok := maps.Load(key); ok {
		ov, _ := oldValue.(int64)
		if ov >= value {
			return
		}
	}

	maps.Store(key, value)
}

func storeAssetHourStorages(tmaps, bmaps *sync.Map, ts time.Time) error {
	var ahss []model.AssetStorageHour

	tmaps.Range(func(key, value any) bool {
		ahs := model.AssetStorageHour{TimeStamp: ts.Unix()}
		hash, ok := key.(string)
		if !ok {
			return true
		}
		ahs.DownloadCount, _ = oprds.GetClient().GetAssetHourDownload(ctx, hash, ts)
		ahs.Hash = hash
		tf, ok := value.(int64)
		if !ok {
			return true
		}
		ahs.TotalTraffic = tf
		if bv, ok := bmaps.LoadAndDelete(hash); ok {
			if bd, ok := bv.(int64); ok {
				ahs.PeakBandwidth = bd
			}
		}
		tmaps.Delete(hash)
		ahss = append(ahss, ahs)

		return true
	})

	return dao.AddAssetHourStorages(ctx, ahss)
}
