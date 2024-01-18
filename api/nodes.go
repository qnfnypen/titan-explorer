package api

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/errors"
	"github.com/gnasnik/titan-explorer/core/filecoin"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/pkg/formatter"
	"github.com/go-redis/redis/v9"
	"github.com/golang-module/carbon/v2"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func GetAllAreas(c *gin.Context) {
	areas, err := dao.GetAllAreaFromDeviceInfo(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"areas": areas,
	}))
}

var (
	ChainHeadKey           = "TITAN::FILECOIN::CHAINHEAD"
	ChainHeadKeyExpiration = 10 * time.Second
)

func GetBlockHeightHandler(c *gin.Context) {
	lastTipSet, err := getChainHead(c.Request.Context())
	if err == nil {
		ts := filecoin.GetTimestampByHeight(lastTipSet.Height)
		c.JSON(http.StatusOK, respJSON(JsonObject{
			"height":    lastTipSet.Height,
			"countDown": time.Now().Unix() - ts,
		}))
		return
	}

	tipSet, err := filecoin.ChainHead(config.Cfg.FilecoinRPCServerAddress)
	if err != nil {
		log.Errorf("get chain head: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	if err := setChainHead(c.Request.Context(), tipSet); err != nil {
		log.Errorf("set chain head: %v", err)
	}

	ts := filecoin.GetTimestampByHeight(tipSet.Height)

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"height":    tipSet.Height,
		"countDown": time.Now().Unix() - ts,
	}))
}

func getChainHead(ctx context.Context) (*filecoin.TipSet, error) {
	result, err := dao.RedisCache.Get(ctx, ChainHeadKey).Result()
	if err != nil {
		return nil, err
	}

	var ts filecoin.TipSet
	err = json.Unmarshal([]byte(result), &ts)
	if err != nil {
		return nil, err
	}

	return &ts, nil
}

func setChainHead(ctx context.Context, val interface{}) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}

	_, err = dao.RedisCache.Set(ctx, ChainHeadKey, data, ChainHeadKeyExpiration).Result()
	if err != nil {
		log.Errorf("set chain head: %v", err)
	}

	return nil
}

func GetIndexInfoHandler(c *gin.Context) {
	fullNodeInfo, err := dao.GetCacheFullNodeInfo(c.Request.Context())
	if err != nil {
		log.Errorf("database GetCacheFullNodeInfo: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	c.JSON(http.StatusOK, respJSON(fullNodeInfo))
}

func GetUserDeviceProfileHandler(c *gin.Context) {
	info := &model.DeviceInfo{}
	info.UserID = c.Query("user_id")
	info.DeviceID = c.Query("device_id")
	info.DeviceStatus = c.Query("device_status")
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))
	option := dao.QueryOption{
		Page:      page,
		PageSize:  pageSize,
		StartTime: c.Query("from"),
		EndTime:   c.Query("to"),
	}

	if option.StartTime == "" {
		option.StartTime = time.Now().AddDate(0, 0, -6).Format(formatter.TimeFormatDateOnly)
	}
	if option.EndTime == "" {
		option.EndTime = time.Now().Format(formatter.TimeFormatDateOnly)
	}

	userDeviceProfile, err := dao.CountUserDeviceInfo(c.Request.Context(), info.UserID)
	if err != nil {
		log.Errorf("database CountUserDeviceInfo: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.NotFound, c))
		return
	}
	m, err := dao.GetUserIncome(info, option)
	if err != nil {
		log.Errorf("database GetUserIncome: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.NotFound, c))
		return
	}

	data := toDeviceStatistic(option.StartTime, option.EndTime, m)
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"profile":     userDeviceProfile,
		"series_data": data,
	}))
}

func GetUserDevicesCountHandler(c *gin.Context) {
	info := &model.DeviceInfo{}
	info.UserID = c.Query("user_id")
	info.DeviceID = c.Query("device_id")
	info.DeviceStatus = c.Query("device_status")
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))
	option := dao.QueryOption{
		Page:      page,
		PageSize:  pageSize,
		StartTime: c.Query("from"),
		EndTime:   c.Query("to"),
	}

	if option.StartTime == "" {
		option.StartTime = time.Now().AddDate(0, 0, -6).Format(formatter.TimeFormatDateOnly)
	}
	if option.EndTime == "" {
		option.EndTime = time.Now().Format(formatter.TimeFormatDateOnly)
	}

	userDeviceProfile, err := dao.CountUserDeviceInfo(c.Request.Context(), info.UserID)
	if err != nil {
		log.Errorf("GetUserDevicesCountHandler CountUserDeviceInfo: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.NotFound, c))
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"profile": userDeviceProfile,
	}))
}

func toDeviceStatistic(start, end string, data map[string]map[string]interface{}) []*dao.DeviceStatistics {
	startTime, _ := time.Parse(formatter.TimeFormatDateOnly, start)
	endTime, _ := time.Parse(formatter.TimeFormatDateOnly, end)
	var oneDay = 24 * time.Hour
	var out []*dao.DeviceStatistics
	for startTime.Before(endTime) || startTime.Equal(endTime) {
		key := startTime.Format(formatter.TimeFormatDateOnly)
		startTime = startTime.Add(oneDay)
		val, ok := data[key]
		if !ok {
			out = append(out, &dao.DeviceStatistics{
				Date: key,
			})
			continue
		}
		out = append(out, &dao.DeviceStatistics{
			Date:   key,
			Income: val["income"].(float64),
		})
	}

	return out
}

func queryDeviceStatisticsDaily(deviceID, startTime, endTime string) []*dao.DeviceStatistics {
	option := dao.QueryOption{
		StartTime: startTime,
		EndTime:   endTime,
	}
	if startTime == "" {
		option.StartTime = carbon.Now().SubDays(14).StartOfDay().String()
	}
	if endTime == "" {
		option.EndTime = carbon.Now().EndOfDay().String()
	} else {
		end, _ := time.Parse(formatter.TimeFormatDateOnly, endTime)
		end = end.Add(24 * time.Hour).Add(-time.Second)
		option.EndTime = end.Format(formatter.TimeFormatDatetime)
	}
	condition := &model.DeviceInfoDaily{
		DeviceID: deviceID,
	}

	list, err := dao.GetDeviceInfoDailyList(context.Background(), condition, option)
	if err != nil {
		log.Errorf("database GetDeviceInfoDailyList: %v", err)
		return nil
	}

	return list
}

func queryDeviceDailyByUserId(userId, startTime, endTime string) []*dao.DeviceStatistics {
	option := dao.QueryOption{
		StartTime: startTime,
		EndTime:   endTime,
	}
	if startTime == "" {
		option.StartTime = carbon.Now().SubDays(14).StartOfDay().String()
	}
	if endTime == "" {
		option.EndTime = carbon.Now().EndOfDay().String()
	} else {
		end, _ := time.Parse(formatter.TimeFormatDateOnly, endTime)
		end = end.Add(24 * time.Hour).Add(-time.Second)
		option.EndTime = end.Format(formatter.TimeFormatDatetime)
	}
	condition := &model.DeviceInfoDaily{
		UserID: userId,
	}

	list, err := dao.GetNodesInfoDailyList(context.Background(), condition, option)
	if err != nil {
		log.Errorf("database GetNodesInfoDailyList: %v", err)
		return nil
	}

	return list
}

func queryDeviceStatisticHourly(deviceID, startTime, endTime string) []*dao.DeviceStatistics {
	option := dao.QueryOption{
		StartTime: startTime,
		EndTime:   endTime,
	}
	if option.StartTime == "" {
		option.StartTime = carbon.Now().StartOfHour().SubHours(25).String()
	}
	if option.EndTime == "" {
		option.EndTime = carbon.Now().String()
	} else {
		end, _ := time.Parse(formatter.TimeFormatDateOnly, endTime)
		end = end.Add(1 * time.Hour).Add(-time.Second)
		option.EndTime = end.Format(formatter.TimeFormatDatetime)
	}

	condition := &model.DeviceInfoHour{
		DeviceID: deviceID,
	}
	list, err := dao.GetDeviceInfoDailyHourList(context.Background(), condition, option)
	if err != nil {
		log.Errorf("database GetDeviceInfoDailyHourList: %v", err)
		return nil
	}

	return list
}

func GetQueryInfoHandler(c *gin.Context) {
	info := &model.DeviceInfo{}
	info.UserID = c.Query("key")
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))
	order := c.Query("order")
	orderField := c.Query("order_field")
	option := dao.QueryOption{
		Page:       page,
		PageSize:   pageSize,
		Order:      order,
		OrderField: orderField,
	}
	list, total, err := dao.GetDeviceInfoListByKey(c.Request.Context(), info, option)
	if err != nil {
		log.Errorf("get device by user id info list: %v", err)
	}
	if total < 1 {
		DetailList := dao.GetDeviceInfoById(context.Background(), info.UserID)
		if DetailList.DeviceID != "" {
			list = append(list, &DetailList)
		}
		if len(list) == 0 {
			c.JSON(http.StatusOK, respJSON(JsonObject{
				"type": "wrong key",
			}))
			return
		}
		c.JSON(http.StatusOK, respJSON(JsonObject{
			"list":  list,
			"total": total,
			"type":  "node_id",
		}))
	} else {
		c.JSON(http.StatusOK, respJSON(JsonObject{
			"list":  list,
			"total": total,
			"type":  "user_id",
		}))
	}
}

func GetDeviceInfoHandler(c *gin.Context) {
	info := &model.DeviceInfo{}
	info.UserID = c.Query("user_id")
	info.DeviceID = c.Query("device_id")
	info.IpLocation = c.Query("ip_location")
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))
	order := c.Query("order")
	orderField := c.Query("order_field")
	nodeTypeStr := c.Query("node_type")
	if nodeTypeStr != "" {
		nodeType, _ := strconv.ParseInt(nodeTypeStr, 10, 64)
		info.NodeType = nodeType
	}
	activeStatusStr := c.Query("active_status")
	if activeStatusStr == "" {
		info.ActiveStatus = 10
	} else {
		activeStatus, _ := strconv.ParseInt(activeStatusStr, 10, 64)
		info.ActiveStatus = activeStatus
	}
	deviceStatus := c.Query("device_status")

	if deviceStatus == "online" || deviceStatus == "offline" || deviceStatus == "abnormal" {
		info.DeviceStatus = deviceStatus
	}
	if deviceStatus == "unbinding" || deviceStatus == "unbound" {
		info.BindStatus = deviceStatus
	}
	option := dao.QueryOption{
		Page:       page,
		PageSize:   pageSize,
		Order:      order,
		OrderField: orderField,
	}
	list, total, err := dao.GetDeviceInfoList(c.Request.Context(), info, option)
	if err != nil {
		log.Errorf("database GetDeviceInfoList: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  handleNodeList(c, info.UserID, list),
		"total": total,
	}))
}

func handleNodeList(ctx *gin.Context, userId string, devicesInfo []*model.DeviceInfo) []*model.DeviceInfo {
	areaId := dao.GetAreaID(ctx.Request.Context(), userId)
	schedulerClient, err := getSchedulerClient(ctx, areaId)
	if err != nil {
		log.Errorf("no scheder found")
		return nil
	}
	for _, deviceIfo := range devicesInfo {
		createAssetRsp, err := schedulerClient.GetNodeInfo(ctx, deviceIfo.DeviceID)
		if err != nil {
			log.Errorf("api GetNodeInfo: %v", err)
		}
		deviceIfo.DeactivateTime = createAssetRsp.DeactivateTime
		dao.HandleMapList(ctx, deviceIfo)
	}
	return devicesInfo
}

func GetDeviceActiveInfoHandler(c *gin.Context) {
	info := &model.DeviceInfo{}
	info.UserID = c.Query("user_id")
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))
	order := c.Query("order")
	orderField := c.Query("order_field")
	activeStatusStr := c.Query("active_status")
	if activeStatusStr == "" {
		info.ActiveStatus = 10
	} else {
		activeStatus, _ := strconv.ParseInt(activeStatusStr, 10, 64)
		info.ActiveStatus = activeStatus
	}
	option := dao.QueryOption{
		Page:       page,
		PageSize:   pageSize,
		Order:      order,
		OrderField: orderField,
	}
	list, total, err := dao.GetDeviceActiveInfoList(c.Request.Context(), info, option)
	if err != nil {
		log.Errorf("GetDeviceActiveInfoHandler GetDeviceActiveInfoList: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  list,
		"total": total,
	}))
}

func GetDeviceStatusHandler(c *gin.Context) {
	info := &model.DeviceInfo{}
	info.UserID = c.Query("user_id")
	info.DeviceID = c.Query("device_id")
	info.DeviceStatus = c.Query("device_status")
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))
	order := c.Query("order")
	orderField := c.Query("order_field")
	info.ActiveStatus = 1
	option := dao.QueryOption{
		Page:       page,
		PageSize:   pageSize,
		Order:      order,
		OrderField: orderField,
	}
	list, total, err := dao.GetDeviceInfoList(c.Request.Context(), info, option)
	if err != nil {
		log.Errorf("GetDeviceStatusHandler GetDeviceInfoList: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  list,
		"total": total,
	}))
}

func GetNodesInfoHandler(c *gin.Context) {
	info := &model.DeviceInfo{}
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))
	order := c.Query("order")
	orderField := c.Query("order_field")
	nodeType, _ := strconv.ParseInt(c.Query("node_type"), 10, 64)
	info.NodeType = nodeType
	option := dao.QueryOption{
		Page:       page,
		PageSize:   pageSize,
		Order:      order,
		OrderField: orderField,
	}
	var total int64
	total, list, err := dao.GetNodesInfo(c.Request.Context(), option)
	if err != nil {
		log.Errorf("GetNodesInfoHandler GetNodesInfo: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  handleNodesRank(&list),
		"total": total,
	}))
}

func handleNodesRank(nodes *[]model.NodesInfo) *[]model.NodesInfo {
	var nodesRank []model.NodesInfo
	for i, info := range *nodes {
		rank := strconv.Itoa(i + 1)
		info.Rank = rank
		nodesRank = append(nodesRank, info)
	}
	return &nodesRank
}

func GetMapInfoHandler(c *gin.Context) {
	info := &model.DeviceInfo{}
	info.UserID = c.Query("user_id")
	info.DeviceID = c.Query("device_id")
	info.DeviceStatus = c.Query("device_status")
	pageSize, _ := strconv.Atoi("page_size")
	page, _ := strconv.Atoi("page")
	order := c.Query("order")
	orderField := c.Query("order_field")
	nodeType, _ := strconv.ParseInt(c.Query("node_type"), 10, 64)
	info.NodeType = nodeType
	info.ActiveStatus = 1
	option := dao.QueryOption{
		Page:       page,
		PageSize:   pageSize,
		Order:      order,
		OrderField: orderField,
	}
	list, total, err := dao.GetDeviceInfoList(c.Request.Context(), info, option)
	if err != nil {
		log.Errorf("GetMapInfoHandler GetDeviceInfoList: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  dao.HandleMapInfo(c, list),
		"total": total,
	}))
}

func GetDeviceDiagnosisDailyByDeviceIdHandler(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	deviceID := c.Query("device_id")
	m := queryDeviceStatisticsDaily(deviceID, from, to)
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"series_data": m,
	}))
}

func GetDeviceDiagnosisDailyByUserIdHandler(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	userId := c.Query("user_id")
	m := queryDeviceDailyByUserId(userId, from, to)
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"series_data": m,
	}))
}

func GetDeviceDiagnosisHourHandler(c *gin.Context) {
	deviceID := c.Query("device_id")
	//date := c.Query("date")
	start := c.Query("from")
	end := c.Query("to")
	m := queryDeviceStatisticHourly(deviceID, start, end)
	if len(m) < 1 {
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	deviceInfo, err := dao.GetDeviceInfoByID(c.Request.Context(), deviceID)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"series_data":  m,
		"cpu_cores":    deviceInfo.CpuCores,
		"cpu_usage":    fmt.Sprintf("%.2f", deviceInfo.CpuUsage),
		"memory":       fmt.Sprintf("%.2f", deviceInfo.Memory/float64(10<<20)),
		"memory_usage": fmt.Sprintf("%.2f", deviceInfo.MemoryUsage*deviceInfo.Memory/float64(10<<20)),
		"disk_usage":   fmt.Sprintf("%.2f", (deviceInfo.DiskUsage*deviceInfo.DiskSpace/100)/float64(10<<30)),
		"disk_space":   fmt.Sprintf("%.2f", deviceInfo.DiskSpace/float64(10<<30)),
		"disk_type":    deviceInfo.DiskType,
		"file_system":  deviceInfo.IoSystem,
		"w":            []float64{},
	}))
}

func GetDeviceInfoDailyHandler(c *gin.Context) {
	cond := &model.DeviceInfoDaily{}
	cond.DeviceID = c.Query("device_id")
	pageSize, _ := strconv.Atoi("page_size")
	page, _ := strconv.Atoi("page")
	option := dao.QueryOption{
		Page:       page,
		PageSize:   pageSize,
		OrderField: "created_at",
		Order:      "DESC",
	}

	list, total, err := dao.GetDeviceInfoDailyByPage(context.Background(), cond, option)
	if err != nil {
		log.Errorf("get device info daily: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  list,
		"total": total,
	}))
}

func GetDiskDaysHandler(c *gin.Context) {
	//date := c.Query("date")
	start := c.Query("from")
	end := c.Query("to")
	m := dao.QueryNodesDailyInfo(start, end)
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"series_data": m,
	}))
}

func GetDeviceProfileHandler(c *gin.Context) {
	type getEarningReq struct {
		NodeID string   `json:"node_id"`
		Keys   []string `json:"keys"`
		Since  int64    `json:"since"`
	}

	var param getEarningReq
	if err := c.BindJSON(&param); err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	response := make(map[string]interface{})

	lastUpdate, err := dao.GetCacheFullNodeInfo(c.Request.Context())
	if err != nil {
		log.Errorf("get last update info: %v", err)
	}

	if lastUpdate != nil && param.Since > 0 {
		sinceT := time.Unix(param.Since, 0)
		if lastUpdate.Time.Before(sinceT) {
			response["since"] = time.Now().Unix()
			c.JSON(http.StatusOK, respJSON(response))
			return
		}
	}

	deviceInfo, err := dao.GetDeviceInfo(c.Request.Context(), param.NodeID)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	for _, key := range param.Keys {
		switch key {
		case "income":
			response[key] = map[string]interface{}{
				"today": deviceInfo.TodayProfit,
				"total": deviceInfo.CumulativeProfit,
			}
		case "online":
			response[key] = map[string]interface{}{
				"today": deviceInfo.TodayOnlineTime,
				"total": deviceInfo.OnlineTime,
			}
		case "day_incomes":
			response[key] = queryHourlyIncome(c.Request.Context(), param.NodeID)
		case "month_incomes":
			response[key] = queryDailyIncome(c.Request.Context(), param.NodeID)
		}
	}

	if param.Since > 0 {
		filterResp, err := filterResponse(c.Request.Context(), param.NodeID, response)
		if err != nil {
			log.Errorf("filter response: %v", err)
		}

		if filterResp != nil {
			response = filterResp
		}
	}

	response["since"] = time.Now().Unix()

	c.JSON(http.StatusOK, respJSON(response))
}

func filterResponse(ctx context.Context, nodeId string, response map[string]interface{}) (map[string]interface{}, error) {
	deviceCache, err := dao.GetDeviceProfileFromCache(ctx, nodeId)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	out := make(map[string]interface{})
	devHash := make(map[string]string)

	for key, val := range response {
		encodeData, err := json.Marshal(val)
		if err != nil {
			log.Errorf("encode %s: %v", key, err)
			continue
		}

		hasher := md5.New()
		hasher.Write(encodeData)
		hash := hex.EncodeToString(hasher.Sum(nil))
		checksum := deviceCache[key]

		if checksum != hash {
			out[key] = val
		}

		devHash[key] = hash
	}

	err = dao.SetDeviceProfileFromCache(ctx, nodeId, devHash)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func queryHourlyIncome(ctx context.Context, nodeId string) interface{} {
	start := carbon.Now().StartOfDay().String()
	option := dao.QueryOption{
		StartTime: start,
	}

	list, err := dao.GetDeviceHourlyIncome(context.Background(), nodeId, option)
	if err != nil {
		log.Errorf("database GetDeviceInfoDailyHourList: %v", err)
		return nil
	}

	out := make([]interface{}, 0)
	for _, item := range list {
		out = append(out, map[string]interface{}{
			"k": fmt.Sprintf("%s:00", strings.TrimLeft(item.Date, " ")),
			"v": item.Income,
		})
	}

	return out
}

func queryDailyIncome(ctx context.Context, nodeId string) interface{} {
	start := carbon.Now().SubDays(30).String()

	option := dao.QueryOption{
		StartTime: start,
	}

	condition := &model.DeviceInfoDaily{
		DeviceID: nodeId,
	}

	list, err := dao.GetDeviceInfoDailyList(context.Background(), condition, option)
	if err != nil {
		log.Errorf("database GetDeviceInfoDailyList: %v", err)
		return nil
	}

	out := make([]interface{}, 0)
	for _, item := range list {
		out = append(out, map[string]interface{}{
			"k": item.Date,
			"v": item.Income,
		})
	}

	return out
}
