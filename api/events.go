package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Filecoin-Titan/titan/api"
	"github.com/Filecoin-Titan/titan/api/types"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/errors"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/utils"
)

func GetCacheListHandler(c *gin.Context) {
	nodeId := c.Query("device_id")
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))
	resp, err := schedulerApi.GetReplicaEventsForNode(c.Request.Context(), nodeId, pageSize, (page-1)*pageSize)
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
	resp, err := schedulerApi.GetValidationResults(c.Request.Context(), nodeId, pageSize, (page-1)*pageSize)
	if err != nil {
		log.Errorf("api GetValidationResults: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	var validationEvents []*model.ValidationEvent
	for _, blockInfo := range resp.ValidationResultInfos {
		validationEvents = append(validationEvents, toValidationEvent(blockInfo))
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  validationEvents,
		"total": resp.Total,
	}))
	return
}

func GetAllocateStorageHandler(c *gin.Context) {
	UserId := c.Query("user_id")
	if UserId == "" {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}
	var userInfo model.User
	userInfo.Username = UserId
	_, err := dao.GetUserByUsername(c.Request.Context(), userInfo.Username)
	if err == nil {
		log.Info("GetUserByUsername user exists")
	} else {
		_ = dao.CreateUser(c.Request.Context(), &userInfo)
	}

	areaId := dao.GetAreaID(c.Request.Context(), UserId)
	schedulerClient := GetNewScheduler(c.Request.Context(), areaId)
	_, err = schedulerClient.AllocateStorage(c.Request.Context(), UserId)
	if err != nil {
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
	UserId := c.Query("user_id")
	areaId := dao.GetAreaID(c.Request.Context(), UserId)
	schedulerClient := GetNewScheduler(c.Request.Context(), areaId)
	StorageSize, err := schedulerClient.GetUserInfo(c.Request.Context(), UserId)
	if err != nil {
		log.Errorf("api GetStorageSize: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.NotFound, c))
		return
	}
	peakBandwidth := GetUserInfo(c.Request.Context(), UserId)
	if peakBandwidth > StorageSize.PeakBandwidth {
		StorageSize.PeakBandwidth = peakBandwidth
	} else {
		var expireTime time.Duration
		expireTime = time.Hour
		// update redis data
		_ = SetUserInfo(c.Request.Context(), UserId, StorageSize.PeakBandwidth, expireTime)
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"PeakBandwidth": StorageSize.PeakBandwidth,
		"TotalTraffic":  StorageSize.TotalTraffic,
		"TotalSize":     StorageSize.TotalSize,
		"UsedSize":      StorageSize.UsedSize,
	}))
	return
}

func GetUserVipInfoHandler(c *gin.Context) {
	UserId := c.Query("user_id")
	areaId := dao.GetAreaID(c.Request.Context(), UserId)
	schedulerClient := GetNewScheduler(c.Request.Context(), areaId)
	StorageSize, err := schedulerClient.GetUserInfo(c.Request.Context(), UserId)
	if err != nil {
		log.Errorf("api GetUserInfo: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.NotFound, c))
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"vip": StorageSize.EnableVIP,
	}))
	return
}

func GetUserAccessTokenHandler(c *gin.Context) {
	UserId := c.Query("user_id")
	areaId := dao.GetAreaID(c.Request.Context(), UserId)
	schedulerClient := GetNewScheduler(c.Request.Context(), areaId)
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

func CreateAssetHandler(c *gin.Context) {
	UserId := c.Query("user_id")
	areaId := dao.GetAreaID(c.Request.Context(), UserId)
	schedulerClient := GetNewScheduler(c.Request.Context(), areaId)

	nodeIPInfos, err := schedulerClient.GetCandidateIPs(c.Request.Context())
	if err != nil {
		log.Warnf("get candidate ips error %s", err.Error())
	}

	var nearestNode string
	if len(nodeIPInfos) > 0 {
		nodeMap := make(map[string]string)
		ips := make([]string, 0, len(nodeIPInfos))
		for _, nodeIPInfo := range nodeIPInfos {
			ips = append(ips, nodeIPInfo.IP)
			nodeMap[nodeIPInfo.IP] = nodeIPInfo.NodeID
		}

		if ip, err := GetUserNearestIP(c.Request.Context(), c.ClientIP(), ips, NewIPCoordinate()); err == nil {
			nearestNode = nodeMap[ip]
		} else {
			log.Warnf("GetUserNearestIP error %s", err.Error())
		}
	}

	log.Debugf("CreateAssetHandler clientIP:%s, areaId:%s, nearestNode:%s\n", c.ClientIP(), areaId, nearestNode)

	var createAssetReq types.CreateAssetReq
	createAssetReq.AssetName = c.Query("asset_name")
	createAssetReq.AssetCID = c.Query("asset_cid")
	createAssetReq.UserID = UserId
	createAssetReq.AssetType = c.Query("asset_type")
	createAssetReq.AssetSize = utils.Str2Int64(c.Query("asset_size"))
	createAssetReq.NodeID = nearestNode
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
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"CandidateAddr": createAssetRsp.UploadURL,
		"Token":         createAssetRsp.Token,
	}))
	return
}

func CreateKeyHandler(c *gin.Context) {
	UserId := c.Query("user_id")
	KeyName := c.Query("key_name")
	areaId := dao.GetAreaID(c.Request.Context(), UserId)
	schedulerClient := GetNewScheduler(c.Request.Context(), areaId)
	keyStr, err := schedulerClient.CreateAPIKey(c.Request.Context(), UserId, KeyName)
	if err != nil {
		log.Errorf("api CreateAPIKey: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.KeyLimit, c))
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"key": keyStr,
	}))
	return
}

func DeleteKeyHandler(c *gin.Context) {
	UserId := c.Query("user_id")
	KeyName := c.Query("key_name")
	areaId := dao.GetAreaID(c.Request.Context(), UserId)
	schedulerClient := GetNewScheduler(c.Request.Context(), areaId)
	err := schedulerClient.DeleteAPIKey(c.Request.Context(), UserId, KeyName)
	if err != nil {
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
	areaId := dao.GetAreaID(c.Request.Context(), UserId)
	schedulerClient := GetNewScheduler(c.Request.Context(), areaId)
	err := schedulerClient.DeleteAsset(c.Request.Context(), UserId, cid)
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
	UserId := c.Query("user_id")
	Cid := c.Query("asset_cid")
	var assetCIDs []string
	assetCIDs = append(assetCIDs, Cid)
	areaId := dao.GetAreaID(c.Request.Context(), UserId)
	schedulerClient := GetNewScheduler(c.Request.Context(), areaId)
	urls, err := schedulerClient.ShareAssets(c.Request.Context(), UserId, assetCIDs)
	if err != nil {
		log.Errorf("api ShareAssets: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	var redirect bool
	if UserId == "f1hl7vbopivazgion4ql25opmjnoj2ldfsvm5fuzi" {
		redirect = false
	} else {
		redirect = true
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"asset_cid": Cid,
		"url":       urls[Cid],
		"redirect":  redirect,
	}))
}

func ShareLinkHandler(c *gin.Context) {
	Username := c.Query("username")
	Cid := c.Query("cid")
	Url := c.Query("url")
	if Cid == "" || Url == "" {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}
	var link model.Link
	link.UserName = Username
	link.Cid = Cid
	link.LongLink = Url
	shortLink := dao.GetShortLink(c.Request.Context(), Url)
	if shortLink == "" {
		link.ShortLink = "/link?" + "cid=" + Cid
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
	Cid := c.Query("cid")
	if Cid == "" {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}
	link := dao.GetLongLink(c.Request.Context(), Cid)
	if link == "" {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}
	c.Redirect(http.StatusMovedPermanently, link)

}

func UpdateShareStatusHandler(c *gin.Context) {
	UserId := c.Query("user_id")
	Cid := c.Query("cid")
	areaId := dao.GetAreaID(c.Request.Context(), UserId)
	schedulerClient := GetNewScheduler(c.Request.Context(), areaId)
	err := schedulerClient.UpdateShareStatus(c.Request.Context(), UserId, Cid)
	if err != nil {
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
	UserId := c.Query("user_id")
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))
	areaId := dao.GetAreaID(c.Request.Context(), UserId)
	schedulerClient := GetNewScheduler(c.Request.Context(), areaId)
	createAssetRsp, err := schedulerClient.ListAssets(c.Request.Context(), UserId, pageSize, (page-1)*pageSize)
	if err != nil {
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
	UserId := c.Query("user_id")
	areaId := dao.GetAreaID(c.Request.Context(), UserId)
	schedulerClient := GetNewScheduler(c.Request.Context(), areaId)
	var total int
	page, size := 1, 100
	var listRsp []*types.AssetOverview
loop:
	createAssetRsp, err := schedulerClient.ListAssets(c.Request.Context(), UserId, size, (page-1)*size)
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
	UserId := c.Query("username")
	Cid := c.Query("cid")
	areaId := dao.GetAreaID(c.Request.Context(), UserId)
	schedulerClient := GetNewScheduler(c.Request.Context(), areaId)
	statusRsp, err := schedulerClient.GetAssetStatus(c.Request.Context(), UserId, Cid)
	if err != nil {
		log.Errorf("api GetAssetStatus: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"data": statusRsp,
	}))
}

func GetAssetCountHandler(c *gin.Context) {
	UserId := c.Query("user_id")
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize = 100
	page = 1
	areaId := dao.GetAreaID(c.Request.Context(), UserId)
	schedulerClient := GetNewScheduler(c.Request.Context(), areaId)
	createAssetRsp, err := schedulerClient.ListAssets(c.Request.Context(), UserId, pageSize, (page-1)*pageSize)
	if err != nil {
		log.Errorf("api ListAssets: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	var deviceIds []string
	deviceExists := make(map[string]int)
	var CandidateCount int64
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
					CandidateCount += 1
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
		"candidate_count": CandidateCount,
		"edge_count":      edgeCount,
	}))
}

func GetCarFileCountHandler(c *gin.Context) {
	userId := c.Query("user_id")
	cid := c.Query("cid")
	lang := model.Language(c.GetHeader("Lang"))
	areaId := dao.GetAreaID(c.Request.Context(), userId)
	schedulerClient := GetNewScheduler(c.Request.Context(), areaId)
	assetRsp, err := schedulerClient.GetAssetRecord(c.Request.Context(), cid)
	if err != nil {
		log.Errorf("api GetAssetRecord: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	var deviceIdAll []string
	deviceExists := make(map[string]int)
	if len(assetRsp.ReplicaInfos) > 0 {
		for _, rep := range assetRsp.ReplicaInfos {
			if rep.Status == 3 {
				deviceIdAll = append(deviceIdAll, rep.NodeID)
				continue
			}
		}
	}

	assetListAll, e := dao.GetAssetList(c.Request.Context(), deviceIdAll, lang)
	if err != nil {
		log.Errorf("GetAssetList err: %v", e)
	}
	for _, nodeInfo := range assetListAll {
		if _, ok := deviceExists[nodeInfo.IpCity]; ok {
			continue
		}
		deviceExists[nodeInfo.IpCity] = 1
	}

	filReplicas, err := dao.CountFilStorage(c.Request.Context(), cid)
	if err != nil {
		log.Errorf("count fil storage: %v", err)
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"cid":               assetRsp.CID,
		"cid_name":          "",
		"ReplicaInfo_count": len(deviceIdAll),
		"area_count":        len(deviceExists),
		"titan_count":       len(deviceIdAll),
		"fileCoin_count":    filReplicas,
	}))
}

func GetLocationHandler(c *gin.Context) {
	UserId := c.Query("user_id")
	Cid := c.Query("cid")
	lang := model.Language(c.GetHeader("Lang"))
	areaId := dao.GetAreaID(c.Request.Context(), UserId)
	schedulerClient := GetNewScheduler(c.Request.Context(), areaId)
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))
	GetRsp, err := schedulerClient.GetReplicas(c.Request.Context(), Cid, pageSize, (page-1)*pageSize)
	if err != nil {
		log.Errorf("api GetReplicas: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	var deviceIds []string
	if len(GetRsp.ReplicaInfos) > 0 {
		for _, rep := range GetRsp.ReplicaInfos {
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
		AssetList, err := dao.GetAssetList(c.Request.Context(), deviceIds, lang)
		if err != nil {
			log.Errorf("GetAssetList err: %v", err)
		}
		for _, nodeInfo := range AssetList {
			assetInfos = append(assetInfos, &DeviceInfoRes{
				DeviceId:   nodeInfo.DeviceID,
				IpLocation: contactIPLocation(nodeInfo.Location, lang),
				Status:     nodeInfo.DeviceStatus,
			})
		}
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"total":     GetRsp.Total,
		"node_list": assetInfos,
	}))
}

func contactIPLocation(loc model.Location, lang model.Language) string {
	var unknown string
	switch lang {
	case model.LanguageCN:
		unknown = "未知"
	default:
		unknown = "Unknown"
	}

	cf := func(in string) string {
		if in == "" {
			return unknown
		}
		return in
	}

	return fmt.Sprintf("%s-%s-%s-%s", cf(loc.Continent), cf(loc.Country), cf(loc.Province), cf(loc.City))
}

func GetMapByCidHandler(c *gin.Context) {
	UserId := c.Query("user_id")
	Cid := c.Query("cid")
	lang := model.Language(c.GetHeader("Lang"))
	areaId := dao.GetAreaID(c.Request.Context(), UserId)
	schedulerClient := GetNewScheduler(c.Request.Context(), areaId)
	AssetRsp, err := schedulerClient.GetAssetRecord(c.Request.Context(), Cid)
	if err != nil {
		log.Errorf("api GetAssetRecord: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	var deviceIds []string
	if len(AssetRsp.ReplicaInfos) > 0 {
		for _, rep := range AssetRsp.ReplicaInfos {
			if rep.Status == 3 {
				deviceIds = append(deviceIds, rep.NodeID)
			}
		}
	}
	AssetList, e := dao.GetAssetList(c.Request.Context(), deviceIds, lang)
	if err != nil {
		log.Errorf("GetAssetList err: %v", e)
	}
	mapList := dao.HandleMapInfo(c, AssetList)
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  mapList,
		"total": len(mapList),
	}))
}

func GetAssetInfoHandler(c *gin.Context) {
	UserId := c.Query("user_id")
	Cid := c.Query("cid")
	areaId := dao.GetAreaID(c.Request.Context(), UserId)
	schedulerClient := GetNewScheduler(c.Request.Context(), areaId)
	AssetRsp, err := schedulerClient.GetAssetRecord(c.Request.Context(), Cid)
	if err != nil {
		log.Errorf("api GetAssetRecord: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	var deviceIds []string
	if len(AssetRsp.ReplicaInfos) > 0 {
		for _, rep := range AssetRsp.ReplicaInfos {
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
	UserId := c.Query("user_id")
	areaId := dao.GetAreaID(c.Request.Context(), UserId)
	schedulerClient := GetNewScheduler(c.Request.Context(), areaId)
	keyRsp, err := schedulerClient.GetAPIKeys(c.Request.Context(), UserId)
	if err != nil {
		log.Errorf("api GetAPIKeys: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	var Rsp []map[string]interface{}
	for k, v := range keyRsp {
		rsp := make(map[string]interface{})
		rsp["name"] = k
		rsp["key"] = v.APIKey
		rsp["time"] = v.CreatedTime
		Rsp = append(Rsp, rsp)
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list": Rsp,
	}))
}

func GetRetrievalListHandler(c *gin.Context) {
	nodeId := c.Query("device_id")
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))
	resp, err := schedulerApi.GetRetrieveEventRecords(c.Request.Context(), nodeId, pageSize, (page-1)*pageSize)
	if err != nil {
		log.Errorf("api GetRetrieveEventRecords: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  resp.RetrieveEventInfos,
		"total": resp.Total,
	}))
	return
}

func toValidationEvent(in types.ValidationResultInfo) *model.ValidationEvent {
	return &model.ValidationEvent{
		DeviceID:        in.NodeID,
		ValidatorID:     in.ValidatorID,
		Status:          int32(in.Status),
		Blocks:          in.BlockNumber,
		Time:            in.StartTime,
		Duration:        in.Duration,
		UpstreamTraffic: utils.ToFixed(float64(in.Duration)*in.Bandwidth, 2),
	}
}

func GetCacheHourHandler(c *gin.Context) {
	deviceID := c.Query("device_id")
	start := c.Query("from")
	end := c.Query("to")
	dataHour := dao.QueryCacheHour(deviceID, start, end)
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"series_data": dataHour,
	}))
}

func GetCacheDaysHandler(c *gin.Context) {
	deviceID := c.Query("device_id")
	start := c.Query("from")
	end := c.Query("to")
	dataDaily := dao.QueryCacheDaily(deviceID, start, end)
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"series_data": dataDaily,
	}))
}
