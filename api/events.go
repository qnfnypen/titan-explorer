package api

import (
	"fmt"
	"github.com/Filecoin-Titan/titan/api/types"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/errors"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/utils"
	"net/http"
	"strconv"
	"time"
)

func GetCacheListHandlerold(c *gin.Context) {
	info := &model.CacheEvent{
		DeviceID: c.Query("device_id"),
	}
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))
	order := c.Query("order")
	orderField := c.Query("order_field")
	option := dao.QueryOption{
		Page:       page,
		PageSize:   pageSize,
		OrderField: orderField,
		Order:      order,
	}

	list, total, err := dao.GetCacheEventsByPage(c.Request.Context(), info, option)
	if err != nil {
		log.Errorf("get cache events by page: %v", err)
		c.JSON(http.StatusOK, respError(errors.ErrNotFound))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  list,
		"total": total,
	}))
}

func GetCacheListHandler(c *gin.Context) {
	startT := time.Now().Format(utils.TimeFormatDatetime)
	fmt.Println("start get get_cache_list:", startT)
	nodeId := c.Query("device_id")
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))
	resp, err := schedulerApi.GetReplicaEventsForNode(c.Request.Context(), nodeId, pageSize, (page-1)*pageSize)
	if err != nil {
		log.Errorf("api GetReplicaEventsForNode: %v", err)
		c.JSON(http.StatusOK, respError(errors.ErrInternalServer))
		return
	}
	startT = time.Now().Format(utils.TimeFormatDatetime)
	fmt.Println("api GetReplicaEventsForNode:", startT)
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  resp.ReplicaEvents,
		"total": resp.Total,
	}))
	startT = time.Now().Format(utils.TimeFormatDatetime)
	fmt.Println("request end :", startT)
	return
}

func GetValidationListHandler(c *gin.Context) {
	startT := time.Now().Format(utils.TimeFormatDatetime)
	fmt.Println("start get get_validation_list:", startT)
	nodeId := c.Query("device_id")
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))
	resp, err := schedulerApi.GetValidationResults(c.Request.Context(), nodeId, pageSize, (page-1)*pageSize)
	if err != nil {
		log.Errorf("api GetValidationResults: %v", err)
		c.JSON(http.StatusOK, respError(errors.ErrInternalServer))
		return
	}
	startT = time.Now().Format(utils.TimeFormatDatetime)
	fmt.Println("api GetValidationResults:", startT)
	var validationEvents []*model.ValidationEvent
	for _, blockInfo := range resp.ValidationResultInfos {
		validationEvents = append(validationEvents, toValidationEvent(blockInfo))
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  validationEvents,
		"total": resp.Total,
	}))
	startT = time.Now().Format(utils.TimeFormatDatetime)
	fmt.Println("request end :", startT)
	return
}

func GetAllocateStorageHandler(c *gin.Context) {
	UserId := c.Query("user_id")
	fmt.Println(UserId)
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
		c.JSON(http.StatusOK, respError(errors.ErrInternalServer))
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
		c.JSON(http.StatusOK, respError(err))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"PeakBandwidth": StorageSize.PeakBandwidth,
		"TotalTraffic":  StorageSize.TotalTraffic,
		"TotalSize":     StorageSize.TotalSize,
		"UsedSize":      StorageSize.UsedSize,
	}))
	return
}

func CreateAssetHandler(c *gin.Context) {
	UserId := c.Query("user_id")
	AssetCID := c.Query("asset_cid")
	AssetName := c.Query("asset_name")
	AssetSize := c.Query("asset_size")
	areaId := dao.GetAreaID(c.Request.Context(), UserId)
	schedulerClient := GetNewScheduler(c.Request.Context(), areaId)
	var createAssetReq types.CreateAssetReq
	createAssetReq.AssetName = AssetName
	createAssetReq.AssetCID = AssetCID
	createAssetReq.UserID = UserId
	createAssetReq.AssetSize = utils.Str2Int64(AssetSize)
	createAssetRsp, err := schedulerClient.CreateAsset(c.Request.Context(), &createAssetReq)
	if createAssetRsp.AlreadyExists {
		c.JSON(http.StatusOK, respError(errors.ErrFileExists))
		return
	}
	if err != nil {
		log.Errorf("api CreateAsset: %v", err)
		c.JSON(http.StatusOK, respError(errors.ErrUnknown))
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
		c.JSON(http.StatusOK, respError(errors.ErrKeyLimit))
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
		c.JSON(http.StatusOK, respError(errors.ErrNotFound))
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"msg": "delete success",
	}))
	return
}

func DeleteAssetHandler(c *gin.Context) {
	UserId := c.Query("user_id")
	cid := c.Query("asset_cid")
	areaId := dao.GetAreaID(c.Request.Context(), UserId)
	schedulerClient := GetNewScheduler(c.Request.Context(), areaId)
	err := schedulerClient.DeleteAsset(c.Request.Context(), UserId, cid)
	if err != nil {
		log.Errorf("api DeleteAsset: %v", err)
		c.JSON(http.StatusOK, respError(errors.ErrInternalServer))
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
	// schedulerApi
	// get area id
	areaId := dao.GetAreaID(c.Request.Context(), UserId)
	schedulerClient := GetNewScheduler(c.Request.Context(), areaId)
	urls, err := schedulerClient.ShareAssets(c.Request.Context(), UserId, assetCIDs)
	if err != nil {
		log.Errorf("api ShareAssets: %v", err)
		c.JSON(http.StatusOK, respError(errors.ErrInternalServer))
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"asset_cid": Cid,
		"url":       urls[Cid],
	}))
}

func GetAssetListHandler(c *gin.Context) {
	startT := time.Now().Format(utils.TimeFormatDatetime)
	fmt.Println("start get get_asset_list:", startT)
	UserId := c.Query("user_id")
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))
	areaId := dao.GetAreaID(c.Request.Context(), UserId)
	schedulerClient := GetNewScheduler(c.Request.Context(), areaId)
	startT = time.Now().Format(utils.TimeFormatDatetime)
	fmt.Println("api GetNewScheduler:", startT)
	createAssetRsp, err := schedulerClient.ListAssets(c.Request.Context(), UserId, pageSize, (page-1)*pageSize)
	if err != nil {
		log.Errorf("api ListAssets: %v", err)
		c.JSON(http.StatusOK, respError(errors.ErrInternalServer))
		return
	}
	startT = time.Now().Format(utils.TimeFormatDatetime)
	fmt.Println("api ListAssets:", startT)
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  createAssetRsp.AssetRecords,
		"total": createAssetRsp.Total,
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
		c.JSON(http.StatusOK, respError(errors.ErrInternalServer))
		return
	}
	var deviceIds []string
	deviceExists := make(map[string]int)
	var CandidateCount int64
	var edgeCount int64
	for _, data := range createAssetRsp.AssetRecords {
		if len(data.ReplicaInfos) > 0 {
			for _, rep := range data.ReplicaInfos {
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
	countArea, err := dao.GetAreaCount(c.Request.Context(), deviceIds)
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"area_count":      countArea,
		"candidate_count": CandidateCount,
		"edge_count":      edgeCount,
	}))
}

func GetAssetInfoHandler(c *gin.Context) {
	UserId := c.Query("user_id")
	Cid := c.Query("cid")
	areaId := dao.GetAreaID(c.Request.Context(), UserId)
	schedulerClient := GetNewScheduler(c.Request.Context(), areaId)
	AssetRsp, err := schedulerClient.GetAssetRecord(c.Request.Context(), Cid)
	if err != nil {
		log.Errorf("api ListAssets: %v", err)
		c.JSON(http.StatusOK, respError(errors.ErrInternalServer))
		return
	}
	var deviceIds []string
	if len(AssetRsp.ReplicaInfos) > 0 {
		for _, rep := range AssetRsp.ReplicaInfos {
			deviceIds = append(deviceIds, rep.NodeID)
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
		c.JSON(http.StatusOK, respError(errors.ErrInternalServer))
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
	startT := time.Now().Format(utils.TimeFormatDatetime)
	fmt.Println("start get get_retrieval_list:", startT)
	nodeId := c.Query("device_id")
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))
	resp, err := schedulerApi.GetRetrieveEventRecords(c.Request.Context(), nodeId, pageSize, (page-1)*pageSize)
	if err != nil {
		log.Errorf("api GetWorkloadRecords: %v", err)
		c.JSON(http.StatusOK, respError(errors.ErrInternalServer))
		return
	}
	startT = time.Now().Format(utils.TimeFormatDatetime)
	fmt.Println("api GetRetrieveEventRecords:", startT)
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  resp.RetrieveEventInfos,
		"total": resp.Total,
	}))
	startT = time.Now().Format(utils.TimeFormatDatetime)
	fmt.Println("request end :", startT)
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
	//date := c.Query("date")
	start := c.Query("from")
	end := c.Query("to")
	m := dao.QueryCacheHour(deviceID, start, end)
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"series_data": m,
	}))
}

func GetCacheDaysHandler(c *gin.Context) {
	deviceID := c.Query("device_id")
	//date := c.Query("date")
	start := c.Query("from")
	end := c.Query("to")
	m := dao.QueryCacheDaily(deviceID, start, end)
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"series_data": m,
	}))
}

//func GetNewScheduler(ctx context.Context, areaId string) api.Scheduler {
//	scheduler, _ := SchedulerConfigs[areaId]
//	if len(scheduler) < 1 {
//		scheduler = SchedulerConfigs["Asia-China-Guangdong-Shenzhen"]
//	}
//	schedulerApiUrl := scheduler[0].SchedulerURL
//	schedulerApiToken := scheduler[0].AccessToken
//	SchedulerURL := strings.Replace(schedulerApiUrl, "https", "http", 1)
//	headers := http.Header{}
//	headers.Add("Authorization", "Bearer "+schedulerApiToken)
//	schedulerClient, _, err := client.NewScheduler(ctx, SchedulerURL, headers)
//	if err != nil {
//		log.Errorf("create scheduler rpc client: %v", err)
//	}
//	return schedulerClient
//}
