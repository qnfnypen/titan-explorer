package api

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/Filecoin-Titan/titan/api"
	"github.com/Filecoin-Titan/titan/api/types"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/storage"
)

var (
	maxCountOfVisitAsset     int64 = 10
	maxCountOfVisitShareLink int64 = 10
)

type (
	// AssetOverview 文件概览
	AssetOverview struct {
		AssetRecord      *types.AssetRecord
		UserAssetDetail  *dao.UserAssetDetail
		VisitCount       int64
		RemainVisitCount int64
	}
	// ListAssetRecordRsp list asset records
	ListAssetRecordRsp struct {
		Total          int64            `json:"total"`
		AssetOverviews []*AssetOverview `json:"asset_infos"`
	}

	// UserAssetSummary user asset and group
	UserAssetSummary struct {
		AssetOverview *AssetOverview
		AssetGroup    *dao.AssetGroup
	}
	// ListAssetSummaryRsp list asset and group
	ListAssetSummaryRsp struct {
		Total int64               `json:"total"`
		List  []*UserAssetSummary `json:"list"`
	}
)

func getAreaIDs(c *gin.Context) []string {
	var aids []string

	areaIDs := c.QueryArray("area_id")
	for _, v := range areaIDs {
		if strings.TrimSpace(v) != "" {
			aids = append(aids, v)
		}
	}
	if len(aids) == 0 {
		areas, _ := GetAllAreasFromCache(c.Request.Context())
		if len(areas) > 0 {
			aids = append(aids, areas...)
		} else {
			aids = append(aids, GetDefaultTitanCandidateEntrypointInfo())
		}
	}

	return aids
}

func getAreaIDsNoDefault(c *gin.Context) []string {
	var aids []string

	areaIDs := c.QueryArray("area_id")
	for _, v := range areaIDs {
		if strings.TrimSpace(v) != "" {
			aids = append(aids, v)
		}
	}

	return aids
}

func getAreaID(c *gin.Context) string {
	areaID := c.Query("area_id")
	if strings.TrimSpace(areaID) == "" {
		areaID = GetDefaultTitanCandidateEntrypointInfo()
	}

	return areaID
}

func listAssets(ctx context.Context, uid string, page, size, groupID int) (*ListAssetRecordRsp, error) {
	var (
		wg = new(sync.WaitGroup)
		mu = new(sync.Mutex)
	)
	uInfo, err := dao.GetUserByUsername(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("get user's info error:%w", err)
	}
	total, infos, err := dao.ListAssets(ctx, uid, page, size, groupID)
	if err != nil {
		return nil, fmt.Errorf("get list of asset error:%w", err)
	}

	list := make([]*AssetOverview, 0)

	for _, info := range infos {
		wg.Add(1)
		go func(info *dao.UserAssetDetail) {
			defer wg.Done()

			// 获取用户文件所有调度器区域
			areaIDs, err := dao.GetUserAssetAreaIDs(ctx, info.Hash, uid)
			if err != nil {
				log.Errorf("get areaids err: %s", err.Error())
				return
			}
			// 将 hash 转换为 cid
			cid, err := storage.HashToCID(info.Hash)
			if err != nil {
				return
			}
			// 获取用户文件分发记录
			records := new(types.AssetRecord)
			records.ReplicaInfos = make([]*types.ReplicaInfo, 0)
			for i, v := range areaIDs {
				sCli, err := getSchedulerClient(ctx, v)
				if err != nil {
					log.Errorf("getSchedulerClient err: %s", err.Error())
					continue
				}
				record, err := sCli.GetAssetRecord(ctx, cid)
				if err != nil {
					log.Errorf("asset LoadAssetRecord err: %s", err.Error())
					continue
				}
				if i == 0 {
					records = record
				} else {
					records.NeedEdgeReplica += record.NeedEdgeReplica
					records.NeedCandidateReplicas += record.ReplenishReplicas
					records.ReplicaInfos = append(records.ReplicaInfos, record.ReplicaInfos...)
				}
			}
			if !uInfo.EnableVIP && info.VisitCount >= maxCountOfVisitAsset {
				info.ShareStatus = 2
			}
			info.AreaIDs = append(info.AreaIDs, areaIDs...)
			r := &AssetOverview{
				AssetRecord:      records,
				UserAssetDetail:  info,
				VisitCount:       info.VisitCount,
				RemainVisitCount: maxCountOfVisitAsset - info.VisitCount,
			}
			mu.Lock()
			list = append(list, r)
			mu.Unlock()
		}(info)
	}
	wg.Wait()

	return &ListAssetRecordRsp{Total: total, AssetOverviews: list}, nil
}

func getAssetStatus(ctx context.Context, uid, cid, areaID string) (*types.AssetStatus, error) {
	resp := new(types.AssetStatus)

	// 将cid转换为hash
	hash, err := storage.CIDToHash(cid)
	if err != nil {
		return nil, err
	}

	uInfo, err := dao.GetUserByUsername(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("get user's info error:%w", err)
	}
	aInfo, err := dao.GetUserAsset(ctx, hash, uid, areaID)
	if err != nil {
		return nil, fmt.Errorf("get asset's info error:%w", err)
	}
	resp.IsExist = true

	_ = aInfo

	// TODO
	// if aInfo.Expiration.Before(time.Now()) {
	// 	resp.IsExpiration = true
	// 	return resp, nil
	// }
	if uInfo.EnableVIP {
		return resp, nil
	}
	if aInfo.VisitCount >= maxCountOfVisitShareLink {
		resp.IsVisitOutOfLimit = true
	}

	return resp, nil
}

func listAssetSummary(ctx context.Context, uid string, parent, page, size int) (*ListAssetSummaryRsp, error) {
	resp := new(ListAssetSummaryRsp)
	offset := (page - 1) * size
	groupRsp, err := dao.ListAssetGroupForUser(ctx, uid, parent, size, offset)
	if err != nil {
		return nil, err
	}

	for _, group := range groupRsp.AssetGroups {
		i := new(UserAssetSummary)
		i.AssetGroup = group
		resp.List = append(resp.List, i)
	}
	resp.Total = groupRsp.Total

	aLimit := size - len(groupRsp.AssetGroups)
	if aLimit < 0 {
		aLimit = 0
	}
	aOffset := offset - int(groupRsp.Total)
	if aOffset < 0 {
		aOffset = 0
	}

	assetRsp, err := listAssets(ctx, uid, page, size, parent)
	if err != nil {
		return nil, err
	}
	for _, asset := range assetRsp.AssetOverviews {
		i := new(UserAssetSummary)
		i.AssetOverview = asset
		resp.List = append(resp.List, i)
	}
	resp.Total += assetRsp.Total

	return resp, nil
}

// SyncShedulers 同步调度器数据
func SyncShedulers(ctx context.Context, sCli api.Scheduler, nodeID, cid string, size int64, areaIds []string) ([]string, error) {
	zStrs := make([]string, 0)
	if len(areaIds) == 0 {
		return zStrs, nil
	}

	info, err := sCli.GenerateTokenForDownloadSource(ctx, nodeID, cid)
	if err != nil {
		log.Errorf("generate token for download source error:%w", err)
		return zStrs, nil
	}
	for _, v := range areaIds {
		scli, err := getSchedulerClient(ctx, v)
		if err != nil {
			log.Errorf("getSchedulerClient error: %v", err)
			continue
		}
		err = scli.CreateSyncAsset(ctx, &types.CreateSyncAssetReq{
			AssetCID:     cid,
			AssetSize:    size,
			DownloadInfo: info,
		})
		if err != nil {
			log.Errorf("GetUserAssetByAreaIDs error: %v", err)
			continue
		}
		zStrs = append(zStrs, v)
	}

	return zStrs, nil
}
