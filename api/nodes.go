package api

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/errors"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/utils"
	"github.com/golang-module/carbon/v2"
	"net/http"
	"strconv"
	"time"
)

func GetAllAreas(c *gin.Context) {
	areas, err := dao.GetAllAreaFromDeviceInfo(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, respError(errors.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"areas": areas,
	}))
}

func GetIndexInfoHandler(c *gin.Context) {
	fullNodeInfo, err := dao.GetCacheFullNodeInfo(c.Request.Context())
	if err != nil {
		log.Errorf("get full node info: %v", err)
		c.JSON(http.StatusOK, respError(errors.ErrInternalServer))
		return
	}
	c.JSON(http.StatusOK, respJSON(fullNodeInfo))
}

// GetUserDeviceProfileHandler devices overview
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
		option.StartTime = time.Now().AddDate(0, 0, -6).Format(utils.TimeFormatDateOnly)
	}
	if option.EndTime == "" {
		option.EndTime = time.Now().Format(utils.TimeFormatDateOnly)
	}

	userDeviceProfile, err := dao.CountUserDeviceInfo(c.Request.Context(), info.UserID)
	if err != nil {
		log.Errorf("get user device profile: %v", err)
		c.JSON(http.StatusOK, respError(errors.ErrNotFound))
		return
	}
	m, err := dao.GetUserIncome(c.Request.Context(), info, option)
	if err != nil {
		log.Errorf("get user income: %v", err)
		c.JSON(http.StatusOK, respError(errors.ErrNotFound))
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
		option.StartTime = time.Now().AddDate(0, 0, -6).Format(utils.TimeFormatDateOnly)
	}
	if option.EndTime == "" {
		option.EndTime = time.Now().Format(utils.TimeFormatDateOnly)
	}

	userDeviceProfile, err := dao.CountUserDeviceInfo(c.Request.Context(), info.UserID)
	if err != nil {
		log.Errorf("get user device profile: %v", err)
		c.JSON(http.StatusOK, respError(errors.ErrNotFound))
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"profile": userDeviceProfile,
	}))
}

func toDeviceStatistic(start, end string, data map[string]map[string]interface{}) []*dao.DeviceStatistics {
	startTime, _ := time.Parse(utils.TimeFormatDateOnly, start)
	endTime, _ := time.Parse(utils.TimeFormatDateOnly, end)

	var oneDay = 24 * time.Hour
	var out []*dao.DeviceStatistics
	for startTime.Before(endTime) || startTime.Equal(endTime) {
		key := startTime.Format(utils.TimeFormatDateOnly)
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
		end, _ := time.Parse(utils.TimeFormatDateOnly, endTime)
		end = end.Add(24 * time.Hour).Add(-time.Second)
		option.EndTime = end.Format(utils.TimeFormatDatetime)
	}
	condition := &model.DeviceInfoDaily{
		DeviceID: deviceID,
	}

	list, err := dao.GetDeviceInfoDailyList(context.Background(), condition, option)
	if err != nil {
		log.Errorf("get incoming daily: %v", err)
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
		end, _ := time.Parse(utils.TimeFormatDateOnly, endTime)
		end = end.Add(24 * time.Hour).Add(-time.Second)
		option.EndTime = end.Format(utils.TimeFormatDatetime)
	}
	condition := &model.DeviceInfoDaily{
		UserID: userId,
	}

	list, err := dao.GetNodesInfoDailyList(context.Background(), condition, option)
	if err != nil {
		log.Errorf("get incoming daily: %v", err)
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
		end, _ := time.Parse(utils.TimeFormatDateOnly, endTime)
		end = end.Add(1 * time.Hour).Add(-time.Second)
		option.EndTime = end.Format(utils.TimeFormatDatetime)
	}

	condition := &model.DeviceInfoHour{
		DeviceID: deviceID,
	}
	list, err := dao.GetDeviceInfoDailyHourList(context.Background(), condition, option)
	if err != nil {
		log.Errorf("get incoming hour daily: %v", err)
		return nil
	}

	return list
}

func GetQueryInfoHandler(c *gin.Context) {
	info := &model.DeviceInfo{}
	keyValue := c.Query("key")
	info.UserID = keyValue
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
		DetailList := dao.GetDeviceInfoById(context.Background(), keyValue)
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
		log.Errorf("get device info list: %v", err)
		c.JSON(http.StatusOK, respError(errors.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  list,
		"total": total,
	}))
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
		log.Errorf("get device info list: %v", err)
		c.JSON(http.StatusOK, respError(errors.ErrInternalServer))
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
	//nodeType, _ := strconv.ParseInt(c.Query("node_type"), 10, 64)
	//info.NodeType = int32(nodeType)
	option := dao.QueryOption{
		Page:       page,
		PageSize:   pageSize,
		Order:      order,
		OrderField: orderField,
	}
	list, total, err := dao.GetDeviceInfoList(c.Request.Context(), info, option)
	if err != nil {
		log.Errorf("get device info list: %v", err)
		c.JSON(http.StatusOK, respError(errors.ErrInternalServer))
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
		log.Errorf("get nodes info list: %v", err)
		c.JSON(http.StatusOK, respError(errors.ErrInternalServer))
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
		log.Errorf("get device info list: %v", err)
		c.JSON(http.StatusOK, respError(errors.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  dao.HandleMapInfo(list),
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
		c.JSON(http.StatusOK, respError(errors.ErrInternalServer))
		return
	}
	deviceInfo, err := dao.GetDeviceInfoByID(c.Request.Context(), deviceID)
	if err != nil {
		c.JSON(http.StatusOK, respError(errors.ErrInternalServer))
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
		c.JSON(http.StatusOK, respError(errors.ErrInternalServer))
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
