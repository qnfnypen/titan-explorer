package api

import (
	"context"

	"github.com/gnasnik/titan-explorer/core/dao"
)

func getUnSyncAreas(ctx context.Context, uid, cid, hash string, allAreaIds, unSyncAreaIds []string) []string {
	var (
		notSyncAreaIds    []string
		notSyncAreaIDMaps = make(map[string]int)
	)

	// 获取文件的状态
	assets, err := dao.GetAssetsByCID(ctx, cid, allAreaIds)
	if err != nil {
		return unSyncAreaIds
	}

	for _, v := range assets {
		if v.State == "Remove" {
			notSyncAreaIds = append(notSyncAreaIds, v.AreaId)
			notSyncAreaIDMaps[v.AreaId] = 1
		}
	}

	for _, v := range unSyncAreaIds {
		if _, ok := notSyncAreaIDMaps[v]; !ok {
			notSyncAreaIds = append(notSyncAreaIds, v)
		}
	}

	// 变更状态为未同步
	dao.UpdateUnSyncAreaIDs(ctx, uid, hash, notSyncAreaIds, false)

	return notSyncAreaIds
}
