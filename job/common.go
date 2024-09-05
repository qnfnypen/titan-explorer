package job

import (
	"context"
	"fmt"

	"github.com/Filecoin-Titan/titan/api"
	"github.com/Filecoin-Titan/titan/api/types"
	eapi "github.com/gnasnik/titan-explorer/api"
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
