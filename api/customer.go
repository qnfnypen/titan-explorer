package api

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/errors"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/golang-module/carbon/v2"
	"net/http"
	"strconv"
)

const (
	TimeFormatYMDHMS = "2006-01-02 15:04:05"
	TimeFormatYMD    = "2006-01-02"
	TimeFormatMD     = "01-02"
	TimeFormatHM     = "15:04"
	TimeFormatM      = "04"
)

var AllM AllMinerInfo

func GetAllMinerInfoHandler(c *gin.Context) {
	c.JSON(http.StatusOK, respJSON(AllM))
}

func GetIndexInfoHandler(c *gin.Context) {
	fullNodeInfo, err := dao.GetFullNodeInfo(c.Request.Context())
	if err != nil {
		log.Errorf("get full node info: %v", err)
		c.JSON(http.StatusBadRequest, respError(errors.ErrInternalServer))
		return
	}
	c.JSON(http.StatusOK, respJSON(fullNodeInfo))
}

// GetUserDeviceInfoHandler devices overview
func GetUserDeviceInfoHandler(c *gin.Context) {
	info := &model.DeviceInfo{}
	info.UserID = c.Query("userId")
	DeviceID := c.Query("device_id")
	info.DeviceID = DeviceID
	info.DeviceStatus = c.Query("device_status")
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))
	option := dao.QueryOption{
		Page:      page,
		PageSize:  pageSize,
		StartTime: c.Query("from"),
		EndTime:   c.Query("to"),
	}

	list, _, err := dao.GetDeviceInfoList(c.Request.Context(), info, option)
	if err != nil {
		log.Errorf("get device info list: %v", err)
		c.JSON(http.StatusBadRequest, respError(errors.ErrNotFound))
		return
	}
	var dataList []*model.DeviceInfo
	var dataRes IndexUserDeviceRes
	for _, data := range list {
		dataRes.CumulativeProfit += data.CumulativeProfit
		dataRes.TodayProfit += data.TodayProfit
		dataRes.SevenDaysProfit += data.SevenDaysProfit
		dataRes.YesterdayProfit += data.YesterdayProfit
		dataRes.MonthProfit += data.MonthProfit
		if err != nil {
			log.Error("getProfitByDeviceID：", data.DeviceID)
		}
		dataList = append(dataList, data)
	}

	// Profit
	p := &model.DeviceInfoDaily{}
	p.UserID = info.UserID
	m := dao.GetIncomeAllList(c.Request.Context(), p, option)
	dataRes.DailyIncome = m

	c.JSON(http.StatusOK, respJSON(dataRes))
}

func timeFormat(deviceID, startTime, endTime string) (m map[string]interface{}) {
	option := dao.QueryOption{
		StartTime: startTime,
		EndTime:   endTime,
	}
	if startTime == "" {
		option.StartTime = carbon.Now().SubDays(14).StartOfDay().String()
	}
	if endTime == "" {
		option.EndTime = carbon.Now().EndOfDay().String()
	}

	condition := &model.DeviceInfoDaily{
		DeviceID: deviceID,
	}

	list, err := dao.GetDeviceInfoDailyList(context.Background(), condition, option)
	if err != nil {
		log.Errorf("get incoming daily: %v", err)
		return
	}

	return getDaysData(list)
}

func timeFormatHour(deviceID, date string) (m map[string]interface{}) {
	option := dao.QueryOption{}
	if date != "" {
		option.StartTime = carbon.Parse(date).StartOfDay().String()
		option.EndTime = carbon.Parse(date).EndOfDay().String()
	} else {
		option.StartTime = carbon.Now().StartOfDay().String()
		option.EndTime = carbon.Now().EndOfDay().String()
	}

	condition := &model.DeviceInfoHour{
		DeviceID: deviceID,
	}

	list, err := dao.GetDeviceInfoDailyHourList(context.Background(), condition, option)
	if err != nil {
		log.Errorf("get incoming hour daily: %v", err)
		return
	}

	return getDaysDataHour(list)
}

func getDaysDataHour(list []*model.DeviceInfoHour) (returnMapList map[string]interface{}) {
	if len(list) == 0 {
		return
	}
	returnMap := make(map[string]interface{})
	queryMapTo := make(map[string]float64)
	pkgLossRatioTo := make(map[string]float64)
	latencyTo := make(map[string]float64)
	onlineJsonDailyTo := make(map[string]float64)
	natTypeTo := make(map[string]float64)
	diskUsageTo := make(map[string]float64)
	incomeHourBefore := list[0].HourIncome
	onlineHourBefore := list[0].OnlineTime
	for _, v := range list {
		timeStr := v.Time.Format(TimeFormatHM)
		minute := v.Time.Minute()
		if minute == 0 {
			queryMapTo[timeStr] = v.HourIncome - incomeHourBefore
			incomeHourBefore = v.HourIncome
			onlineJsonDailyTo[timeStr] = v.OnlineTime - onlineHourBefore
			onlineHourBefore = v.OnlineTime
		}
		if minute == 0 || minute == 30 {
			pkgLossRatioTo[timeStr] = v.PkgLossRatio * 100
			latencyTo[timeStr] = v.Latency
			natTypeTo[timeStr] = v.NatRatio
			diskUsageTo[timeStr] = v.DiskUsage
		}
	}
	returnMap["income"] = queryMapTo
	returnMap["online"] = onlineJsonDailyTo
	returnMap["pkg_loss"] = pkgLossRatioTo
	returnMap["latency"] = latencyTo
	returnMap["nat_type"] = natTypeTo
	returnMap["disk_usage"] = diskUsageTo
	// TODO:
	returnMap["traffic"] = latencyTo
	returnMap["retrieval"] = latencyTo
	returnMapList = returnMap
	return
}

func getDaysData(list []*model.DeviceInfoDaily) (returnMapList map[string]interface{}) {
	returnMap := make(map[string]interface{})
	queryMapTo := make(map[string]float64)
	pkgLossRatioTo := make(map[string]float64)
	latencyTo := make(map[string]float64)
	onlineJsonDailyTo := make(map[string]float64)
	natTypeTo := make(map[string]float64)
	diskUsageTo := make(map[string]float64)
	for _, v := range list {
		timeStr := v.Time.Format(TimeFormatMD)
		queryMapTo[timeStr] += v.Income
		pkgLossRatioTo[timeStr] = v.PkgLossRatio
		latencyTo[timeStr] = v.Latency
		onlineJsonDailyTo[timeStr] = v.OnlineTime
		natTypeTo[timeStr] = v.NatRatio
		diskUsageTo[timeStr] = v.DiskUsage
	}
	returnMap["income"] = queryMapTo
	returnMap["online"] = onlineJsonDailyTo
	returnMap["pkg_loss"] = pkgLossRatioTo
	returnMap["latency"] = latencyTo
	returnMap["nat_type"] = natTypeTo
	returnMap["disk_usage"] = diskUsageTo
	returnMapList = returnMap
	return
}

func RetrievalHandler(c *gin.Context) {
	taskInfo := &model.TaskInfo{}
	taskInfo.UserID = c.Query("userId")
	taskInfo.Status = c.Query("status")
	taskInfo.Cid = c.Query("cid")
	var res RetrievalPageRes
	list, total, err := dao.GetTaskInfoList(c.Request.Context(), taskInfo, dao.QueryOption{})
	if err != nil {
		log.Error(err.Error())
		c.JSON(http.StatusBadRequest, respError(errors.ErrInvalidParams))
		return
	}
	res.List = list
	res.Count = total
	res.StorageT = AllM.StorageT
	res.BandwidthMb = AllM.BandwidthMb
	// AllMinerNum MinerInfo
	res.AllCandidate = AllM.AllCandidate
	res.AllEdgeNode = AllM.AllEdgeNode
	res.AllVerifier = AllM.AllVerifier

	c.JSON(http.StatusOK, respJSON(res))
}

func GetDeviceInfoHandler(c *gin.Context) {
	info := &model.DeviceInfo{}
	info.UserID = c.Query("userId")
	info.DeviceID = c.Query("device_id")
	info.DeviceStatus = c.Query("device_status")
	pageSize, _ := strconv.Atoi("page_size")
	page, _ := strconv.Atoi("page")
	option := dao.QueryOption{
		Page:     page,
		PageSize: pageSize,
	}
	list, total, err := dao.GetDeviceInfoList(c.Request.Context(), info, option)
	if err != nil {
		log.Errorf("get device info list: %v", err)
		c.JSON(http.StatusBadRequest, respError(errors.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  list,
		"total": total,
	}))
}

func GetDeviceDiagnosisDailyHandler(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	deviceID := c.Query("device_id")
	m := timeFormat(deviceID, from, to)
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"series_data": m,
	}))
}

func GetDeviceDiagnosisHourHandler(c *gin.Context) {
	deviceID := c.Query("device_id")
	date := c.Query("date")
	m := timeFormatHour(deviceID, date)
	deviceInfo, err := dao.GetDeviceInfoByID(c.Request.Context(), deviceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, respError(errors.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"series_data":  m,
		"cpu_usage":    deviceInfo.CpuUsage,
		"cpu_cores":    deviceInfo.CpuCores,
		"memory":       deviceInfo.Memory,
		"memory_usage": deviceInfo.MemoryUsage * deviceInfo.Memory,
		"disk_usage":   deviceInfo.DiskUsage,
		"disk_type":    deviceInfo.DiskType,
		"file_system":  deviceInfo.IoSystem,
	}))
}
