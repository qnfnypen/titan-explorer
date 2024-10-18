package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/Filecoin-Titan/titan/api"
	"github.com/Filecoin-Titan/titan/api/terrors"
	"github.com/Filecoin-Titan/titan/api/types"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/errors"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/core/oprds"
	"github.com/gnasnik/titan-explorer/core/storage"
	"github.com/rs/xid"
)

const (
	maxTempAssetDownloadCount int64 = 20
	maxTempAssetShareCount    int64 = 60
)

type (
	// UploadTempFileReq 上传临时文件
	UploadTempFileReq struct {
		AssetName string   `json:"asset_name" binding:"required"`
		AssetCID  string   `json:"asset_cid" binding:"required"`
		NodeID    string   `json:"node_id"`
		AssetSize int64    `json:"asset_size" binding:"required"`
		AreaIDs   []string `json:"area_ids"` // 最多3个
		NeedTrace bool     `json:"need_trace"`
	}
)

// UploadTmepFile 未登陆用户受限制上传文件
// @Summary 首页上传文件
// @Description 首页上传文件，如果返回的为空数组，则不调用上传接口
// @Tags temp_file
// @Param req body UploadTempFileReq true "文件上传参数"
// @Success 200 {object} JsonObject "{[]{CandidateAddr: “”, Token: “”}}"
// @Router /api/v1/storage/temp_file/upload [post]
func UploadTempFile(c *gin.Context) {
	var (
		req     UploadTempFileReq
		rsp     = make([]JsonObject, 0)
		payload = oprds.UnLoginSyncArea{}
		areaIDs []string
	)

	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}
	if len(req.AreaIDs) > 3 {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}
	allAreaIDs, maps := getAreaIDsByAreaID(c, req.AreaIDs)
	// 由于这里是匹配好的数据，所以不用做边界检查
	areaIDs = append(areaIDs, allAreaIDs[0])
	as := strings.Split(areaIDs[0], "-")
	if len(req.AreaIDs) != 0 {
		for _, v := range allAreaIDs {
			if !strings.EqualFold(as[1], v) {
				continue
			}
			areaIDs = append(areaIDs, maps[v][0])
		}
	} else {
		for k, v := range maps {
			if !strings.EqualFold(k, as[1]) {
				areaIDs = append(areaIDs, v[0])
			}
			if len(areaIDs) == 3 {
				break
			}
		}
	}
	req.AreaIDs = areaIDs

	// 最多只能是100M
	if req.AssetSize > 100*1024*1024 {
		c.JSON(http.StatusOK, respErrorCode(int(terrors.UserStorageSizeNotEnough), c))
		return
	}

	// 获取文件hash
	hash, err := storage.CIDToHash(req.AssetCID)
	if err != nil {
		log.Errorf("UploadTmepFile CIDToHash error: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	// 判断文件是否已经存在
	aids, _ := oprds.GetClient().GetUnloginAssetAreaIDs(c.Request.Context(), hash)
	// 判断文件是否已经被上传分享60次了
	taInfo, err := dao.GetTempAssetInfo(c.Request.Context(), hash)
	switch err {
	case sql.ErrNoRows:
	case nil:
		if taInfo.ShareCount >= maxTempAssetShareCount {
			c.JSON(http.StatusOK, respErrorCode(errors.TempAssetUploadErr, c))
			return
		}
		if len(aids) != 0 {
			dao.AddTempAssetShareCount(c.Request.Context(), hash)
			c.JSON(http.StatusOK, respJSON(rsp))
			return
		}
	default:
		log.Errorf("GetTempAssetInfo error: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	schCli, err := getSchedulerClient(c.Request.Context(), req.AreaIDs[0])
	if err != nil {
		log.Errorf("get scheduler client error:%v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	uid := xid.New().String()

	var traceID string
	if req.NeedTrace {
		traceID, err = dao.NewLogTrace(c.Request.Context(), uid, dao.AssetTransferTypeUpload, req.AreaIDs[0])
		if err != nil {
			log.Errorf("NewLogTrace error: %v", err)
			c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
			return
		}
	}

	careq := &types.CreateAssetReq{
		AssetCID:      req.AssetCID,
		AssetSize:     req.AssetSize,
		NodeID:        req.NodeID,
		ExpirationDay: 1,
		UserID:        uid,
		Owner:         uid,
		TraceID:       traceID,
	}
	if len(req.AreaIDs) == 1 {
		careq.ReplicaCount = 20
	} else {
		careq.ReplicaCount = 10
	}
	createAssetRsp, err := schCli.CreateAsset(c.Request.Context(), careq)
	if err != nil {
		log.Errorf("CreateAssetHandler CreateAsset error: %v", err)
		if webErr, ok := err.(*api.ErrWeb); ok {
			c.JSON(http.StatusOK, respErrorCode(webErr.Code, c))
			return
		}
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	// 塞到redis中去
	if len(req.AreaIDs) > 1 {
		oprds.GetClient().PushAreaIDs(c.Request.Context(), &oprds.AreaIDPayload{CID: req.AssetCID, Hash: hash, AreaIDs: req.AreaIDs})
	}
	// 不管是否成功，都先塞到redis中去
	for i, v := range req.AreaIDs {
		isSync := false
		if i == 0 {
			isSync = true
		}
		payload.List = append(payload.List, oprds.UnloginSyncAreaDetail{AreaID: v, IsSync: isSync})
	}
	oprds.GetClient().SetUnloginAssetInfo(c.Request.Context(), hash, &payload)

	if !createAssetRsp.AlreadyExists {
		for _, v := range createAssetRsp.List {
			rsp = append(rsp, JsonObject{"CandidateAddr": v.UploadURL, "Token": v.Token, "TraceID": traceID})
		}
	}

	dao.AddTempAssetShareCount(c.Request.Context(), hash)
	dao.AddTempAssetInfo(c.Request.Context(), hash, req.AssetSize)

	c.JSON(http.StatusOK, respJSON(rsp))
}

// ShareTempFile 分享
func ShareTempFile(c *gin.Context) {
	cid := c.Param("cid")

	// 获取文件hash
	hash, err := storage.CIDToHash(cid)
	if err != nil {
		log.Errorf("ShareTempFile CIDToHash error: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	// 判断文件是否已经被上传分享60次了
	taInfo, err := dao.GetTempAssetInfo(c.Request.Context(), hash)
	switch err {
	case sql.ErrNoRows:
		c.JSON(http.StatusOK, respErrorCode(errors.NotFound, c))
		return
	case nil:
		if taInfo.ShareCount >= maxTempAssetShareCount {
			c.JSON(http.StatusOK, respErrorCode(errors.TempAssetUploadErr, c))
			return
		}
	default:
		log.Errorf("ShareTempFile error: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	// 获取文件所有区域
	aids, err := oprds.GetClient().GetUnloginAssetAreaIDs(c.Request.Context(), hash)
	if len(aids) == 0 {
		c.JSON(http.StatusOK, respErrorCode(errors.NotFound, c))
		return
	}
	// 获取用户的访问的ip
	ip, err := GetIPFromRequest(c.Request)
	if err != nil {
		log.Errorf("get user's ip of request error:%w", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	areaID, err := GetNearestAreaID(c.Request.Context(), ip, aids)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	schedulerClient, err := getSchedulerClient(c.Request.Context(), areaID)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}
	urls, err := schedulerClient.ShareAssets(c.Request.Context(), "", []string{cid}, time.Time{})
	if err != nil {
		if webErr, ok := err.(*api.ErrWeb); ok {
			c.JSON(http.StatusOK, respErrorCode(webErr.Code, c))
			return
		}
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	if len(urls) == 0 {
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	dao.AddTempAssetShareCount(c.Request.Context(), hash)
	c.Redirect(301, urls[cid][0])
}

// DownloadTempFile 下载 ·
// @Summary 下载首页上传文件
// @Description 下载首页上传文件
// @Tags temp_file
// @Param cid path string true "文件的cid"
// @Success 200 {object} JsonObject ""
// @Router /api/v1/storage/temp_file/download/{cid} [get]
func DownloadTempFile(c *gin.Context) {
	cid := c.Param("cid")

	// 获取文件hash
	hash, err := storage.CIDToHash(cid)
	if err != nil {
		log.Errorf("ShareTempFile CIDToHash error: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	// 判断文件是否已经被下载20次了
	taInfo, err := dao.GetTempAssetInfo(c.Request.Context(), hash)
	switch err {
	case sql.ErrNoRows:
		log.Debug("get no rows")
		c.JSON(http.StatusOK, respErrorCode(errors.NotFound, c))
		return
	case nil:
		if taInfo.DownloadCount >= maxTempAssetDownloadCount {
			c.JSON(http.StatusOK, respErrorCode(errors.TempAssetDownErr, c))
			return
		}
	default:
		log.Errorf("ShareTempFile error: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	// 获取文件所有区域
	aids, err := oprds.GetClient().GetUnloginAssetAreaIDs(c.Request.Context(), hash)
	if len(aids) == 0 {
		c.JSON(http.StatusOK, respErrorCode(errors.NotFound, c))
		return
	}
	// 获取用户的访问的ip
	ip, err := GetIPFromRequest(c.Request)
	if err != nil {
		log.Errorf("get user's ip of request error:%w", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	areaID, err := GetNearestAreaID(c.Request.Context(), ip, aids)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	var traceid string
	if c.Query("need_trace") == "true" {
		traceid, err = dao.NewLogTrace(c.Request.Context(), "", dao.AssetTransferTypeDownload, areaID)
		if err != nil {
			log.Errorf("NewLogTrace error: %v", err)
			c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
			return
		}
	}

	schedulerClient, err := getSchedulerClient(c.Request.Context(), areaID)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}
	// urls, err := schedulerClient.ShareAssets(c.Request.Context(), "", []string{cid}, time.Time{})
	urls, err := schedulerClient.ShareAssetV2(c.Request.Context(), &types.ShareAssetReq{
		TraceID:    traceid,
		AssetCID:   cid,
		ExpireTime: time.Time{},
	})
	if err != nil {
		if webErr, ok := err.(*api.ErrWeb); ok {
			c.JSON(http.StatusOK, respErrorCode(webErr.Code, c))
			return
		}
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	if len(urls) == 0 {
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	dao.AddTempAssetDownloadCount(c.Request.Context(), hash)
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"asset_cid": cid,
		"size":      taInfo.Size,
		"url":       urls,
		"trace_id":  traceid,
	}))
}

// GetUploadInfo 获取上传详情
// @Summary 获取上传详情
// @Description 获取上传详情
// @Tags temp_file
// @Param cid path string true "文件的cid"
// @Success 200 {object} JsonObject "{total:0,cid:"",share_url:[]{}}"
// @Router /api/v1/storage/temp_file/info/{cid} [get]
func GetUploadInfo(c *gin.Context) {
	var (
		deviceIds []string
		complete  int64
	)
	lang := model.Language(c.GetHeader("Lang"))
	cid := c.Param("cid")

	// 获取文件hash
	hash, err := storage.CIDToHash(cid)
	if err != nil {
		log.Errorf("ShareTempFile CIDToHash error: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	// 获取调度器区域
	aids, err := oprds.GetClient().GetUnloginAssetAreaIDs(c.Request.Context(), hash)
	if err != nil || len(aids) == 0 {
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	resp := new(types.ListReplicaRsp)
	resp.ReplicaInfos = make([]*types.ReplicaInfo, 0)
	for _, v := range aids {
		schedulerClient, err := getSchedulerClient(c.Request.Context(), v)
		if err != nil {
			log.Errorf("getSchedulerClient error: %v", err)
			continue
		}
		rsp, err := schedulerClient.GetReplicas(c.Request.Context(), cid, 20, 0)
		if err != nil {
			log.Errorf("GetReplicas error: %v", err)
			continue
		}
		resp.Total += rsp.Total
		resp.ReplicaInfos = append(resp.ReplicaInfos, rsp.ReplicaInfos...)
	}
	if resp.Total > 20 {
		resp.Total = 20
	}
	if len(resp.ReplicaInfos) > 20 {
		resp.ReplicaInfos = resp.ReplicaInfos[:20]
	}
	for _, v := range resp.ReplicaInfos {
		if v.Status == types.ReplicaStatusSucceeded {
			complete++
		}
		deviceIds = append(deviceIds, v.NodeID)
	}
	deviceInfos, err := dao.GetDeviceInfoListByIds(c.Request.Context(), deviceIds)
	if err != nil {
		log.Errorf("GetAssetList err: %v", err)
	}
	mapList := dao.GenerateDeviceMapInfo(deviceInfos, lang, true)
	sort.Slice(mapList, func(i, j int) bool {
		return mapList[i]["nodeType"].(int64) > mapList[j]["nodeType"].(int64)
	})

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"total":     resp.Total,
		"complete":  complete,
		"share_url": fmt.Sprintf("%s/api/v1/storage/temp_file/share/%s", config.Cfg.BaseURL, cid),
		"maplist":   mapList,
		"list":      resp.ReplicaInfos,
	}))
}

// GetStorageCount 获取首页
func GetStorageCount(c *gin.Context) {
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"hot":  1000,
		"warm": 15000,
		"cold": 2.5,
	}))
}

// UploadTempFileCar 首页上传，后台切car
func UploadTempFileCar(c *gin.Context) {
	var (
		randomPassNonce, aid string
		urlModel             bool
	)

	areaId := getAreaIDs(c)
	userId := xid.New().String()

	schedulerClient, err := getSchedulerClient(c.Request.Context(), areaId[0])
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}

	if c.Query("encrypted") == "true" {
		passKey := fmt.Sprintf(FileUploadPassKey, userId)
		randomPassNonce = string(md5Str(userId + time.Now().String()))
		if _, err = dao.RedisCache.SetEx(c.Request.Context(), passKey, randomPassNonce, 24*time.Hour).Result(); err != nil {
			c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
			return
		}
	}
	if c.Query("urlModel") == "true" {
		urlModel = true
	}

	var traceid string
	if c.Query("need_trace") == "true" {
		traceid, err = dao.NewLogTrace(c.Request.Context(), userId, dao.AssetTransferTypeDownload, areaId[0])
		if err != nil {
			log.Errorf("NewLogTrace error: %v", err)
			c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
			return
		}
	}

	res, err := schedulerClient.GetNodeUploadInfo(c.Request.Context(), userId, randomPassNonce, urlModel)
	if err != nil {
		if webErr, ok := err.(*api.ErrWeb); ok {
			c.JSON(http.StatusOK, respErrorCode(webErr.Code, c))
			return
		}
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	as := strings.Split(areaId[0], "-")
	if len(as) < 2 {
		aid = areaId[0]
	} else {
		aid = as[1]
	}

	c.JSON(http.StatusOK, respJSON(gin.H{
		"AlreadyExists": res.AlreadyExists,
		"List":          res.List,
		"AreaID":        aid,
		"TraceID":       traceid,
		"Log":           areaId[0],
	}))
}
