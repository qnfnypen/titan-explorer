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
		deviceInfoHour.HourIncome = device.TodayProfit
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

		timeNow := time.Now().Format("2006-01-02")
		DateFrom := timeNow + " 00:00:00"
		DateTo := timeNow + " 23:59:59"
		sqlClause := fmt.Sprintf("select user_id,date_format(time, '%%Y-%%m-%%d') as date, avg(nat_ratio) as nat_ratio, avg(disk_usage) as disk_usage, avg(latency) as latency, avg(pkg_loss_ratio) as pkg_loss_ratio, max(hour_income) as hour_income,max(online_time) as online_time_max,min(online_time) as online_time_min from device_info_hour "+
			"where device_id='%s' and time>='%s' and time<='%s' group by date", device.DeviceID, DateFrom, DateTo)
		datas, err := dao.GetQueryDataList(sqlClause)
		if err != nil {
			log.Error(err.Error())
			return err
		}
		for _, data := range datas {
			var InPage model.DeviceInfoDaily
			InPage.Time, _ = time.Parse(TimeFormatYMD, data["date"])
			InPage.DiskUsage = Str2Float64(data["disk_usage"])
			InPage.NatRatio = Str2Float64(data["nat_ratio"])
			InPage.Income = Str2Float64(data["hour_income"])
			InPage.OnlineTime = Str2Float64(data["online_time_max"]) - Str2Float64(data["online_time_min"])
			InPage.PkgLossRatio = Str2Float64(data["pkg_loss_ratio"])
			InPage.Latency = Str2Float64(data["latency"])
			InPage.DeviceID = device.DeviceID
			InPage.UserID = data["user_id"]
			err = SavaDeviceInfoDailyInfo(InPage)
			if err != nil {
				log.Errorf("save daily info: %v", err)
			}
		}
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

func SavaDeviceInfoDailyInfo(daily model.DeviceInfoDaily) error {
	if daily.DeviceID == "" {
		return nil
	}

	daily.UpdatedAt = time.Now()
	_, ok := DeviceIDAndUserId[daily.DeviceID]
	if ok {
		daily.UserID = DeviceIDAndUserId[daily.DeviceID]
	}

	ctx := context.Background()
	old, err := dao.GetDeviceInfoDailyByTime(ctx, daily.DeviceID, daily.Time)
	if err != nil {
		log.Errorf("get hour daily by time: %v", err)
		return err
	}

	if old == nil {
		daily.CreatedAt = time.Now()
		return dao.CreateDeviceInfoDaily(ctx, &daily)
	}

	daily.ID = old.ID
	return dao.UpdateDeviceInfoDaily(ctx, &daily)
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

	var sum int64
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
		if len(dataY) > 0 {
			device.YesterdayProfit = Str2Float64(dataY["income"])
		}
		if len(dataS) > 0 {
			device.SevenDaysProfit = Str2Float64(dataS["income"])
		}
		if len(dataM) > 0 {
			device.MonthProfit = Str2Float64(dataM["income"])
		}
		if len(dataA) > 0 {
			device.CumulativeProfit = Str2Float64(dataA["income"])
		}
		if len(dataT) > 0 {
			device.TodayProfit = Str2Float64(dataT["income"])
			device.TodayOnlineTime = Str2Float64(dataT["online_time"])
		}
		device.UpdatedAt = time.Now()
		_, ok := DeviceIDAndUserId[device.DeviceID]
		if ok {
			device.UserID = DeviceIDAndUserId[device.DeviceID]
		}

		err = dao.UpdateDeviceInfo(context.Background(), device)
		if err != nil {
			log.Errorf("update device info: %v", err)
		}
	}

	opt.Page++
	sum += int64(len(devices))

	if sum < total {
		goto loop
	}

	return nil
}
