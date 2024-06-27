package api

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gnasnik/titan-explorer/config"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gnasnik/titan-explorer/pkg/formatter"

	"github.com/Filecoin-Titan/titan/api"
	"github.com/Filecoin-Titan/titan/api/types"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/errors"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

// GetDefaultTitanCandidateEntrypointInfo  specify candidate to upload file in testnet, only for storage api
func GetDefaultTitanCandidateEntrypointInfo() string {
	cfg := config.Cfg.SpecifyCandidate
	return cfg.AreaId
}

func GetCacheListHandler(c *gin.Context) {
	nodeId := c.Query("device_id")
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))

	if nodeId == "" {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	deviceInfo, err := dao.GetDeviceInfoByID(c.Request.Context(), nodeId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.DeviceNotExists, c))
		return
	}

	if deviceInfo == nil {
		c.JSON(http.StatusOK, respErrorCode(errors.DeviceNotExists, c))
		return
	}

	schedulerClient, err := getSchedulerClient(c.Request.Context(), deviceInfo.AreaID)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}

	// todo: get scheduler from area id
	resp, err := schedulerClient.GetReplicaEventsForNode(c.Request.Context(), nodeId, pageSize, (page-1)*pageSize)
	if err != nil {
		log.Errorf("api GetReplicaEventsForNode: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  resp.ReplicaEvents,
		"total": resp.Total,
	}))
	return
}

func GetValidationListHandler(c *gin.Context) {
	nodeId := c.Query("device_id")
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))

	deviceInfo, err := dao.GetDeviceInfoByID(c.Request.Context(), nodeId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.DeviceNotExists, c))
		return
	}

	if deviceInfo == nil {
		c.JSON(http.StatusOK, respErrorCode(errors.DeviceNotExists, c))
		return
	}

	schedulerClient, err := getSchedulerClient(c.Request.Context(), deviceInfo.AreaID)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}

	resp, err := schedulerClient.GetValidationResults(c.Request.Context(), nodeId, pageSize, (page-1)*pageSize)
	if err != nil {
		log.Errorf("api GetValidationResults: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	//var validationEvents []*model.ValidationEvent
	//for _, blockInfo := range resp.ValidationResultInfos {
	//	validationEvents = append(validationEvents, toValidationEvent(blockInfo))
	//}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  resp.ValidationResultInfos,
		"total": resp.Total,
	}))
	return
}

func GetReplicaListHandler(c *gin.Context) {
	nodeId := c.Query("device_id")
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))
	queryStatus := c.Query("status")

	if nodeId == "" {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	var status []types.ReplicaStatus
	for _, s := range strings.Split(queryStatus, ",") {
		statusVal, _ := strconv.ParseInt(s, 10, 64)
		status = append(status, types.ReplicaStatus(statusVal))
	}

	deviceInfo, err := dao.GetDeviceInfoByID(c.Request.Context(), nodeId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.DeviceNotExists, c))
		return
	}

	if deviceInfo == nil {
		c.JSON(http.StatusOK, respErrorCode(errors.DeviceNotExists, c))
		return
	}

	schedulerClient, err := getSchedulerClient(c.Request.Context(), deviceInfo.AreaID)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}

	resp, err := schedulerClient.GetReplicasForNode(c.Request.Context(), nodeId, pageSize, (page-1)*pageSize, status)
	if err != nil {
		log.Errorf("api GetReplicaEventsForNode: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  resp.NodeReplicaInfos,
		"total": resp.Total,
	}))
}

func GetProfitDetailsHandler(c *gin.Context) {
	nodeId := c.Query("device_id")
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))
	queryStatus := c.Query("ts")

	if nodeId == "" {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	var ts []int
	for _, s := range strings.Split(queryStatus, ",") {
		statusVal, _ := strconv.ParseInt(s, 10, 64)
		ts = append(ts, int(statusVal))
	}

	deviceInfo, err := dao.GetDeviceInfoByID(c.Request.Context(), nodeId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.DeviceNotExists, c))
		return
	}

	if deviceInfo == nil {
		c.JSON(http.StatusOK, respErrorCode(errors.DeviceNotExists, c))
		return
	}

	schedulerClient, err := getSchedulerClient(c.Request.Context(), deviceInfo.AreaID)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}

	resp, err := schedulerClient.GetProfitDetailsForNode(c.Request.Context(), nodeId, pageSize, (page-1)*pageSize, ts)
	if err != nil {
		log.Errorf("api GetReplicaEventsForNode: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  resp.Infos,
		"total": resp.Total,
	}))
}

//	     _______________  ____  ___   ____________   ___    ____  ____
//		/ ___/_  __/ __ \/ __ \/   | / ____/ ____/  /   |  / __ \/  _/
//		\__ \ / / / / / / /_/ / /| |/ / __/ __/    / /| | / /_/ // /
//     ___/ // / / /_/ / _, _/ ___ / /_/ / /___   / ___ |/ ____// /
//	  /____//_/  \____/_/ |_/_/  |_\____/_____/  /_/  |_/_/   /___/

func GetAllocateStorageHandler(c *gin.Context) {
	userId := c.Query("user_id")
	if userId == "" {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}
	var userInfo model.User
	userInfo.Username = userId
	_, err := dao.GetUserByUsername(c.Request.Context(), userInfo.Username)
	if err == nil {
		log.Info("GetUserByUsername user exists")
	} else {
		_ = dao.CreateUser(c.Request.Context(), &userInfo)
	}

	areaId := GetDefaultTitanCandidateEntrypointInfo()
	schedulerClient, err := getSchedulerClient(c.Request.Context(), areaId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}
	_, err = schedulerClient.AllocateStorage(c.Request.Context(), userId)
	if err != nil {
		if webErr, ok := err.(*api.ErrWeb); ok {
			c.JSON(http.StatusOK, respErrorCode(webErr.Code, c))
			return
		}
		log.Errorf("api GetValidationResults: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"msg": "success",
	}))
	return
}

func GetStorageSizeHandler(c *gin.Context) {
	userId := c.Query("user_id")
	areaId := GetDefaultTitanCandidateEntrypointInfo()
	schedulerClient, err := getSchedulerClient(c.Request.Context(), areaId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}
	storageSize, err := schedulerClient.GetUserInfo(c.Request.Context(), userId)
	if err != nil {
		if webErr, ok := err.(*api.ErrWeb); ok {
			c.JSON(http.StatusOK, respErrorCode(webErr.Code, c))
			return
		}
		log.Errorf("api GetStorageSize: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.NotFound, c))
		return
	}
	peakBandwidth := GetUserInfo(c.Request.Context(), userId)
	if peakBandwidth > storageSize.PeakBandwidth {
		storageSize.PeakBandwidth = peakBandwidth
	} else {
		var expireTime time.Duration
		expireTime = time.Hour
		// update redis data
		_ = SetUserInfo(c.Request.Context(), userId, storageSize.PeakBandwidth, expireTime)
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"PeakBandwidth": storageSize.PeakBandwidth,
		"TotalTraffic":  storageSize.TotalTraffic,
		"TotalSize":     storageSize.TotalSize,
		"UsedSize":      storageSize.UsedSize,
	}))
	return
}

func GetUserVipInfoHandler(c *gin.Context) {
	userId := c.Query("user_id")
	areaId := GetDefaultTitanCandidateEntrypointInfo()
	schedulerClient, err := getSchedulerClient(c.Request.Context(), areaId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}
	storageSize, err := schedulerClient.GetUserInfo(c.Request.Context(), userId)
	if err != nil {
		log.Errorf("api GetUserInfo: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.NotFound, c))
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"vip": storageSize.EnableVIP,
	}))
	return
}

func GetUserAccessTokenHandler(c *gin.Context) {
	UserId := c.Query("user_id")
	areaId := GetDefaultTitanCandidateEntrypointInfo()
	schedulerClient, err := getSchedulerClient(c.Request.Context(), areaId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}
	token, err := schedulerClient.GetUserAccessToken(c.Request.Context(), UserId)
	if err != nil {
		log.Errorf("api GetUserAccessToken: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.NotFound, c))
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"AccessToken": token,
	}))
}

func GetUploadInfoHandler(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	username := claims[identityKey].(string)

	areaId := GetDefaultTitanCandidateEntrypointInfo()
	schedulerClient, err := getSchedulerClient(c.Request.Context(), areaId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}

	res, err := schedulerClient.GetNodeUploadInfo(c.Request.Context(), username)
	if err != nil {
		if webErr, ok := err.(*api.ErrWeb); ok {
			c.JSON(http.StatusOK, respErrorCode(webErr.Code, c))
			return
		}
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(res))
}

func CreateAssetHandler(c *gin.Context) {
	userId := c.Query("user_id")
	areaId := GetDefaultTitanCandidateEntrypointInfo()
	schedulerClient, err := getSchedulerClient(c.Request.Context(), areaId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}

	user, err := dao.GetUserByUsername(c.Request.Context(), userId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	log.Debugf("CreateAssetHandler clientIP:%s, areaId:%s\n", c.ClientIP(), areaId)

	var createAssetReq types.CreateAssetReq
	createAssetReq.AssetName = c.Query("asset_name")
	createAssetReq.AssetCID = c.Query("asset_cid")
	createAssetReq.NodeID = c.Query("node_id")
	createAssetReq.UserID = userId
	createAssetReq.AssetType = c.Query("asset_type")
	createAssetReq.AssetSize = formatter.Str2Int64(c.Query("asset_size"))
	createAssetReq.GroupID, _ = strconv.Atoi(c.Query("group_id"))

	createAssetRsp, err := schedulerClient.CreateAsset(c.Request.Context(), &createAssetReq)
	if err != nil {
		if webErr, ok := err.(*api.ErrWeb); ok {
			c.JSON(http.StatusOK, respErrorCode(webErr.Code, c))
			return
		}
	}

	if createAssetRsp.AlreadyExists {
		c.JSON(http.StatusOK, respErrorCode(errors.FileExists, c))
		return
	}

	if err := dao.AddAssets(c.Request.Context(), []*model.Asset{
		{
			UserId:    userId,
			Name:      createAssetReq.AssetName,
			Cid:       createAssetReq.AssetCID,
			Type:      createAssetReq.AssetType,
			NodeID:    createAssetRsp.NodeID,
			TotalSize: createAssetReq.AssetSize,
			Event:     -1,
			ProjectId: user.ProjectId,
		},
	}); err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"CandidateAddr": createAssetRsp.UploadURL,
		"Token":         createAssetRsp.Token,
	}))
}

type createAssetRequest struct {
	AssetName string `json:"asset_name"`
	AssetCID  string `json:"asset_cid"`
	NodeID    string `json:"node_id"`
	AssetType string `json:"asset_type"`
	AssetSize int64  `json:"asset_size"`
	GroupId   int    `json:"group_id"`
}

func CreateAssetPostHandler(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	username := claims[identityKey].(string)

	var createAssetReq createAssetRequest
	if err := c.BindJSON(&createAssetReq); err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	areaId := GetDefaultTitanCandidateEntrypointInfo()
	schedulerClient, err := getSchedulerClient(c.Request.Context(), areaId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}

	user, err := dao.GetUserByUsername(c.Request.Context(), username)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	log.Debugf("CreateAssetHandler clientIP:%s, areaId:%s\n", c.ClientIP(), areaId)

	createAssetRsp, err := schedulerClient.CreateAsset(c.Request.Context(), &types.CreateAssetReq{
		AssetProperty: types.AssetProperty{
			AssetName: createAssetReq.AssetName,
			AssetCID:  createAssetReq.AssetCID,
			NodeID:    createAssetReq.NodeID,
			AssetType: createAssetReq.AssetType,
			AssetSize: createAssetReq.AssetSize,
			GroupID:   createAssetReq.GroupId,
		},
		UserID: username,
	})
	if err != nil {
		if webErr, ok := err.(*api.ErrWeb); ok {
			c.JSON(http.StatusOK, respErrorCode(webErr.Code, c))
			return
		}
	}

	if createAssetRsp.AlreadyExists {
		c.JSON(http.StatusOK, respErrorCode(errors.FileExists, c))
		return
	}

	if err := dao.AddAssets(c.Request.Context(), []*model.Asset{
		{
			UserId:    username,
			Name:      createAssetReq.AssetName,
			Cid:       createAssetReq.AssetCID,
			Type:      createAssetReq.AssetType,
			NodeID:    createAssetRsp.NodeID,
			TotalSize: createAssetReq.AssetSize,
			Event:     -1,
			ProjectId: user.ProjectId,
		},
	}); err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"CandidateAddr": createAssetRsp.UploadURL,
		"Token":         createAssetRsp.Token,
	}))
}

func CreateKeyHandler(c *gin.Context) {
	userId := c.Query("user_id")
	keyName := c.Query("key_name")
	permsStr := c.Query("perms")

	perms := strings.Split(permsStr, ",")
	acl := make([]types.UserAccessControl, 0, len(perms))
	for _, perm := range perms {
		acl = append(acl, types.UserAccessControl(perm))
	}

	areaId := GetDefaultTitanCandidateEntrypointInfo()
	schedulerClient, err := getSchedulerClient(c.Request.Context(), areaId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}
	keyStr, err := schedulerClient.CreateAPIKey(c.Request.Context(), userId, keyName, acl)
	if err != nil {
		log.Errorf("api CreateAPIKey: %v", err)
		if webErr, ok := err.(*api.ErrWeb); ok {
			c.JSON(http.StatusOK, respErrorCode(webErr.Code, c))
		} else {
			c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		}
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"key": keyStr,
	}))
}

func DeleteKeyHandler(c *gin.Context) {
	userId := c.Query("user_id")
	keyName := c.Query("key_name")
	areaId := GetDefaultTitanCandidateEntrypointInfo()
	schedulerClient, err := getSchedulerClient(c.Request.Context(), areaId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}
	err = schedulerClient.DeleteAPIKey(c.Request.Context(), userId, keyName)
	if err != nil {
		if webErr, ok := err.(*api.ErrWeb); ok {
			c.JSON(http.StatusOK, respErrorCode(webErr.Code, c))
			return
		}

		log.Errorf("api DeleteAPIKey: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.NotFound, c))
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"msg": "delete success",
	}))
}

func DeleteAssetHandler(c *gin.Context) {
	UserId := c.Query("user_id")
	cid := c.Query("asset_cid")
	areaId := GetDefaultTitanCandidateEntrypointInfo()
	schedulerClient, err := getSchedulerClient(c.Request.Context(), areaId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}
	err = schedulerClient.DeleteAsset(c.Request.Context(), UserId, cid)
	if err != nil {
		log.Errorf("api DeleteAsset: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"msg": "delete success",
	}))
}

func ShareAssetsHandler(c *gin.Context) {
	userId := c.Query("user_id")
	cid := c.Query("asset_cid")
	areaId := GetDefaultTitanCandidateEntrypointInfo()
	schedulerClient, err := getSchedulerClient(c.Request.Context(), areaId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}
	urls, err := schedulerClient.ShareAssets(c.Request.Context(), userId, []string{cid})
	if err != nil {
		if webErr, ok := err.(*api.ErrWeb); ok {
			c.JSON(http.StatusOK, respErrorCode(webErr.Code, c))
			return
		}
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	// todo: remove this
	var redirect bool
	if userId == "f1hl7vbopivazgion4ql25opmjnoj2ldfsvm5fuzi" {
		redirect = false
	} else {
		redirect = true
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"asset_cid": cid,
		"url":       urls[cid],
		"redirect":  redirect,
	}))
}

func ShareLinkHandler(c *gin.Context) {
	username := c.Query("username")
	cid := c.Query("cid")
	url := c.Query("url")
	if cid == "" || url == "" {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}
	var link model.Link
	link.UserName = username
	link.Cid = cid
	link.LongLink = url
	shortLink := dao.GetShortLink(c.Request.Context(), url)
	if shortLink == "" {
		link.ShortLink = "/link?" + "cid=" + cid
		shortLink = link.ShortLink
		err := dao.CreateLink(c.Request.Context(), &link)
		if err != nil {
			log.Errorf("database createLink: %v", err)
			c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
			return
		}
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"url": shortLink,
	}))
}

func GetShareLinkHandler(c *gin.Context) {
	cid := c.Query("cid")
	if cid == "" {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}
	link := dao.GetLongLink(c.Request.Context(), cid)
	if link == "" {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}
	c.Redirect(http.StatusMovedPermanently, link)

}

func UpdateShareStatusHandler(c *gin.Context) {
	userId := c.Query("user_id")
	cid := c.Query("cid")
	areaId := GetDefaultTitanCandidateEntrypointInfo()
	schedulerClient, err := getSchedulerClient(c.Request.Context(), areaId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}
	err = schedulerClient.UpdateShareStatus(c.Request.Context(), userId, cid)
	if err != nil {
		if webErr, ok := err.(*api.ErrWeb); ok {
			c.JSON(http.StatusOK, respErrorCode(webErr.Code, c))
			return
		}

		log.Errorf("api UpdateShareStatus: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"msg": "success",
	}))
}

type AccessOverview struct {
	AssetRecord      *types.AssetRecord
	UserAssetDetail  *types.UserAssetDetail
	VisitCount       int
	RemainVisitCount int
	FilcoinCount     int64
}

func GetAssetListHandler(c *gin.Context) {
	userId := c.Query("user_id")
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))
	groupId, _ := strconv.Atoi(c.Query("group_id"))
	areaId := GetDefaultTitanCandidateEntrypointInfo()
	schedulerClient, err := getSchedulerClient(c.Request.Context(), areaId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}
	createAssetRsp, err := schedulerClient.ListAssets(c.Request.Context(), userId, pageSize, (page-1)*pageSize, groupId)
	if err != nil {
		if webErr, ok := err.(*api.ErrWeb); ok {
			c.JSON(http.StatusOK, respErrorCode(webErr.Code, c))
			return
		}

		log.Errorf("api ListAssets: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	var list []*AccessOverview
	for _, asset := range createAssetRsp.AssetOverviews {
		filReplicas, err := dao.CountFilStorage(c.Request.Context(), asset.AssetRecord.CID)
		if err != nil {
			log.Errorf("count fil storage: %v", err)
			continue
		}

		list = append(list, &AccessOverview{
			AssetRecord:      asset.AssetRecord,
			UserAssetDetail:  asset.UserAssetDetail,
			VisitCount:       asset.VisitCount,
			RemainVisitCount: asset.RemainVisitCount,
			FilcoinCount:     filReplicas,
		})
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  list,
		"total": createAssetRsp.Total,
	}))
}

func GetAssetAllListHandler(c *gin.Context) {
	userId := c.Query("user_id")
	groupId, _ := strconv.Atoi(c.Query("group_id"))
	areaId := GetDefaultTitanCandidateEntrypointInfo()
	schedulerClient, err := getSchedulerClient(c.Request.Context(), areaId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}
	var total int
	page, size := 1, 100
	var listRsp []*types.AssetOverview
loop:
	createAssetRsp, err := schedulerClient.ListAssets(c.Request.Context(), userId, size, (page-1)*size, groupId)
	if err != nil {
		log.Errorf("api ListAssets: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	listRsp = append(listRsp, createAssetRsp.AssetOverviews...)
	total += len(createAssetRsp.AssetOverviews)
	page++
	if total < createAssetRsp.Total {
		goto loop
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  listRsp,
		"total": createAssetRsp.Total,
	}))
}

func GetAssetStatusHandler(c *gin.Context) {
	userId := c.Query("username")
	cid := c.Query("cid")
	areaId := GetDefaultTitanCandidateEntrypointInfo()
	schedulerClient, err := getSchedulerClient(c.Request.Context(), areaId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}
	statusRsp, err := schedulerClient.GetAssetStatus(c.Request.Context(), userId, cid)
	if err != nil {
		if webErr, ok := err.(*api.ErrWeb); ok {
			c.JSON(http.StatusOK, respErrorCode(webErr.Code, c))
			return
		}
		log.Errorf("api GetAssetStatus: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"data": statusRsp,
	}))
}

func GetAssetCountHandler(c *gin.Context) {
	userId := c.Query("user_id")
	groupId, _ := strconv.Atoi(c.Query("group_id"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize = 100
	page = 1
	areaId := GetDefaultTitanCandidateEntrypointInfo()
	schedulerClient, err := getSchedulerClient(c.Request.Context(), areaId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}

	createAssetRsp, err := schedulerClient.ListAssets(c.Request.Context(), userId, pageSize, (page-1)*pageSize, groupId)
	if err != nil {
		log.Errorf("api ListAssets: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	if createAssetRsp.Total == 0 {
		c.JSON(http.StatusOK, respJSON(JsonObject{
			"area_count":      0,
			"candidate_count": 0,
			"edge_count":      0,
		}))
		return
	}

	var deviceIds []string
	deviceExists := make(map[string]int)
	var candidateCount int64
	var edgeCount int64
	for _, data := range createAssetRsp.AssetOverviews {
		if len(data.AssetRecord.ReplicaInfos) > 0 {
			for _, rep := range data.AssetRecord.ReplicaInfos {
				if _, ok := deviceExists[rep.NodeID]; ok {
					continue
				}
				deviceExists[rep.NodeID] = 1
				deviceIds = append(deviceIds, rep.NodeID)
				switch rep.IsCandidate {
				case true:
					candidateCount += 1
				default:
					edgeCount += 1
				}
			}
		}
	}

	countArea, e := dao.GetAreaCount(c.Request.Context(), deviceIds)
	if e != nil {
		log.Errorf("GetAssetList err: %v", e)
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"area_count":      countArea,
		"candidate_count": candidateCount,
		"edge_count":      edgeCount,
	}))
}

func GetAssetDetailHandler(c *gin.Context) {
	cid := c.Query("cid")
	lang := model.Language(c.GetHeader("Lang"))
	areaId := GetDefaultTitanCandidateEntrypointInfo()
	schedulerClient, err := getSchedulerClient(c.Request.Context(), areaId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}

	resp, err := schedulerClient.GetAssetRecord(c.Request.Context(), cid)
	if err != nil {
		if webErr, ok := err.(*api.ErrWeb); ok {
			c.JSON(http.StatusOK, respErrorCode(webErr.Code, c))
			return
		}

		log.Errorf("api GetAssetRecord: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	cityMap := make(map[string]struct{})

	var deviceIds []string
	for _, replicas := range resp.ReplicaInfos {
		if replicas.Status != 3 {
			continue
		}
		deviceIds = append(deviceIds, replicas.NodeID)
	}

	deviceInfos, e := dao.GetDeviceInfoListByIds(c.Request.Context(), deviceIds)
	if err != nil {
		log.Errorf("GetAssetList err: %v", e)
	}

	for _, nodeInfo := range deviceInfos {
		if _, ok := cityMap[nodeInfo.IpCity]; ok {
			continue
		}
		cityMap[nodeInfo.IpCity] = struct{}{}
	}

	mapList := dao.GenerateDeviceMapInfo(deviceInfos, lang)

	filReplicas, err := dao.CountFilStorage(c.Request.Context(), cid)
	if err != nil {
		log.Errorf("count fil storage: %v", err)
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"cid":               resp.CID,
		"cid_name":          "",
		"ReplicaInfo_count": len(deviceIds),
		"area_count":        len(cityMap),
		"titan_count":       len(deviceIds),
		"fileCoin_count":    filReplicas,
		"list":              mapList,
		"total":             len(mapList),
	}))
}

func GetLocationHandler(c *gin.Context) {
	//userId := c.Query("user_id")
	cid := c.Query("cid")
	lang := model.Language(c.GetHeader("Lang"))
	areaId := GetDefaultTitanCandidateEntrypointInfo()
	schedulerClient, err := getSchedulerClient(c.Request.Context(), areaId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))

	limit := pageSize
	offset := (page - 1) * pageSize

	resp, err := schedulerClient.GetReplicas(c.Request.Context(), cid, limit, offset)
	if err != nil {
		if webErr, ok := err.(*api.ErrWeb); ok {
			c.JSON(http.StatusOK, respErrorCode(webErr.Code, c))
			return
		}
		log.Errorf("api GetReplicas: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	var deviceIds []string
	if len(resp.ReplicaInfos) > 0 {
		for _, rep := range resp.ReplicaInfos {
			deviceIds = append(deviceIds, rep.NodeID)
		}
	}

	type DeviceInfoRes struct {
		DeviceId   string
		IpLocation string
		Status     string
	}

	var assetInfos []*DeviceInfoRes
	if len(deviceIds) > 0 {
		assetList, err := dao.GetDeviceInfoListByIds(c.Request.Context(), deviceIds)
		if err != nil {
			log.Errorf("GetAssetList err: %v", err)
		}
		for _, nodeInfo := range assetList {
			loc, lErr := dao.GetCacheLocation(c.Request.Context(), nodeInfo.ExternalIp, lang)
			if lErr == nil && loc != nil {
				nodeInfo.Location = *loc
			}

			assetInfos = append(assetInfos, &DeviceInfoRes{
				DeviceId:   nodeInfo.DeviceID,
				IpLocation: dao.ContactIPLocation(nodeInfo.Location, lang),
				Status:     nodeInfo.DeviceStatus,
			})
		}
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"total":     resp.Total,
		"node_list": assetInfos,
	}))
}

func GetMapByCidHandler(c *gin.Context) {
	//userId := c.Query("user_id")
	cid := c.Query("cid")
	lang := model.Language(c.GetHeader("Lang"))
	areaId := GetDefaultTitanCandidateEntrypointInfo()
	schedulerClient, err := getSchedulerClient(c.Request.Context(), areaId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}

	assetRsp, err := schedulerClient.GetAssetRecord(c.Request.Context(), cid)
	if err != nil {
		log.Errorf("api GetAssetRecord: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	var deviceIds []string
	if len(assetRsp.ReplicaInfos) > 0 {
		for _, rep := range assetRsp.ReplicaInfos {
			if rep.Status == 3 {
				deviceIds = append(deviceIds, rep.NodeID)
			}
		}
	}

	deviceInfos, e := dao.GetDeviceInfoListByIds(c.Request.Context(), deviceIds)
	if err != nil {
		log.Errorf("GetAssetList err: %v", e)
	}

	mapList := dao.GenerateDeviceMapInfo(deviceInfos, lang)

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  mapList,
		"total": len(mapList),
	}))
}

func GetAssetInfoHandler(c *gin.Context) {
	//userId := c.Query("user_id")
	cid := c.Query("cid")
	areaId := GetDefaultTitanCandidateEntrypointInfo()
	schedulerClient, err := getSchedulerClient(c.Request.Context(), areaId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}

	assetRsp, err := schedulerClient.GetAssetRecord(c.Request.Context(), cid)
	if err != nil {
		if webErr, ok := err.(*api.ErrWeb); ok {
			c.JSON(http.StatusOK, respErrorCode(webErr.Code, c))
			return
		}

		log.Errorf("api GetAssetRecord: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	var deviceIds []string
	if len(assetRsp.ReplicaInfos) > 0 {
		for _, rep := range assetRsp.ReplicaInfos {
			if rep.Status == 3 {
				deviceIds = append(deviceIds, rep.NodeID)
			}
		}
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  deviceIds,
		"total": len(deviceIds),
	}))
}

func GetKeyListHandler(c *gin.Context) {
	userId := c.Query("user_id")
	areaId := GetDefaultTitanCandidateEntrypointInfo()
	schedulerClient, err := getSchedulerClient(c.Request.Context(), areaId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}
	keyResp, err := schedulerClient.GetAPIKeys(c.Request.Context(), userId)
	if err != nil {
		log.Errorf("api GetAPIKeys: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	var out []map[string]interface{}
	for k, v := range keyResp {
		item := make(map[string]interface{})
		item["name"] = k
		item["key"] = v.APIKey
		item["time"] = v.CreatedTime
		out = append(out, item)
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list": out,
	}))
}

func GetRetrievalListHandler(c *gin.Context) {
	nodeId := c.Query("device_id")
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))

	if nodeId == "" {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	deviceInfo, err := dao.GetDeviceInfoByID(c.Request.Context(), nodeId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.DeviceNotExists, c))
		return
	}

	if deviceInfo == nil {
		c.JSON(http.StatusOK, respErrorCode(errors.DeviceNotExists, c))
		return
	}

	schedulerClient, err := getSchedulerClient(c.Request.Context(), deviceInfo.IpLocation)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}
	resp, err := schedulerClient.GetRetrieveEventRecords(c.Request.Context(), nodeId, pageSize, (page-1)*pageSize)
	if err != nil {
		log.Errorf("api GetRetrieveEventRecords: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  resp.RetrieveEventInfos,
		"total": resp.Total,
	}))
}

func toValidationEvent(in types.ValidationResultInfo) *model.ValidationEvent {
	return &model.ValidationEvent{
		DeviceID:        in.NodeID,
		ValidatorID:     in.ValidatorID,
		Status:          int32(in.Status),
		Blocks:          in.BlockNumber,
		Time:            in.StartTime,
		Duration:        in.Duration,
		UpstreamTraffic: formatter.ToFixed(float64(in.Duration)*in.Bandwidth, 2),
	}
}

func GetCacheHourHandler(c *gin.Context) {
	deviceID := c.Query("device_id")
	start := c.Query("from")
	end := c.Query("to")

	if deviceID == "" {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	dataHour := dao.QueryCacheHour(deviceID, start, end)
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"series_data": dataHour,
	}))
}

func GetCacheDaysHandler(c *gin.Context) {
	deviceID := c.Query("device_id")
	start := c.Query("from")
	end := c.Query("to")

	if deviceID == "" {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	dataDaily := dao.QueryCacheDaily(deviceID, start, end)
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"series_data": dataDaily,
	}))
}

func GetAPIKeyPermsHandler(c *gin.Context) {
	userId := c.Query("user_id")
	keyName := c.Query("key_name")

	areaId := GetDefaultTitanCandidateEntrypointInfo()
	schedulerClient, err := getSchedulerClient(c.Request.Context(), areaId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}
	perms, err := schedulerClient.GetAPPKeyPermissions(c.Request.Context(), userId, keyName)
	if err != nil {
		log.Errorf("api GetAPPKeyPermissions: %v", err)
		if webErr, ok := err.(*api.ErrWeb); ok {
			c.JSON(http.StatusOK, respErrorCode(webErr.Code, c))
		} else {
			c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		}
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"perms": perms,
	}))
}

func CreateGroupHandler(c *gin.Context) {
	userId := c.Query("user_id")
	name := c.Query("name")
	parent, _ := strconv.Atoi(c.Query("parent"))
	areaId := GetDefaultTitanCandidateEntrypointInfo()
	schedulerClient, err := getSchedulerClient(c.Request.Context(), areaId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}

	group, err := schedulerClient.CreateAssetGroup(c.Request.Context(), userId, name, parent)
	if err != nil {
		log.Errorf("api CreateAssetGroup: %v", err)
		if webErr, ok := err.(*api.ErrWeb); ok {
			c.JSON(http.StatusOK, respErrorCode(webErr.Code, c))
		} else {
			c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		}
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"group": group,
	}))
}

func GetGroupsHandler(c *gin.Context) {
	userId := c.Query("user_id")
	parent, _ := strconv.Atoi(c.Query("parent"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))

	areaId := GetDefaultTitanCandidateEntrypointInfo()
	schedulerClient, err := getSchedulerClient(c.Request.Context(), areaId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}
	rsp, err := schedulerClient.ListAssetGroup(c.Request.Context(), userId, parent, pageSize, (page-1)*pageSize)
	if err != nil {
		log.Errorf("api ListAssetGroup: %v", err)
		if webErr, ok := err.(*api.ErrWeb); ok {
			c.JSON(http.StatusOK, respErrorCode(webErr.Code, c))
		} else {
			c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		}
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  rsp.AssetGroups,
		"total": rsp.Total,
	}))
}

type AssetOrGroup struct {
	AssetOverview *AccessOverview
	Group         interface{}
}

func GetAssetGroupListHandler(c *gin.Context) {
	userId := c.Query("user_id")
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))
	parentId, _ := strconv.Atoi(c.Query("parent"))
	areaId := GetDefaultTitanCandidateEntrypointInfo()
	schedulerClient, err := getSchedulerClient(c.Request.Context(), areaId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}
	assetSummary, err := schedulerClient.ListAssetSummary(c.Request.Context(), userId, parentId, pageSize, (page-1)*pageSize)
	if err != nil {
		if webErr, ok := err.(*api.ErrWeb); ok {
			c.JSON(http.StatusOK, respErrorCode(webErr.Code, c))
			return
		}
		log.Errorf("api ListAssetSummary: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	var list []*AssetOrGroup
	for _, assetGroup := range assetSummary.List {
		if assetGroup.AssetGroup != nil {
			list = append(list, &AssetOrGroup{Group: assetGroup.AssetGroup})
			continue
		}

		asset := assetGroup.AssetOverview
		filReplicas, err := dao.CountFilStorage(c.Request.Context(), asset.AssetRecord.CID)
		if err != nil {
			log.Errorf("count fil storage: %v", err)
			continue
		}

		ao := &AccessOverview{
			AssetRecord:      asset.AssetRecord,
			UserAssetDetail:  asset.UserAssetDetail,
			VisitCount:       asset.VisitCount,
			RemainVisitCount: asset.RemainVisitCount,
			FilcoinCount:     filReplicas,
		}

		list = append(list, &AssetOrGroup{AssetOverview: ao})
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  list,
		"total": assetSummary.Total,
	}))
}

func DeleteGroupHandler(c *gin.Context) {
	userId := c.Query("user_id")
	groupId, _ := strconv.Atoi(c.Query("group_id"))

	areaId := GetDefaultTitanCandidateEntrypointInfo()
	schedulerClient, err := getSchedulerClient(c.Request.Context(), areaId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}
	err = schedulerClient.DeleteAssetGroup(c.Request.Context(), userId, groupId)
	if err != nil {
		log.Errorf("api DeleteAssetGroup: %v", err)
		if webErr, ok := err.(*api.ErrWeb); ok {
			c.JSON(http.StatusOK, respErrorCode(webErr.Code, c))
		} else {
			c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		}
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"msg": "success",
	}))
}

func RenameGroupHandler(c *gin.Context) {
	userId := c.Query("user_id")
	newName := c.Query("new_name")
	groupId, _ := strconv.Atoi(c.Query("group_id"))

	areaId := GetDefaultTitanCandidateEntrypointInfo()
	schedulerClient, err := getSchedulerClient(c.Request.Context(), areaId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}
	err = schedulerClient.RenameAssetGroup(c.Request.Context(), userId, newName, groupId)
	if err != nil {
		log.Errorf("api RenameAssetGroup: %v", err)
		if webErr, ok := err.(*api.ErrWeb); ok {
			c.JSON(http.StatusOK, respErrorCode(webErr.Code, c))
		} else {
			c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		}
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"msg": "success",
	}))
}

func MoveGroupToGroupHandler(c *gin.Context) {
	userId := c.Query("user_id")
	groupId, _ := strconv.Atoi(c.Query("group_id"))
	targetGroupId, _ := strconv.Atoi(c.Query("target_group_id"))

	areaId := GetDefaultTitanCandidateEntrypointInfo()
	schedulerClient, err := getSchedulerClient(c.Request.Context(), areaId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}
	err = schedulerClient.MoveAssetGroup(c.Request.Context(), userId, groupId, targetGroupId)
	if err != nil {
		log.Errorf("api MoveAssetGroup: %v", err)
		if webErr, ok := err.(*api.ErrWeb); ok {
			c.JSON(http.StatusOK, respErrorCode(webErr.Code, c))
		} else {
			c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		}
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"msg": "success",
	}))
}

func MoveAssetToGroupHandler(c *gin.Context) {
	userId := c.Query("user_id")
	assetCid := c.Query("asset_cid")
	groupId, _ := strconv.Atoi(c.Query("group_id"))

	areaId := GetDefaultTitanCandidateEntrypointInfo()
	schedulerClient, err := getSchedulerClient(c.Request.Context(), areaId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}
	err = schedulerClient.MoveAssetToGroup(c.Request.Context(), userId, assetCid, groupId)
	if err != nil {
		log.Errorf("api MoveAssetToGroup: %v", err)
		if webErr, ok := err.(*api.ErrWeb); ok {
			c.JSON(http.StatusOK, respErrorCode(webErr.Code, c))
		} else {
			c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		}
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"msg": "success",
	}))
}
