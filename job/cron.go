package job

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/Filecoin-Titan/titan/api/types"
	"github.com/gnasnik/titan-explorer/api"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/core/oprds"
	logging "github.com/ipfs/go-log/v2"
	"github.com/robfig/cron/v3"
)

var (
	ctx      = context.Background()
	l1States = []string{"EdgesSelect", "EdgesPulling", "Servicing", "EdgesFailed"}
	cronLog  = logging.Logger("cron")
)

// SyncShedulersAsset 同步调度器文件
func SyncShedulersAsset() {
	c := cron.New(cron.WithLocation(time.Local))

	c.AddFunc("@every 10s", syncUserScheduler())
	c.AddFunc("@every 15s", syncUnLoginAsset())
	c.AddFunc("0,10,20,30,40,50 * * * *", syncDashboard())
	c.AddFunc("@every 60s", getSyncSuccessAsset)

	c.Start()
}

// syncUserScheduler 同步登陆后用户的调度器信息
func syncUserScheduler() func() {
	return func() {
		// 获取 schedulers
		payloads, err := oprds.GetClient().GetAllSchedulerInfos(ctx)
		if err != nil {
			cronLog.Errorf("get all scheduler infos error:%v", err)
			return
		}
		wg := new(sync.WaitGroup)
		for _, v := range payloads {
			wg.Add(1)
			go func(v *oprds.Payload) {
				defer wg.Done()

				scli, err := api.GetSchedulerClient(ctx, v.AreaID)
				if err != nil {
					cronLog.Errorf("get client of scheduler error:%v", err)
					return
				}
				// 判断L1节点是否同步完成
				rs, err := scli.GetAssetRecord(ctx, v.CID)
				if err != nil {
					cronLog.Errorf("GetAssetRecord error:%v", err)
					return
				}
				if !checkSyncState(rs.State) {
					return
				}
				unSyncAids, err := dao.GetUnSyncAreaIDs(ctx, v.UserID, v.Hash)
				if err != nil {
					cronLog.Errorf("GetUnSyncAreaIDs error:%v", err)
					return
				}
				// 对于已经有的节点不再进行同步，并变更状态
				aids, err := SyncShedulers(ctx, scli, v.CID, 0, unSyncAids)
				if err != nil {
					cronLog.Errorf("SyncShedulers error:%v", err)
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
			cronLog.Errorf("get all scheduler infos error:%v", err)
			return
		}
		wg := new(sync.WaitGroup)
		for _, v := range payloads {
			wg.Add(1)
			go func(v *oprds.AreaIDPayload) {
				defer wg.Done()

				scli, err := api.GetSchedulerClient(ctx, v.AreaIDs[0])
				if err != nil {
					cronLog.Errorf("get client of scheduler error:%v", err)
					return
				}
				// 判断L1节点是否同步完成
				rs, err := scli.GetAssetRecord(ctx, v.CID)
				if err != nil {
					cronLog.Errorf("GetAssetRecord error:%v", err)
					return
				}
				if len(rs.ReplicaInfos) == 0 {
					return
				}
				aids, err := api.SyncAreaIDs(ctx, scli, "", v.CID, 0, v.AreaIDs[1:])
				if err != nil {
					cronLog.Errorf("SyncShedulers error:%v", err)
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
			wg            = new(sync.WaitGroup)
			trafficMaps   = new(sync.Map)
			bandwidthMaps = new(sync.Map)
		)

		// 获取当前时间
		now := time.Now()
		pendTime := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, now.Location())
		startTime := pendTime.Add(-10 * time.Minute)
		endTime := pendTime.Add(-1 * time.Second)

		areaIDs, err := getAllAreaIDs()
		if err != nil {
			cronLog.Errorf("get all areaids error:%v", err)
			return
		}

		for _, v := range areaIDs {
			wg.Add(1)
			go func(v string) {
				defer wg.Done()

				scli, err := api.GetSchedulerClient(ctx, v)
				if err != nil {
					cronLog.Errorf("get client of scheduler error:%v", err)
					return
				}
				infos, err := scli.GetDownloadResultsFromAssets(ctx, nil, startTime, endTime)
				if err != nil {
					cronLog.Errorf("GetDownloadResultsFromAssets error:%v", err)
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
			cronLog.Errorf("storeAssetHourStorages error:%v", err)
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
		if ts.Minute() == 0 {
			ahs.DownloadCount, _ = oprds.GetClient().GetAssetHourDownload(ctx, hash, ts)
		}
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

// getSyncSuccessAsset 更新
func getSyncSuccessAsset() {
	var (
		wg = new(sync.WaitGroup)
	)

	// 获取所有调度器区域
	areaIDs, err := getAllAreaIDs()
	if err != nil {
		log.Println(err)
		return
	}
	for _, v := range areaIDs {
		wg.Add(1)
		go func(v string) {
			defer wg.Done()

			hashs, err := getSyncSuccessHash(v)
			if err != nil {
				cronLog.Errorf("getSyncSuccessHash error:%v", err)
				return
			}
			err = dao.UpdateSyncAssetAreas(ctx, v, hashs)
			if err != nil {
				cronLog.Errorf("UpdateSyncAssetAreas error:%v", err)
			}
		}(v)
	}
	wg.Wait()
}

func getAllAreaIDs() ([]string, error) {
	var areaIDs []string

	_, maps, err := api.GetAndStoreAreaIDs()
	if err != nil {
		return nil, err
	}
	for _, v := range maps {
		areaIDs = append(areaIDs, v...)
	}

	return areaIDs, nil
}

func getSyncSuccessHash(v string) ([]string, error) {
	var (
		limit    = 100
		records  []*types.AssetRecord
		syncHash []string
	)

	// 获取文件信息
	scli, err := api.GetSchedulerClient(ctx, v)
	if err != nil {
		return nil, fmt.Errorf("get client of scheduler error:%w", err)
	}
	rsp, err := scli.GetActiveAssetRecords(ctx, 0, limit)
	if err != nil {
		return nil, fmt.Errorf("get active asset records error:%w", err)
	}
	records = append(records, rsp.List...)
	// 处理offset
	page := rsp.Total / int64(limit)
	if rsp.Total%int64(limit) > 0 {
		page++
	}
	for i := 2; i <= int(page); i++ {
		offset := (i - 1) * limit
		rsp, err := scli.GetActiveAssetRecords(ctx, int(offset), limit)
		if err != nil {
			continue
		}
		records = append(records, rsp.List...)
	}
	cronLog.Debugf("area:%v total:%v page:%v records:%v", v, rsp.Total, page, len(records))
	// 处理同步完成的状态
	for _, vv := range records {
		if checkSyncState(vv.State) {
			syncHash = append(syncHash, vv.Hash)
		} else {
			// 如果5分钟后还没有同步完成，则删除该区域的同步，重新进行同步
			if time.Now().After(vv.CreatedTime.Add(5*time.Second)) && dao.CheckAssetHashIsExist(ctx, vv.Hash) {
				if err = scli.RemoveAssetRecord(ctx, vv.CID); err == nil {
					if aid, err := dao.GetOneSyncSuccessArea(ctx, vv.Hash); err == nil {
						oprds.GetClient().PushSchedulerInfo(ctx, &oprds.Payload{CID: vv.CID, Hash: vv.Hash, AreaID: aid})
					}
				}
			}
		}
	}

	return syncHash, nil
}

func checkSyncState(state string) bool {
	for _, v := range l1States {
		if strings.EqualFold(v, state) {
			return true
		}
	}

	return false
}
