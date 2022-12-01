package statistics

import (
	"context"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"strconv"
	"time"
)

const (
	TimeFormatYMD = "2006-01-02"
)

var (
	DeviceIDAndUserId map[string]string
)

func (s *Statistic) FetchIncomeDaily() error {
	log.Info("start fetch income daily")
	start := time.Now()
	defer func() {
		log.Infof("fetch income daily done, cost: %v", time.Since(start))
	}()

	var deviceInfos []*model.DeviceInfo
	opt := dao.QueryOption{
		Page:     1,
		PageSize: 100,
	}
loop:
	devices, total, err := dao.GetDeviceInfoList(context.Background(), &model.DeviceInfo{}, opt)
	if err != nil {
		return err
	}

	for _, device := range devices {
		deviceInfos = append(deviceInfos, device)
	}

	opt.Page++

	if int64(len(deviceInfos)) < total {
		goto loop
	}

	for _, device := range deviceInfos {
		var hourDaily model.HourDaily
		hourDaily.Time = start
		hourDaily.DiskUsage = device.DiskUsage
		hourDaily.DeviceID = device.DeviceID
		hourDaily.PkgLossRatio = device.PkgLossRatio
		hourDaily.HourIncome = device.TodayProfit
		hourDaily.OnlineTime = device.OnlineTime
		hourDaily.Latency = device.Latency
		hourDaily.DiskUsage = device.DiskUsage
		_, ok := DeviceIDAndUserId[hourDaily.DeviceID]
		if ok {
			hourDaily.UserID = DeviceIDAndUserId[hourDaily.DeviceID]
		}
		err = TransferData(hourDaily)
		if err != nil {
			log.Error(err.Error())
		}

		timeNow := time.Now().Format("2006-01-02")
		DateFrom := timeNow + " 00:00:00"
		DateTo := timeNow + " 23:59:59"
		sqlClause := fmt.Sprintf("select user_id,date_format(time, '%%Y-%%m-%%d') as date, avg(nat_ratio) as nat_ratio, avg(disk_usage) as disk_usage, avg(latency) as latency, avg(pkg_loss_ratio) as pkg_loss_ratio, max(hour_income) as hour_income,max(online_time) as online_time_max,min(online_time) as online_time_min from hour_daily "+
			"where device_id='%s' and time>='%s' and time<='%s' group by date", device.DeviceID, DateFrom, DateTo)
		datas, err := dao.GetQueryDataList(sqlClause)
		if err != nil {
			log.Error(err.Error())
			return err
		}
		for _, data := range datas {
			var InPage model.IncomeDaily
			InPage.Time, _ = time.Parse(TimeFormatYMD, data["date"])
			InPage.DiskUsage = Str2Float64(data["disk_usage"])
			InPage.NatRatio = Str2Float64(data["nat_ratio"])
			InPage.Income = Str2Float64(data["hour_income"])
			InPage.OnlineTime = Str2Float64(data["online_time_max"]) - Str2Float64(data["online_time_min"])
			InPage.PkgLossRatio = Str2Float64(data["pkg_loss_ratio"])
			InPage.Latency = Str2Float64(data["latency"])
			InPage.DeviceID = device.DeviceID
			InPage.UserID = data["user_id"]
			err = SavaIncomeDailyInfo(InPage)
			if err != nil {
				log.Errorf("save daily info: %v", err)
			}
		}
	}

	return nil
}

func TransferData(hourDaily model.HourDaily) error {
	if hourDaily.DeviceID == "" {
		return nil
	}

	hourDaily.UpdatedAt = time.Now()
	ctx := context.Background()
	old, err := dao.GetHourDailyByTime(ctx, hourDaily.DeviceID, hourDaily.Time)
	if err != nil {
		log.Errorf("get hour daily by time: %v", err)
		return err
	}

	if old == nil {
		hourDaily.CreatedAt = time.Now()
		return dao.CreateHourDaily(ctx, &hourDaily)
	}

	hourDaily.ID = old.ID
	return dao.UpdateHourDaily(ctx, &hourDaily)
}

func SavaIncomeDailyInfo(daily model.IncomeDaily) error {
	if daily.DeviceID == "" {
		return nil
	}

	daily.UpdatedAt = time.Now()
	_, ok := DeviceIDAndUserId[daily.DeviceID]
	if ok {
		daily.UserID = DeviceIDAndUserId[daily.DeviceID]
	}

	ctx := context.Background()
	old, err := dao.GetIncomeDailyByTime(ctx, daily.DeviceID, daily.Time)
	if err != nil {
		log.Errorf("get hour daily by time: %v", err)
		return err
	}

	if old == nil {
		daily.CreatedAt = time.Now()
		return dao.CreateIncomeDaily(ctx, &daily)
	}

	daily.ID = old.ID
	return dao.UpdateIncomeDaily(ctx, &daily)
}

func Str2Float64(s string) float64 {
	ret, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Error(err.Error())
		return 0.00
	}
	return ret
}

func QueryDataByDate(DeviceID, DateFrom, DateTo string) map[string]string {

	sqlClause := fmt.Sprintf("select sum(income) as income,online_time from income_daily "+
		"where  time>='%s' and time<='%s' and device_id='%s' group by user_id;", DateFrom, DateTo, DeviceID)
	if DateFrom == "" {
		sqlClause = fmt.Sprintf("select sum(income) as income,online_time from income_daily "+
			"where device_id='%s' group by user_id;", DeviceID)
	}
	data, err := dao.GetQueryDataList(sqlClause)
	if err != nil {
		log.Error(err.Error())
		return nil
	}
	if len(data) > 0 {
		return data[0]
	}
	return nil
}

func (s *Statistic) FetchYesTodayIncome() error {
	log.Info("start fetch yesterday income")
	start := time.Now()
	defer func() {
		log.Infof("fetch yesterday income done, cost: %v", time.Since(start))
	}()

	var deviceInfos []*model.DeviceInfo
	opt := dao.QueryOption{
		Page:     1,
		PageSize: 100,
	}
loop:
	devices, total, err := dao.GetDeviceInfoList(context.Background(), &model.DeviceInfo{}, opt)
	if err != nil {
		return err
	}

	for _, device := range devices {
		deviceInfos = append(deviceInfos, device)
	}

	opt.Page++

	if int64(len(deviceInfos)) < total {
		goto loop
	}

	for _, device := range deviceInfos {
		dd, _ := time.ParseDuration("-24h")
		timeBase := time.Now().Add(dd * 1).Format("2006-01-02")
		timeNow := time.Now().Format("2006-01-02")
		DateFrom := timeBase + " 00:00:00"
		DateTo := timeBase + " 23:59:59"
		dataY := QueryDataByDate(device.DeviceID, DateFrom, DateTo)
		timeBase = time.Now().Add(dd * 6).Format("2006-01-02")
		DateFrom = timeBase + " 00:00:00"
		DateTo = timeNow + " 23:59:59"
		dataS := QueryDataByDate(device.DeviceID, DateFrom, DateTo)
		timeBase = time.Now().Add(dd * 29).Format("2006-01-02")
		DateFrom = timeBase + " 00:00:00"
		dataM := QueryDataByDate(device.DeviceID, DateFrom, DateTo)
		dataA := QueryDataByDate(device.DeviceID, "", "")
		DateFrom = timeNow + " 00:00:00"
		DateTo = timeNow + " 23:59:59"
		dataT := QueryDataByDate(device.DeviceID, DateFrom, DateTo)
		var dataUpdate model.DeviceInfo
		dataUpdate.YesterdayProfit = 0
		dataUpdate.SevenDaysProfit = 0
		dataUpdate.MonthProfit = 0
		dataUpdate.CumuProfit = 0
		dataUpdate.TodayOnlineTime = 0
		dataUpdate.TodayProfit = 0
		if len(dataY) > 0 {
			dataUpdate.YesterdayProfit = Str2Float64(dataY["income"])
		}
		if len(dataS) > 0 {
			dataUpdate.SevenDaysProfit = Str2Float64(dataS["income"])
		}
		if len(dataM) > 0 {
			dataUpdate.MonthProfit = Str2Float64(dataM["income"])
		}
		if len(dataA) > 0 {
			dataUpdate.CumuProfit = Str2Float64(dataA["income"])
		}
		if len(dataT) > 0 {
			dataUpdate.TodayProfit = Str2Float64(dataT["income"])
			dataUpdate.TodayOnlineTime = Str2Float64(dataT["online_time"])
		}
		dataUpdate.UpdatedAt = time.Now()
		_, ok := DeviceIDAndUserId[device.DeviceID]
		if ok {
			dataUpdate.UserID = DeviceIDAndUserId[device.DeviceID]
		}
		//err := dao.DB.Save(&data).Error

		ctx := context.Background()
		old, err := dao.GetDeviceInfoByID(ctx, device.DeviceID)
		if err != nil {
			log.Errorf("get device info by id: %v", err)
			return err
		}

		if old == nil {
			dataUpdate.CreatedAt = time.Now()
			return dao.AddDeviceInfo(ctx, &dataUpdate)
		}
		old.YesterdayProfit = dataUpdate.YesterdayProfit
		old.SevenDaysProfit = dataUpdate.SevenDaysProfit
		old.MonthProfit = dataUpdate.MonthProfit
		old.CumuProfit = dataUpdate.CumuProfit
		old.UpdatedAt = dataUpdate.UpdatedAt
		old.TodayOnlineTime = dataUpdate.TodayOnlineTime
		old.TodayProfit = dataUpdate.TodayProfit
		if dataUpdate.UserID != "" {
			old.UserID = dataUpdate.UserID
		}
		err = dao.UpdateDeviceInfo(ctx, old)
		if err != nil {
			log.Errorf("update device info: %v", err)
		}
	}
	return nil
}
