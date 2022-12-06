package statistics

import (
	"context"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/golang-module/carbon/v2"
	"strconv"
	"time"
)

const (
	TimeFormatYMD = "2006-01-02"
)

var (
	DeviceIDAndUserId map[string]string
)

func addDeviceInfoHours(ctx context.Context, deviceInfo []*model.DeviceInfo) error {
	log.Info("start fetch device info hours")
	start := time.Now()
	defer func() {
		log.Infof("fetch device info hours done, cost: %v", time.Since(start))
	}()

	for _, device := range deviceInfo {
		var deviceInfoHour model.DeviceInfoHour
		deviceInfoHour.Time = start
		deviceInfoHour.DiskUsage = device.DiskUsage
		deviceInfoHour.DeviceID = device.DeviceID
		deviceInfoHour.PkgLossRatio = device.PkgLossRatio
		deviceInfoHour.HourIncome = device.CumulativeProfit
		deviceInfoHour.OnlineTime = device.OnlineTime
		deviceInfoHour.Latency = device.Latency
		deviceInfoHour.DiskUsage = device.DiskUsage
		_, ok := DeviceIDAndUserId[deviceInfoHour.DeviceID]
		if ok {
			deviceInfoHour.UserID = DeviceIDAndUserId[deviceInfoHour.DeviceID]
		}
		err := TransferData(deviceInfoHour)
		if err != nil {
			log.Error(err.Error())
		}

		//timeNow := time.Now().Format("2006-01-02")
		//DateFrom := timeNow + " 00:00:00"
		//DateTo := timeNow + " 23:59:59"
		//sqlClause := fmt.Sprintf("select user_id,date_format(time, '%%Y-%%m-%%d') as date, avg(nat_ratio) as nat_ratio, avg(disk_usage) as disk_usage, avg(latency) as latency, avg(pkg_loss_ratio) as pkg_loss_ratio, max(hour_income) as hour_income_max, min(hour_income) as hour_income_min ,max(online_time) as online_time_max,min(online_time) as online_time_min from device_info_hour "+
		//	"where device_id='%s' and time>='%s' and time<='%s' group by date", device.DeviceID, DateFrom, DateTo)
		//datas, err := dao.GetQueryDataList(sqlClause)
		//if err != nil {
		//	log.Error(err.Error())
		//	return err
		//}
		//for _, data := range datas {
		//	var InPage model.DeviceInfoDaily
		//	InPage.Time, _ = time.Parse(TimeFormatYMD, data["date"])
		//	InPage.DiskUsage = Str2Float64(data["disk_usage"])
		//	InPage.NatRatio = Str2Float64(data["nat_ratio"])
		//	InPage.Income = Str2Float64(data["hour_income_max"]) - Str2Float64(data["hour_income_min"])
		//	InPage.OnlineTime = Str2Float64(data["online_time_max"]) - Str2Float64(data["online_time_min"])
		//	InPage.PkgLossRatio = Str2Float64(data["pkg_loss_ratio"])
		//	InPage.Latency = Str2Float64(data["latency"])
		//	InPage.DeviceID = device.DeviceID
		//	InPage.UserID = data["user_id"]
		//	err = SavaDeviceInfoDailyInfo(InPage)
		//	if err != nil {
		//		log.Errorf("save daily info: %v", err)
		//	}
		//}
	}

	return nil
}

func TransferData(deviceInfoHour model.DeviceInfoHour) error {
	if deviceInfoHour.DeviceID == "" {
		return nil
	}

	deviceInfoHour.UpdatedAt = time.Now()
	ctx := context.Background()
	old, err := dao.GetDeviceInfoHourByTime(ctx, deviceInfoHour.DeviceID, deviceInfoHour.Time)
	if err != nil {
		log.Errorf("get hour daily by time: %v", err)
		return err
	}

	if old == nil {
		deviceInfoHour.CreatedAt = time.Now()
		return dao.CreateDeviceInfoHour(ctx, &deviceInfoHour)
	}

	deviceInfoHour.ID = old.ID
	return dao.UpdateDeviceInfoHour(ctx, &deviceInfoHour)
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

	sqlClause := fmt.Sprintf("select sum(income) as income,online_time from device_info_daily "+
		"where  time>='%s' and time<='%s' and device_id='%s' group by user_id;", DateFrom, DateTo, DeviceID)
	if DateFrom == "" {
		sqlClause = fmt.Sprintf("select sum(income) as income,online_time from device_info_daily "+
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

func (s *Statistic) SumDeviceInfoDaily() error {
	log.Info("start sum device info daily")
	start := time.Now()
	defer func() {
		log.Infof("sum device info daily done, cost: %v", time.Since(start))
	}()

	startOfTodayTime := carbon.Now().StartOfDay().String()
	endOfTodayTime := carbon.Now().EndOfDay().String()
	sqlClause := fmt.Sprintf("select user_id, device_id, date_format(time, '%%Y-%%m-%%d') as date, avg(nat_ratio) as nat_ratio, avg(disk_usage) as disk_usage, avg(latency) as latency, avg(pkg_loss_ratio) as pkg_loss_ratio, max(hour_income) as hour_income_max, min(hour_income) as hour_income_min ,max(online_time) as online_time_max,min(online_time) as online_time_min from device_info_hour "+
		"where time>='%s' and time<='%s' group by date, device_id", startOfTodayTime, endOfTodayTime)
	datas, err := dao.GetQueryDataList(sqlClause)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	var dailyInfos []*model.DeviceInfoDaily
	for _, data := range datas {
		var daily model.DeviceInfoDaily
		daily.Time, _ = time.Parse(TimeFormatYMD, data["date"])
		daily.DiskUsage = Str2Float64(data["disk_usage"])
		daily.NatRatio = Str2Float64(data["nat_ratio"])
		daily.Income = Str2Float64(data["hour_income_max"]) - Str2Float64(data["hour_income_min"])
		daily.OnlineTime = Str2Float64(data["online_time_max"]) - Str2Float64(data["online_time_min"])
		daily.PkgLossRatio = Str2Float64(data["pkg_loss_ratio"])
		daily.Latency = Str2Float64(data["latency"])
		daily.DeviceID = data["device_id"]
		daily.UserID = data["user_id"]
		daily.CreatedAt = time.Now()
		daily.UpdatedAt = time.Now()
		dailyInfos = append(dailyInfos, &daily)
	}

	err = dao.BulkUpsertDeviceInfoDaily(context.Background(), dailyInfos)
	if err != nil {
		log.Errorf("upsert device info daily: %v", err)
		return err
	}

	return nil
}

func (s *Statistic) SumDeviceInfoWeeklyMonthly() error {
	log.Info("start sum device info weekly and monthly")
	start := time.Now()
	defer func() {
		log.Infof("sum device info weekly and monthly done, cost: %v", time.Since(start))
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

	deviceInfos = append(deviceInfos, devices...)
	opt.Page++
	if int64(len(deviceInfos)) < total {
		goto loop
	}

	for i := 0; i < len(deviceInfos); i++ {
		startOfTodayTime := carbon.Now().StartOfDay().String()
		endOfTodayTime := carbon.Now().EndOfDay().String()

		startOfYesterday := carbon.Yesterday().StartOfDay().String()
		endOfYesterday := carbon.Yesterday().EndOfDay().String()
		dataY := QueryDataByDate(deviceInfos[i].DeviceID, startOfYesterday, endOfYesterday)

		startOfWeekTime := carbon.Now().SubDays(6).StartOfDay().String()
		dataS := QueryDataByDate(deviceInfos[i].DeviceID, startOfWeekTime, endOfTodayTime)

		startOfMonthTime := carbon.Now().SubDays(29).StartOfDay().String()
		dataM := QueryDataByDate(deviceInfos[i].DeviceID, startOfMonthTime, endOfTodayTime)
		dataA := QueryDataByDate(deviceInfos[i].DeviceID, "", "")

		dataT := QueryDataByDate(deviceInfos[i].DeviceID, startOfTodayTime, endOfTodayTime)
		if len(dataY) > 0 {
			deviceInfos[i].YesterdayProfit = Str2Float64(dataY["income"])
		}
		if len(dataS) > 0 {
			deviceInfos[i].SevenDaysProfit = Str2Float64(dataS["income"])
		}
		if len(dataM) > 0 {
			deviceInfos[i].MonthProfit = Str2Float64(dataM["income"])
		}
		if len(dataA) > 0 {
			deviceInfos[i].CumulativeProfit = Str2Float64(dataA["income"])
		}
		if len(dataT) > 0 {
			deviceInfos[i].TodayProfit = Str2Float64(dataT["income"])
			deviceInfos[i].TodayOnlineTime = Str2Float64(dataT["online_time"])
		}

		deviceInfos[i].UpdatedAt = time.Now()
		_, ok := DeviceIDAndUserId[deviceInfos[i].DeviceID]
		if ok {
			deviceInfos[i].UserID = DeviceIDAndUserId[deviceInfos[i].DeviceID]
		}
	}

	if err = dao.BulkUpdateDeviceInfo(context.Background(), deviceInfos); err != nil {
		log.Errorf("bulk update device: %v", err)
	}

	return nil
}
