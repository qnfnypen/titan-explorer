package api

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/errors"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"net/http"
	"strconv"
	"time"
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

	list, total, err := dao.GetDeviceInfoList(c.Request.Context(), info, option)
	if err != nil {
		log.Errorf("get device info list: %v", err)
		c.JSON(http.StatusBadRequest, respError(errors.ErrNotFound))
		return
	}
	var dataList []*model.DeviceInfo
	var res DeviceInfoPage
	var dataRes IndexUserDeviceRes
	for _, data := range list {
		err = getProfitByDeviceID(data, &res)
		dataRes.CumulativeProfit += data.CumuProfit
		dataRes.TodayProfit += data.TodayProfit
		dataRes.SevenDaysProfit += data.SevenDaysProfit
		dataRes.YesterdayProfit += data.YesterdayProfit
		dataRes.MonthProfit += data.MonthProfit
		if err != nil {
			log.Error("getProfitByDeviceID：", data.DeviceID)
		}
		dataList = append(dataList, data)
	}

	// Devices
	dataRes.AbnormalNum = res.Abnormal
	dataRes.OfflineNum = res.Offline
	dataRes.OnlineNum = res.Online
	dataRes.TotalNum = total
	dataRes.TotalBandwidth = res.BandwidthMb
	// Profit
	p := &model.IncomeDaily{}
	p.UserID = info.UserID
	m := dao.GetIncomeAllList(c.Request.Context(), p, option)
	dataRes.DailyIncome = m

	c.JSON(http.StatusOK, respJSON(dataRes))
}

func timeFormat(p IncomeDailySearch) (m map[string]interface{}) {
	timeNow := time.Now().Format("2006-01-02")

	dd, _ := time.ParseDuration("-24h")
	FromTime := time.Now().Add(dd * 14).Format("2006-01-02")
	if p.DateFrom == "" && p.Date == "" {
		p.DateFrom = FromTime
	}
	if p.DateTo == "" && p.Date == "" {
		p.DateTo = timeNow
	}
	p.DateFrom = p.DateFrom + " 00:00:00"
	p.DateTo = p.DateTo + " 23:59:59"

	option := dao.QueryOption{
		StartTime: p.DateFrom,
		EndTime:   p.DateTo,
	}

	condition := &model.IncomeDaily{
		DeviceID: p.DeviceID,
	}

	list, _, err := dao.GetIncomeDailyList(context.Background(), condition, option)
	if err != nil {
		log.Errorf("get incoming daily: %v", err)
		return
	}

	return getDaysData(list)
}

func timeFormatHour(p IncomeDailySearch) (m map[string]interface{}) {
	timeNow := time.Now().Format("2006-01-02")

	dd, _ := time.ParseDuration("-24h")
	FromTime := time.Now().Add(dd * 14).Format("2006-01-02")
	if p.DateFrom == "" && p.Date == "" {
		p.DateFrom = FromTime
	}
	if p.DateTo == "" && p.Date == "" {
		p.DateTo = timeNow
	}
	if p.Date == "" {
		p.Date = time.Now().Format("2006-01-02")
	}
	p.DateFrom = p.Date + " 00:00:00"
	p.DateTo = p.Date + " 23:59:59"

	option := dao.QueryOption{
		StartTime: p.DateFrom,
		EndTime:   p.DateTo,
	}

	condition := &model.HourDaily{
		DeviceID: p.DeviceID,
		UserID:   p.UserID,
	}

	list, _, err := dao.GetIncomeDailyHourList(context.Background(), condition, option)
	if err != nil {
		log.Errorf("get incoming hour daily: %v", err)
		return
	}

	return getDaysDataHour(list)
}

func getDaysDataHour(list []*model.HourDaily) (returnMapList map[string]interface{}) {
	returnMap := make(map[string]interface{})
	queryMapTo := make(map[string]float64)
	pkgLossRatioTo := make(map[string]float64)
	latencyTo := make(map[string]float64)
	onlineJsonDailyTo := make(map[string]float64)
	natTypeTo := make(map[string]float64)
	diskUsageTo := make(map[string]float64)
	incomeHourBefore := float64(0)
	onlineHourBefore := float64(0)
	firstData := true
	for _, v := range list {
		timeStr := v.Time.Format(TimeFormatHM)
		if firstData {
			incomeHourBefore = v.HourIncome
			onlineHourBefore = v.OnlineTime
			firstData = false
			continue
		}
		timeMinStr := v.Time.Format(TimeFormatM)
		if timeMinStr == "00" {
			queryMapTo[timeStr] = v.HourIncome - incomeHourBefore
			incomeHourBefore = v.HourIncome
			onlineJsonDailyTo[timeStr] = v.OnlineTime - onlineHourBefore
			onlineHourBefore = v.OnlineTime
		}
		if timeMinStr == "00" || timeMinStr == "30" {
			pkgLossRatioTo[timeStr] = v.PkgLossRatio * 100
			latencyTo[timeStr] = v.Latency
			natTypeTo[timeStr] = v.NatRatio
			diskUsageTo[timeStr] = v.DiskUsage
		}
	}
	returnMap["income"] = queryMapTo
	returnMap["online"] = onlineJsonDailyTo
	returnMap["pkgLoss"] = pkgLossRatioTo
	returnMap["latency"] = latencyTo
	returnMap["natType"] = natTypeTo
	returnMap["diskUsage"] = diskUsageTo
	returnMapList = returnMap
	return
}

func getDaysData(list []*model.IncomeDaily) (returnMapList map[string]interface{}) {
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
	returnMap["pkgLoss"] = pkgLossRatioTo
	returnMap["latency"] = latencyTo
	returnMap["natType"] = natTypeTo
	returnMap["diskUsage"] = diskUsageTo
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
	var res DeviceInfoPage
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
	var dataList []*model.DeviceInfo
	for _, data := range list {
		err = getProfitByDeviceID(data, &res)
		if err != nil {
			log.Errorf("get profit by device id: %v", data.DeviceID)
			c.JSON(http.StatusBadRequest, respError(errors.ErrInternalServer))
			return
		}
		dataList = append(dataList, data)
	}
	res.List = dataList
	res.Count = total
	res.AllDevices = total
	c.JSON(http.StatusOK, respJSON(res))
}

func GetDeviceDiagnosisDailyHandler(c *gin.Context) {
	var p IncomeDailySearch
	from := c.Query("from")
	to := c.Query("to")
	p.DateFrom = from
	p.DateTo = to
	p.DeviceID = c.Query("device_id")
	var res IncomeDailyRes
	m := timeFormat(p)
	res.DailyIncome = m
	res.DeviceDiagnosis = "good"
	c.JSON(http.StatusOK, respJSON(res))
}

func GetDeviceDiagnosisHourHandler(c *gin.Context) {
	var p IncomeDailySearch
	p.DeviceID = c.Query("device_id")
	p.Date = c.Query("date")
	p.UserID = c.Query("userId")
	var res IncomeDailyRes
	m := timeFormatHour(p)

	res.DailyIncome = m
	res.DeviceDiagnosis = "good"
	c.JSON(http.StatusOK, respJSON(res))
}

func getProfitByDeviceID(rt *model.DeviceInfo, dt *DeviceInfoPage) error {
	switch rt.DeviceStatus {
	case "online":
		dt.Online += 1
	case "offline":
		dt.Offline += 1
	case "abnormal":
		dt.Abnormal += 1
	}
	dt.BandwidthMb += rt.BandwidthUp
	return nil
}
