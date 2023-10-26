package statistics

import (
	"context"
	"fmt"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/utils"
	"github.com/golang-module/carbon/v2"
	"time"
)

func addDeviceInfoHours(ctx context.Context, deviceInfo []*model.DeviceInfo) error {
	log.Info("start to fetch 【device info hours】")
	start := time.Now()
	defer func() {
		log.Infof("fetch device info hours done, cost: %v", time.Since(start))
	}()

	var upsertDevice []*model.DeviceInfoHour
	for _, device := range deviceInfo {
		var deviceInfoHour model.DeviceInfoHour
		deviceOrdinaryInfo := dao.GetDeviceInfo(ctx, device.DeviceID)
		deviceInfoHour.RetrievalCount = device.RetrievalCount
		deviceInfoHour.BlockCount = device.CacheCount
		deviceInfoHour.DeviceID = device.DeviceID
		deviceInfoHour.UserID = deviceOrdinaryInfo.UserID
		deviceInfoHour.Time = start
		deviceInfoHour.DiskUsage = device.DiskUsage
		deviceInfoHour.DiskSpace = device.DiskSpace
		deviceInfoHour.HourIncome = device.CumulativeProfit
		deviceInfoHour.BandwidthUp = device.BandwidthUp
		deviceInfoHour.BandwidthDown = device.BandwidthDown
		deviceInfoHour.UpstreamTraffic = device.UploadTraffic
		deviceInfoHour.DownstreamTraffic = device.DownloadTraffic
		deviceInfoHour.OnlineTime = device.OnlineTime
		deviceInfoHour.CreatedAt = time.Now()
		deviceInfoHour.UpdatedAt = time.Now()
		upsertDevice = append(upsertDevice, &deviceInfoHour)
	}
	err := dao.BulkUpsertDeviceInfoHours(ctx, upsertDevice)
	if err != nil {
		log.Errorf("bulk upsert device info: %v", err)
	}

	if start.Minute() != 0 {
		return nil
	}

	// Add a redundant record to make it easier to count data within the range of 0-60 minutes
	for i := 0; i < len(upsertDevice); i++ {
		upsertDevice[i].Time = start.Add(-1 * time.Second)
	}

	err = dao.BulkUpsertDeviceInfoHours(ctx, upsertDevice)
	if err != nil {
		log.Errorf("bulk upsert device info: %v", err)
	}

	return nil
}

func QueryDataByDate(DateFrom, DateTo string) []map[string]string {

	sqlClause := fmt.Sprintf("select device_id, sum(income) as income,online_time from device_info_daily "+
		"where  time>='%s' and time<='%s' group by device_id;", DateFrom, DateTo)
	if DateFrom == "" {
		sqlClause = fmt.Sprintf("select device_id, sum(income) as income,online_time from device_info_daily " +
			" group by device_id;")
	}
	data, err := dao.GetQueryDataList(sqlClause)
	if err != nil {
		log.Error(err.Error())
		return nil
	}

	return data
}

func (s *Statistic) SumDeviceInfoDaily() error {
	log.Info("start to sum device info daily")
	start := time.Now()
	defer func() {
		log.Infof("sum device info daily done, cost: %v", time.Since(start))
	}()

	startOfTodayTime := carbon.Now().StartOfDay().String()
	endOfTodayTime := carbon.Now().Tomorrow().StartOfDay().String()
	var total int64
	where := fmt.Sprintf("where time>='%s' and time<='%s'", startOfTodayTime, endOfTodayTime)
	err := dao.DB.GetContext(s.ctx, &total, fmt.Sprintf(
		`SELECT count(*) FROM %s %s`, "device_info_hour", where,
	))
	sqlClause := fmt.Sprintf(`select i.user_id, i.device_id, date_format(i.time, '%%Y-%%m-%%d') as date, 
			i.nat_ratio, i.disk_usage, i.disk_space,i.latency, i.pkg_loss_ratio, i.bandwidth_up, i.bandwidth_down, 
			max(i.hour_income) - min(i.hour_income) as hour_income,
			max(i.online_time) - min(i.online_time) as online_time,
			max(i.upstream_traffic) - min(i.upstream_traffic) as upstream_traffic,
			max(i.downstream_traffic) - min(i.downstream_traffic) as downstream_traffic,
			max(i.retrieval_count) - min(i.retrieval_count) as retrieval_count,
			max(i.block_count) - min(i.block_count) as block_count
			from (select * from device_info_hour %s order by id desc limit %d) i                                                                                   
			group by device_id`, where, total)
	dataList, err := dao.GetQueryDataList(sqlClause)
	if err != nil {
		return err
	}

	var dailyInfos []*model.DeviceInfoDaily
	for _, data := range dataList {
		var daily model.DeviceInfoDaily
		daily.Time, _ = time.Parse(utils.TimeFormatDateOnly, data["date"])
		daily.DiskUsage = utils.Str2Float64(data["disk_usage"])
		daily.DiskSpace = utils.Str2Float64(data["disk_space"])
		daily.NatRatio = utils.Str2Float64(data["nat_ratio"])
		daily.Income = utils.Str2Float64(data["hour_income"])
		daily.OnlineTime = utils.Str2Float64(data["online_time"])
		if daily.OnlineTime > 1440 {
			daily.OnlineTime = 1440
		}
		daily.UpstreamTraffic = utils.Str2Float64(data["upstream_traffic"])
		daily.DownstreamTraffic = utils.Str2Float64(data["downstream_traffic"])
		daily.RetrievalCount = utils.Str2Int64(data["retrieval_count"])
		daily.BlockCount = utils.Str2Int64(data["block_count"])
		daily.PkgLossRatio = utils.Str2Float64(data["pkg_loss_ratio"])
		daily.Latency = utils.Str2Float64(data["latency"])
		daily.DeviceID = data["device_id"]
		daily.UserID = data["user_id"]
		daily.BandwidthUp = utils.Str2Float64(data["bandwidth_up"])
		daily.BandwidthDown = utils.Str2Float64(data["bandwidth_down"])
		daily.UserID = data["user_id"]
		daily.CreatedAt = time.Now()
		daily.UpdatedAt = time.Now()
		dailyInfos = append(dailyInfos, &daily)
	}

	err = dao.BulkUpsertDeviceInfoDaily(context.Background(), dailyInfos)
	if err != nil {
		return err
	}

	return nil
}

func (s *Statistic) SumDeviceInfoProfit() error {
	log.Info("start to sum device info profit")
	start := time.Now()
	defer func() {
		log.Infof("sum device info profit done, cost: %v", time.Since(start))
	}()

	if err := s.SumDeviceInfoDaily(); err != nil {
		log.Errorf("sum device info daily: %v", err)
		return err
	}

	updatedDevices := make(map[string]*model.DeviceInfo)
	startOfTodayTime := carbon.Now().StartOfDay().String()
	endOfTodayTime := carbon.Now().EndOfDay().String()
	startOfYesterday := carbon.Yesterday().StartOfDay().String()
	endOfYesterday := carbon.Yesterday().EndOfDay().String()
	dataY := QueryDataByDate(startOfYesterday, endOfYesterday)
	for _, data := range dataY {
		_, ok := updatedDevices[data["device_id"]]
		if !ok {
			updatedDevices[data["device_id"]] = &model.DeviceInfo{
				DeviceID: data["device_id"],
			}
		}
		updatedDevices[data["device_id"]].YesterdayProfit = utils.Str2Float64(data["income"])
	}

	startOfWeekTime := carbon.Now().SubDays(6).StartOfDay().String()
	dataS := QueryDataByDate(startOfWeekTime, endOfTodayTime)

	for _, data := range dataS {
		_, ok := updatedDevices[data["device_id"]]
		if !ok {
			updatedDevices[data["device_id"]] = &model.DeviceInfo{
				DeviceID: data["device_id"],
			}
		}
		updatedDevices[data["device_id"]].SevenDaysProfit = utils.Str2Float64(data["income"])
	}

	startOfMonthTime := carbon.Now().SubDays(29).StartOfDay().String()
	dataM := QueryDataByDate(startOfMonthTime, endOfTodayTime)

	for _, data := range dataM {
		_, ok := updatedDevices[data["device_id"]]
		if !ok {
			updatedDevices[data["device_id"]] = &model.DeviceInfo{
				DeviceID: data["device_id"],
			}
		}
		updatedDevices[data["device_id"]].MonthProfit = utils.Str2Float64(data["income"])
	}

	dataT := QueryDataByDate(startOfTodayTime, endOfTodayTime)
	for _, data := range dataT {
		_, ok := updatedDevices[data["device_id"]]
		if !ok {
			updatedDevices[data["device_id"]] = &model.DeviceInfo{
				DeviceID: data["device_id"],
			}
		}
		updatedDevices[data["device_id"]].TodayProfit = utils.Str2Float64(data["income"])
		updatedDevices[data["device_id"]].TodayOnlineTime = utils.Str2Float64(data["online_time"])
	}

	var deviceInfos []*model.DeviceInfo
	for _, deviceInfo := range updatedDevices {
		deviceInfos = append(deviceInfos, deviceInfo)
	}

	if err := dao.BulkUpdateDeviceInfo(context.Background(), deviceInfos); err != nil {
		log.Errorf("bulk update devices: %v", err)
	}

	return nil
}

func (s *Statistic) SumAllNodes() error {
	log.Info("start to sum all nodes")
	start := time.Now()
	defer func() {
		log.Infof("sum all nodes done, cost: %v", time.Since(start))
	}()
	fullNodeInfo, err := dao.SumFullNodeInfoFromDeviceInfo(s.ctx)
	if err != nil {
		log.Errorf("count full node: %v", err)
		return err
	}
	// todo NextElectionTime
	systemInfo, err := dao.SumSystemInfo(s.ctx)
	if err != nil {
		log.Errorf("sum system info: %v", err)
		return err
	}
	AssetCount := config.GNodesInfo.AssetCount
	if AssetCount == 0 {
		AssetCount = 1
	}
	fullNodeInfo.TAverageReplica = utils.ToFixed(float64(fullNodeInfo.TUpstreamFileCount)/float64(AssetCount), 2)
	fullNodeInfo.TotalCarfile = systemInfo.CarFileCount
	fullNodeInfo.RetrievalCount = systemInfo.DownloadCount
	fullNodeInfo.NextElectionTime = systemInfo.NextElectionTime
	fullNodeInfo.Time = time.Now()
	fullNodeInfo.CreatedAt = time.Now()

	stats, err := dao.CountStorageStats(s.ctx)
	if err != nil {
		log.Errorf("CountStorageStats: %v", err)
	}
	fullNodeInfo.FBackupsFromTitan = stats.StorageSize

	err = dao.CacheFullNodeInfo(s.ctx, fullNodeInfo)
	if err != nil {
		log.Errorf("cache full node info: %v", err)
		return err
	}

	fTime := fullNodeInfo.Time
	fullNodeInfo.Time = time.Date(fTime.Year(), fTime.Month(), fTime.Day(), 0, 0, 0, 0, time.Local)
	if err = dao.UpsertFullNodeInfo(s.ctx, fullNodeInfo); err != nil {
		log.Errorf("upsert full node: %v", err)
		return err
	}

	return nil
}

func (s *Statistic) UpdateDeviceRank() error {
	log.Info("start to rank device info")
	start := time.Now()
	defer func() {
		log.Infof("rank device info done, cost: %v", time.Since(start))
	}()
	if err := dao.RankDeviceInfo(s.ctx); err != nil {
		log.Errorf("ranking device info: %v", err)
		return err
	}
	return nil
}

func (s *Statistic) UpdateDeviceLocation() error {
	log.Info("start to update device location info")
	start := time.Now()
	defer func() {
		log.Infof("update device location done, cost: %v", time.Since(start))
	}()
	if err := dao.RankDeviceInfo(s.ctx); err != nil {
		log.Errorf("ranking device info: %v", err)
		return err
	}
	return nil
}
