package job

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Filecoin-Titan/titan/api"
	"github.com/Filecoin-Titan/titan/api/types"
	eapi "github.com/gnasnik/titan-explorer/api"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/core/oprds"
)

// SyncShedulers 同步调度器数据
func SyncShedulers(ctx context.Context, sCli api.Scheduler, cid string, size int64, areaIds []string) ([]string, error) {
	zStrs := make([]string, 0)
	if len(areaIds) == 0 {
		return zStrs, nil
	}

	info, err := sCli.GenerateTokenForDownloadSources(ctx, cid)
	if err != nil {
		return zStrs, fmt.Errorf("generate token for download source error:%w", err)
	}
	for _, v := range areaIds {
		scli, err := eapi.GetSchedulerClient(ctx, v)
		if err != nil {
			cronLog.Errorf("getSchedulerClient error: %v", err)
			continue
		}
		ar, err := scli.GetAssetRecord(ctx, cid)
		if err == nil && checkSyncState(ar.State) {
			zStrs = append(zStrs, v)
			continue
		}
		err = scli.CreateSyncAsset(ctx, &types.CreateSyncAssetReq{
			AssetCID:      cid,
			AssetSize:     size,
			DownloadInfos: info,
		})
		if err != nil {
			cronLog.Errorf("GetUserAssetByAreaIDs error: %v", err)
		}
	}

	return zStrs, nil
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

func getAllAreaIDs() ([]string, error) {
	var areaIDs []string

	_, maps, err := eapi.GetAndStoreAreaIDs()
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
	scli, err := eapi.GetSchedulerClient(ctx, v)
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
