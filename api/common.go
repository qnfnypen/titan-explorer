package api

import (
	"context"
	"fmt"
	"time"

	"github.com/Filecoin-Titan/titan/api"
	"github.com/Filecoin-Titan/titan/api/types"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/storage"
)

var (
	maxCountOfVisitAsset     int64 = 20
	maxCountOfVisitShareLink int64 = 20
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

func getAreaID(c *gin.Context) string {
	areaID := c.Query("area_id")
	if areaID == "" {
		areaID = GetDefaultTitanCandidateEntrypointInfo()
	}

	return areaID
}

func listAssets(ctx context.Context, sCli api.Scheduler, uid string, page, size, groupID int) (*ListAssetRecordRsp, error) {
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
		// 将 hash 转换为 cid
		cid, err := storage.HashToCID(info.Hash)
		if err != nil {
			continue
		}
		record, err := sCli.GetAssetRecord(ctx, cid)
		if err != nil {
			log.Errorf("asset LoadAssetRecord err: %s", err.Error())
			continue
		}

		if !uInfo.EnableVIP && info.VisitCount >= maxCountOfVisitAsset {
			info.ShareStatus = 2
		} else {
			info.ShareStatus = 1
		}

		r := &AssetOverview{
			AssetRecord:      record,
			UserAssetDetail:  info,
			VisitCount:       info.VisitCount,
			RemainVisitCount: maxCountOfVisitAsset - info.VisitCount,
		}

		list = append(list, r)
	}

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

	if aInfo.Expiration.Before(time.Now()) {
		resp.IsExpiration = true
		return resp, nil
	}
	if uInfo.EnableVIP {
		return resp, nil
	}
	if aInfo.VisitCount >= maxCountOfVisitShareLink {
		resp.IsVisitOutOfLimit = true
	}

	return resp, nil
}

func listAssetSummary(ctx context.Context, sCli api.Scheduler, uid string, parent, page, size int) (*ListAssetSummaryRsp, error) {
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

	assetRsp, err := listAssets(ctx, sCli, uid, page, size, parent)
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
