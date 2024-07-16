package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gnasnik/titan-explorer/config"

	"github.com/gnasnik/titan-explorer/pkg/formatter"

	"github.com/Filecoin-Titan/titan/api"
	"github.com/Filecoin-Titan/titan/api/terrors"
	"github.com/Filecoin-Titan/titan/api/types"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/errors"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/core/storage"
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

	// areaId := GetDefaultTitanCandidateEntrypointInfo()
	// schedulerClient, err := getSchedulerClient(c.Request.Context(), areaId)
	// if err != nil {
	// 	c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
	// 	return
	// }
	// _, err = schedulerClient.AllocateStorage(c.Request.Context(), userId)
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
	claims := jwt.ExtractClaims(c)
	username := claims[identityKey].(string)
	user, err := dao.GetUserByUsername(c.Request.Context(), username)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	peakBandwidth := GetUserInfo(c.Request.Context(), username)
	if peakBandwidth > user.PeakBandwidth {
		user.PeakBandwidth = peakBandwidth
	} else {
		var expireTime time.Duration
		expireTime = time.Hour
		// update redis data
		_ = SetUserInfo(c.Request.Context(), username, user.PeakBandwidth, expireTime)
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"PeakBandwidth": user.PeakBandwidth,
		"TotalTraffic":  user.TotalTraffic,
		"TotalSize":     user.TotalStorageSize,
		"UsedSize":      user.UsedStorageSize,
	}))
	return
}

func GetUserVipInfoHandler(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	username := claims[identityKey].(string)
	user, err := dao.GetUserByUsername(c.Request.Context(), username)
	if err != nil {
		log.Errorf("api GetUserInfo: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.NotFound, c))
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"vip": user.EnableVIP,
	}))
	return
}

func GetUserAccessTokenHandler(c *gin.Context) {
	// UserId := c.Query("user_id")
	// claims := jwt.ExtractClaims(c)
	// UserId := claims[identityKey].(string)
	// areaId := getAreaID(c)
	// schedulerClient, err := getSchedulerClient(c.Request.Context(), areaId)
	// if err != nil {
	// 	c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
	// 	return
	// }
	// token, err := schedulerClient.GetUserAccessToken(c.Request.Context(), UserId)
	// if err != nil {
	// 	log.Errorf("api GetUserAccessToken: %v", err)
	// 	c.JSON(http.StatusOK, respErrorCode(errors.NotFound, c))
	// 	return
	// }
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"AccessToken": "token",
	}))
}

func GetUploadInfoHandler(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	username := claims[identityKey].(string)

	areaId := getAreaID(c)
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
	// userId := c.Query("user_id")
	claims := jwt.ExtractClaims(c)
	userId := claims[identityKey].(string)
	areaId := getAreaID(c)

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

	schedulerClient, err := getSchedulerClient(c.Request.Context(), areaId)
	// createAssetRsp, err := schedulerClient.CreateAsset(c.Request.Context(), &createAssetReq)
	// if err != nil {
	// 	if webErr, ok := err.(*api.ErrWeb); ok {
	// 		c.JSON(http.StatusOK, respErrorCode(webErr.Code, c))
	// 		return
	// 	}
	// }
	ainfo, _ := dao.GetAssetByCIDAndUser(c.Request.Context(), createAssetReq.AssetCID, userId)
	if ainfo != nil && ainfo.ID > 0 {
		c.JSON(http.StatusOK, respErrorCode(errors.FileExists, c))
		return
	}
	// 判断用户存储空间是否够用
	if user.TotalStorageSize-user.UsedStorageSize < createAssetReq.AssetSize {
		c.JSON(http.StatusOK, respErrorCode(int(terrors.UserStorageSizeNotEnough), c))
		return
	}
	// 获取文件hash
	hash, err := storage.CIDToHash(createAssetReq.AssetCID)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	createAssetRsp, err := schedulerClient.CreateAsset(c.Request.Context(), &types.CreateAssetReq{AssetProperty: types.AssetProperty{AssetCID: createAssetReq.AssetCID}})
	if err != nil {
		if webErr, ok := err.(*api.ErrWeb); ok {
			c.JSON(http.StatusOK, respErrorCode(webErr.Code, c))
			return
		}
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	if err := dao.AddAssetAndUpdateSize(c.Request.Context(), &model.Asset{
		UserId:    userId,
		Name:      createAssetReq.AssetName,
		Cid:       createAssetReq.AssetCID,
		Type:      createAssetReq.AssetType,
		NodeID:    createAssetReq.NodeID,
		TotalSize: createAssetReq.AssetSize,
		GroupID:   int64(createAssetReq.GroupID),
		AreaID:    areaId,
		Hash:      hash,
		Event:     -1,
		ProjectId: user.ProjectId,
	}); err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	rsp := make([]JsonObject, len(createAssetRsp.Candidators))
	for i, v := range createAssetRsp.Candidators {
		rsp[i] = JsonObject{"CandidateAddr": v.UploadURL, "Token": v.Token}
	}

	c.JSON(http.StatusOK, respJSON(rsp))
}

type createAssetRequest struct {
	AssetName string `json:"asset_name"`
	AssetCID  string `json:"asset_cid"`
	AreaID    string `json:"area_id"`
	NodeID    string `json:"node_id"`
	AssetType string `json:"asset_type"`
	AssetSize int64  `json:"asset_size"`
	GroupId   int64  `json:"group_id"`
}

// CreateAssetPostHandler 创建文件
func CreateAssetPostHandler(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	username := claims[identityKey].(string)
	areaId := getAreaID(c)

	var createAssetReq createAssetRequest
	if err := c.BindJSON(&createAssetReq); err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	// TODO:
	// areaId := GetDefaultTitanCandidateEntrypointInfo()
	schedulerClient, err := getSchedulerClient(c.Request.Context(), areaId)
	// if err != nil {
	// 	c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
	// 	return
	// }

	user, err := dao.GetUserByUsername(c.Request.Context(), username)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	ainfo, _ := dao.GetAssetByCIDAndUser(c.Request.Context(), createAssetReq.AssetCID, username)
	if ainfo != nil && ainfo.ID > 0 {
		c.JSON(http.StatusOK, respErrorCode(errors.FileExists, c))
		return
	}
	// 获取文件hash
	hash, err := storage.CIDToHash(createAssetReq.AssetCID)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	// 判断用户存储空间是否够用
	if user.TotalStorageSize-user.UsedStorageSize < createAssetReq.AssetSize {
		c.JSON(http.StatusOK, respErrorCode(int(terrors.UserStorageSizeNotEnough), c))
		return
	}

	log.Debugf("CreateAssetHandler clientIP:%s, areaId:%s\n", c.ClientIP(), createAssetReq.AreaID)
	createAssetRsp, err := schedulerClient.CreateAsset(c.Request.Context(), &types.CreateAssetReq{AssetProperty: types.AssetProperty{AssetCID: createAssetReq.AssetCID}})
	if err != nil {
		if webErr, ok := err.(*api.ErrWeb); ok {
			c.JSON(http.StatusOK, respErrorCode(webErr.Code, c))
			return
		}
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	if err := dao.AddAssetAndUpdateSize(c.Request.Context(), &model.Asset{
		UserId:    username,
		Name:      createAssetReq.AssetName,
		Cid:       createAssetReq.AssetCID,
		Type:      createAssetReq.AssetType,
		NodeID:    createAssetReq.NodeID,
		TotalSize: createAssetReq.AssetSize,
		GroupID:   createAssetReq.GroupId,
		AreaID:    createAssetReq.AreaID,
		Hash:      hash,
		Event:     -1,
		ProjectId: user.ProjectId,
	}); err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	rsp := make([]JsonObject, len(createAssetRsp.Candidators))
	for i, v := range createAssetRsp.Candidators {
		rsp[i] = JsonObject{"CandidateAddr": v.UploadURL, "Token": v.Token}
	}

	c.JSON(http.StatusOK, respJSON(rsp))
}

func CreateKeyHandler(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	userId := claims[identityKey].(string)
	// userId := c.Query("user_id")
	keyName := c.Query("key_name")
	permsStr := c.Query("perms")
	// areaId := getAreaID(c)
	perms := strings.Split(permsStr, ",")
	acl := make([]types.UserAccessControl, 0, len(perms))
	for _, perm := range perms {
		acl = append(acl, types.UserAccessControl(perm))
	}
	// 获取apikey
	info, err := dao.GetUserByUsername(c.Request.Context(), userId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.UserNotFound, c))
		return
	}
	buf, keyStr, err := storage.CreateAPIKey(c.Request.Context(), userId, keyName, perms, info.ApiKeys)
	if err != nil {
		if webErr, ok := err.(*api.ErrWeb); ok {
			c.JSON(http.StatusOK, respErrorCode(webErr.Code, c))
			return
		}
		c.JSON(http.StatusOK, respErrorCode(errors.NotFound, c))
		return
	}
	err = dao.UpdateUserAPIKeys(c.Request.Context(), info.ID, buf)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"key": keyStr,
	}))
}

func DeleteKeyHandler(c *gin.Context) {
	// userId := c.Query("user_id")
	claims := jwt.ExtractClaims(c)
	userId := claims[identityKey].(string)
	keyName := c.Query("key_name")

	info, err := dao.GetUserByUsername(c.Request.Context(), userId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.UserNotFound, c))
		return
	}
	if len(info.ApiKeys) > 0 {
		keyMaps, err := storage.DecodeAPIKeys(info.ApiKeys)
		if err != nil {
			c.JSON(http.StatusOK, respErrorCode(errors.NotFound, c))
			return
		}
		if _, ok := keyMaps[keyName]; !ok {
			c.JSON(http.StatusOK, respErrorCode(errors.NotFound, c))
			return
		}
		delete(keyMaps, keyName)
		buf, err := storage.EncodeAPIKeys(keyMaps)
		if err != nil {
			c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
			return
		}
		err = dao.UpdateUserAPIKeys(c.Request.Context(), info.ID, buf)
		if err != nil {
			c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
			return
		}
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"msg": "delete success",
	}))
}

// DeleteAssetHandler 删除文件
func DeleteAssetHandler(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	userID := claims[identityKey].(string)
	cid := c.Query("asset_cid")
	areaId := getAreaID(c)
	// 获取文件信息
	asset, err := dao.GetAssetByCIDAndUser(c.Request.Context(), cid, userID)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(int(terrors.NotFound), c))
		return
	}
	// TODO: 调用scheduler接口删除文件
	// areaId := GetDefaultTitanCandidateEntrypointInfo()
	schedulerClient, err := getSchedulerClient(c.Request.Context(), areaId)
	// if err != nil {
	// 	c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
	// 	return
	// }
	err = schedulerClient.RemoveAssetRecord(c.Request.Context(), cid)
	if err != nil {
		if webErr, ok := err.(*api.ErrWeb); ok {
			c.JSON(http.StatusOK, respErrorCode(webErr.Code, c))
			return
		}
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	err = dao.DelAssetAndUpdateSize(c.Request.Context(), cid, userID, asset.TotalSize)
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
	claims := jwt.ExtractClaims(c)
	userId := claims[identityKey].(string)
	cid := c.Query("cid")
	err := dao.UpdateAssetShareStatus(c.Request.Context(), cid, userId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"msg": "success",
	}))
}

type AccessOverview struct {
	AssetRecord      *types.AssetRecord
	UserAssetDetail  *model.Asset
	VisitCount       int64
	RemainVisitCount int64
	FilcoinCount     int64
}

func GetAssetListHandler(c *gin.Context) {
	// userId := c.Query("user_id")
	claims := jwt.ExtractClaims(c)
	userId := claims[identityKey].(string)
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))
	groupId, _ := strconv.Atoi(c.Query("group_id"))
	areaId := getAreaID(c)
	schedulerClient, err := getSchedulerClient(c.Request.Context(), areaId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}
	// createAssetRsp, err := schedulerClient.ListAssets(c.Request.Context(), userId, pageSize, (page-1)*pageSize, groupId)
	// if err != nil {
	// 	if webErr, ok := err.(*api.ErrWeb); ok {
	// 		c.JSON(http.StatusOK, respErrorCode(webErr.Code, c))
	// 		return
	// 	}

	// 	log.Errorf("api ListAssets: %v", err)
	// 	c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
	// 	return
	// }
	createAssetRsp, err := listAssets(c.Request.Context(), schedulerClient, userId, page, pageSize, groupId)
	if err != nil {
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
	// userId := c.Query("user_id")
	claims := jwt.ExtractClaims(c)
	userId := claims[identityKey].(string)
	groupId, _ := strconv.Atoi(c.Query("group_id"))
	areaId := GetDefaultTitanCandidateEntrypointInfo()
	schedulerClient, err := getSchedulerClient(c.Request.Context(), areaId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}
	var total int64
	page, size := 1, 100
	var listRsp []*AssetOverview
loop:
	createAssetRsp, err := listAssets(c.Request.Context(), schedulerClient, userId, size, size, groupId)
	if err != nil {
		log.Errorf("api ListAssets: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	listRsp = append(listRsp, createAssetRsp.AssetOverviews...)
	total += int64(len(createAssetRsp.AssetOverviews))
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

	statusRsp, err := getAssetStatus(c.Request.Context(), userId, cid)
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
	// userId := c.Query("user_id")
	claims := jwt.ExtractClaims(c)
	userId := claims[identityKey].(string)
	groupId, _ := strconv.Atoi(c.Query("group_id"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize = 100
	page = 1
	areaId := c.Query("area_id")
	if areaId == "" {
		areaId = GetDefaultTitanCandidateEntrypointInfo()
	}
	schedulerClient, err := getSchedulerClient(c.Request.Context(), areaId)
	// if err != nil {
	// 	c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
	// 	return
	// }
	// createAssetRsp, err := schedulerClient.ListAssets(c.Request.Context(), userId, pageSize, (page-1)*pageSize, groupId)
	// if err != nil {
	// 	log.Errorf("api ListAssets: %v", err)
	// 	c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
	// 	return
	// }
	total, infos, err := dao.ListAssets(c.Request.Context(), userId, page, pageSize, groupId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	if total == 0 {
		c.JSON(http.StatusOK, respJSON(JsonObject{
			"area_count":      0,
			"candidate_count": 0,
			"edge_count":      0,
		}))
		return
	}
	//

	var deviceIds []string
	deviceExists := make(map[string]int)
	var candidateCount int64
	var edgeCount int64
	for _, data := range infos {
		assetRsp, err := schedulerClient.GetAssetRecord(c.Request.Context(), data.Cid)
		if err != nil {
			if webErr, ok := err.(*api.ErrWeb); ok {
				c.JSON(http.StatusOK, respErrorCode(webErr.Code, c))
				return
			}

			log.Errorf("api GetAssetRecord: %v", err)
			c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
			return
		}
		if len(assetRsp.ReplicaInfos) > 0 {
			for _, rep := range assetRsp.ReplicaInfos {
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
	areaId := getAreaID(c)
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
	// userId := c.Query("user_id")
	claims := jwt.ExtractClaims(c)
	userId := claims[identityKey].(string)

	info, err := dao.GetUserByUsername(c.Request.Context(), userId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.UserNotFound, c))
		return
	}
	var out []map[string]interface{}
	if len(info.ApiKeys) > 0 {
		keyResp, err := storage.DecodeAPIKeys(info.ApiKeys)
		if err != nil {
			c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
			return
		}
		for k, v := range keyResp {
			item := make(map[string]interface{})
			item["name"] = k
			item["key"] = v.APIKey
			item["time"] = v.CreatedTime
			out = append(out, item)
		}
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
	// userId := c.Query("user_id")
	claims := jwt.ExtractClaims(c)
	userId := claims[identityKey].(string)
	keyName := c.Query("key_name")

	var perms []string

	info, err := dao.GetUserByUsername(c.Request.Context(), userId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.UserNotFound, c))
		return
	}
	if len(info.ApiKeys) > 0 {
		keyMap, err := storage.DecodeAPIKeys(info.ApiKeys)
		if err != nil {
			c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
			return
		}
		key, ok := keyMap[keyName]
		if !ok {
			c.JSON(http.StatusOK, respErrorCode(int(terrors.APPKeyNotFound), c))
			return
		}
		payload, err := storage.AuthVerify(key.APIKey)
		if err != nil {
			c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
			return
		}
		for _, v := range payload.AccessControlList {
			perms = append(perms, v)
		}
	} else {
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"perms": perms,
	}))
}

func CreateGroupHandler(c *gin.Context) {
	// userId := c.Query("user_id")
	claims := jwt.ExtractClaims(c)
	userId := claims[identityKey].(string)
	name := c.Query("name")
	parent, _ := strconv.Atoi(c.Query("parent"))
	group, err := dao.CreateAssetGroup(c.Request.Context(), userId, name, parent)
	if err != nil {
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
	// userId := c.Query("user_id")
	claims := jwt.ExtractClaims(c)
	userId := claims[identityKey].(string)
	parent, _ := strconv.Atoi(c.Query("parent"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 100
	}

	rsp, err := dao.ListAssetGroupForUser(c.Request.Context(), userId, parent, pageSize, (page-1)*pageSize)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
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
	// userId := c.Query("user_id")
	claims := jwt.ExtractClaims(c)
	userId := claims[identityKey].(string)
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))
	parentId, _ := strconv.Atoi(c.Query("parent"))
	areaId := getAreaID(c)
	schedulerClient, err := getSchedulerClient(c.Request.Context(), areaId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}
	assetSummary, err := listAssetSummary(c.Request.Context(), schedulerClient, userId, parentId, page, pageSize)
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
	// userId := c.Query("user_id")
	claims := jwt.ExtractClaims(c)
	userId := claims[identityKey].(string)
	groupId, _ := strconv.Atoi(c.Query("group_id"))

	err := dao.DeleteAssetGroup(c.Request.Context(), userId, groupId)
	if err != nil {
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
	// userId := c.Query("user_id")
	claims := jwt.ExtractClaims(c)
	userId := claims[identityKey].(string)
	newName := c.Query("new_name")
	groupId, _ := strconv.Atoi(c.Query("group_id"))

	err := dao.UpdateAssetGroupName(c.Request.Context(), userId, newName, groupId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"msg": "success",
	}))
}

func MoveGroupToGroupHandler(c *gin.Context) {
	// userId := c.Query("user_id")
	claims := jwt.ExtractClaims(c)
	userId := claims[identityKey].(string)
	groupId, _ := strconv.Atoi(c.Query("group_id"))
	targetGroupId, _ := strconv.Atoi(c.Query("target_group_id"))

	err := dao.MoveAssetGroup(c.Request.Context(), userId, groupId, targetGroupId)
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
	// userId := c.Query("user_id")
	claims := jwt.ExtractClaims(c)
	userId := claims[identityKey].(string)
	assetCid := c.Query("asset_cid")
	groupId, _ := strconv.Atoi(c.Query("group_id"))

	err := dao.UpdateAssetGroup(c.Request.Context(), userId, assetCid, groupId)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"msg": "success",
	}))
}
