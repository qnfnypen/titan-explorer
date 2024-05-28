package statistics

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/pkg/formatter"
	"github.com/golang-module/carbon/v2"
	errs "github.com/pkg/errors"
	"time"
)

const (
	DefaultCommissionPercent = 5
)

const batchSize = 1000

// AddDeviceInfoHours 写入 device_info_hour 表
func AddDeviceInfoHours(ctx context.Context, upsertDevice []*model.DeviceInfoHour) error {
	log.Info("start to fetch device info hours")
	start := time.Now()
	defer func() {
		log.Infof("fetch device info hours done, cost: %v", time.Since(start))
	}()

	err := dao.BulkUpsertDeviceInfoHours(ctx, upsertDevice)
	if err != nil {
		log.Errorf("bulk upsert device info: %v", err)
	}

	if start.Minute() != 0 {
		return nil
	}

	// Add a redundant record to make it easier to count data within the range of 0-60 minutes
	for i := 0; i < len(upsertDevice); i++ {
		upsertDevice[i].Time = start.Add(-1 * time.Minute)
	}

	err = dao.BulkUpsertDeviceInfoHours(ctx, upsertDevice)
	if err != nil {
		log.Errorf("bulk upsert device info: %v", err)
	}

	return nil
}

func GetDeviceUserId(ctx context.Context, deviceId string) string {
	deviceOrdinaryInfo, err := dao.GetDeviceInfo(ctx, deviceId)
	if err != nil {
		log.Errorf("get device info: %v", err)
		return ""
	}

	if deviceOrdinaryInfo.UserID != "" {
		err = dao.SetDeviceUserIdToCache(ctx, deviceId, deviceOrdinaryInfo.UserID, deviceOrdinaryInfo.AreaID)
		if err != nil {
			log.Errorf("set device user to cahce: %v", err)
		}

		return deviceOrdinaryInfo.UserID
	}

	signature, err := dao.GetSignatureByNodeId(ctx, deviceId)
	if err != nil {
		log.Errorf("get signatrue by node id: %v", err)
		//return ""
	}

	if signature == nil {
		return ""
	}

	if err = dao.UpdateUserDeviceInfo(ctx, &model.DeviceInfo{
		UserID:     signature.Username,
		DeviceID:   deviceId,
		BindStatus: "binding",
	}); err != nil {
		log.Errorf("update device binding status: %v", err)
	}

	return signature.Username
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

	if err := s.calculateAndUpsertDailyInfo(); err != nil {
		log.Errorf("calculate and upsert daily info: %v", err)
		return err
	}

	return nil
}

func (s *Statistic) calculateAndUpsertDailyInfo() error {
	startOfTodayTime := carbon.Now().StartOfDay().String()
	endOfTodayTime := carbon.Now().Tomorrow().StartOfDay().String()

	total, err := s.countDeviceInfoHour(startOfTodayTime, endOfTodayTime)
	if err != nil {
		return errs.Wrap(err, "count device info hour")
	}

	where := fmt.Sprintf("where time>='%s' and time<='%s'", startOfTodayTime, endOfTodayTime)
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
		return errs.Wrap(err, "get query data list")
	}

	dailyInfos, err := s.buildDailyInfos(dataList)
	if err != nil {
		return errs.Wrap(err, "build daily infos")
	}

	err = dao.BulkUpsertDeviceInfoDaily(context.Background(), dailyInfos)
	if err != nil {
		return errs.Wrap(err, "bulk upsert device info daily")
	}

	return nil
}

func (s *Statistic) countDeviceInfoHour(start, end string) (int64, error) {
	where := fmt.Sprintf("where time>='%s' and time<='%s'", start, end)
	var total int64
	err := dao.DB.GetContext(s.ctx, &total, fmt.Sprintf(`SELECT count(*) FROM %s %s`, "device_info_hour", where))
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (s *Statistic) buildDailyInfos(dataList []map[string]string) ([]*model.DeviceInfoDaily, error) {
	var dailyInfos []*model.DeviceInfoDaily
	for _, data := range dataList {
		var daily model.DeviceInfoDaily
		daily.Time, _ = time.Parse(formatter.TimeFormatDateOnly, data["date"])
		daily.DiskUsage = formatter.Str2Float64(data["disk_usage"])
		daily.DiskSpace = formatter.Str2Float64(data["disk_space"])
		daily.NatRatio = formatter.Str2Float64(data["nat_ratio"])
		daily.Income = formatter.Str2Float64(data["hour_income"])
		daily.OnlineTime = formatter.Str2Float64(data["online_time"])
		if daily.OnlineTime > 1440 {
			daily.OnlineTime = 1440
		}
		daily.UpstreamTraffic = formatter.Str2Float64(data["upstream_traffic"])
		daily.DownstreamTraffic = formatter.Str2Float64(data["downstream_traffic"])
		daily.RetrievalCount = formatter.Str2Int64(data["retrieval_count"])
		daily.BlockCount = formatter.Str2Int64(data["block_count"])
		daily.PkgLossRatio = formatter.Str2Float64(data["pkg_loss_ratio"])
		daily.Latency = formatter.Str2Float64(data["latency"])
		daily.DeviceID = data["device_id"]
		daily.UserID = data["user_id"]
		daily.BandwidthUp = formatter.Str2Float64(data["bandwidth_up"])
		daily.BandwidthDown = formatter.Str2Float64(data["bandwidth_down"])
		daily.UserID = data["user_id"]
		daily.CreatedAt = time.Now()
		daily.UpdatedAt = time.Now()
		dailyInfos = append(dailyInfos, &daily)
	}
	return dailyInfos, nil
}

func SumDeviceInfoProfit() error {
	log.Info("start to sum device info profit")
	start := time.Now()
	defer func() {
		log.Infof("sum device info profit done, cost: %v", time.Since(start))
	}()

	updatedDevices := make(map[string]*model.DeviceInfo)

	updateDeviceInfoForTimeRange(updatedDevices, carbon.Yesterday(), carbon.Yesterday(), "YesterdayProfit")
	updateDeviceInfoForTimeRange(updatedDevices, carbon.Now().SubDays(6), carbon.Now(), "SevenDaysProfit")
	updateDeviceInfoForTimeRange(updatedDevices, carbon.Now().SubDays(29), carbon.Now(), "MonthProfit")
	updateDeviceInfoForTimeRange(updatedDevices, carbon.Now(), carbon.Now(), "TodayProfit", "TodayOnlineTime")

	var count int
	deviceInfos := make([]*model.DeviceInfo, 0)
	for _, deviceInfo := range updatedDevices {
		count++

		deviceInfos = append(deviceInfos, deviceInfo)

		if len(deviceInfos) == 1000 || count == len(updatedDevices) {
			if err := dao.BulkUpdateDeviceInfo(context.Background(), deviceInfos); err != nil {
				log.Errorf("bulk update devices: %v", err)
			}
			deviceInfos = make([]*model.DeviceInfo, 0)
		}
	}

	return nil
}

func updateDeviceInfoForTimeRange(updatedDevices map[string]*model.DeviceInfo, start, end carbon.Carbon, profitFields ...string) {
	data := QueryDataByDate(start.StartOfDay().String(), end.EndOfDay().String())
	for _, d := range data {
		deviceID := d["device_id"]
		device, ok := updatedDevices[deviceID]
		if !ok {
			device = &model.DeviceInfo{DeviceID: deviceID}
			updatedDevices[deviceID] = device
		}

		for _, field := range profitFields {
			switch field {
			case "YesterdayProfit":
				updatedDevices[deviceID].YesterdayProfit = formatter.Str2Float64(d["income"])
			case "SevenDaysProfit":
				updatedDevices[deviceID].SevenDaysProfit = formatter.Str2Float64(d["income"])
			case "MonthProfit":
				updatedDevices[deviceID].MonthProfit = formatter.Str2Float64(d["income"])
			case "TodayProfit":
				updatedDevices[deviceID].TodayProfit = formatter.Str2Float64(d["income"])
				updatedDevices[deviceID].TodayOnlineTime = formatter.Str2Float64(d["online_time"])
			}
		}
	}
}

// UpdateUserRewardFields
// update users table reward, referral_reward, referrer_commission_reward, from_kol_bonus_reward, device_count fields
func UpdateUserRewardFields(ctx context.Context) error {
	log.Info("start to update user reward")
	start := time.Now()
	defer func() {
		log.Infof("update user reward done, cost: %v", time.Since(start))
	}()

	beforeDate := carbon.Now().EndOfDay().String()
	userRewards, err := dao.GetAllUsersRewardBefore(ctx, beforeDate)
	if err != nil {
		log.Errorf("GetAllUsersRewardBefore: %v", err)
		return err
	}

	var (
		count              int
		todos              []*model.User
		referrerUpdateList = make(map[string]float64)
	)

	for _, user := range userRewards {

		count++
		todos = append(todos, &model.User{
			Username:                 user.UserId,
			Reward:                   user.CumulativeReward,
			RefereralReward:          user.ReferralReward,
			DeviceCount:              user.TotalDeviceCount,
			DeviceOnlineCount:        user.DeviceOnlineCount,
			ReferrerCommissionReward: user.ReferrerReward,
			FromKOLBonusReward:       user.KOLBonus,
		})

		// referrer hasn't run any nodes, so he isn't in the map.
		if _, ok := userRewards[user.ReferrerUserId]; user.ReferrerUserId != "" && !ok {
			referrerUpdateList[user.ReferrerUserId] += user.ReferrerReward
		}

		if len(todos)%batchSize == 0 || count == len(userRewards) {
			if err = dao.BulkUpdateUserReward(ctx, todos); err != nil {
				log.Errorf("BulkUpdateUserReward: %v", err)
			}
			todos = todos[:0]
		}
	}

	todos = todos[:0]
	for userId, reward := range referrerUpdateList {
		todos = append(todos, &model.User{
			Username:        userId,
			RefereralReward: reward,
		})
	}

	if err = dao.BulkUpdateUserReferralReward(ctx, todos); err != nil {
		log.Errorf("BulkUpdateUserReferralReward: %v", err)
	}

	return nil
}

// SumUserDailyReward
// 奖励规则
// 普通用户:
//   - 邀请人, 可得5%的佣金
//   - 受邀人, 无津贴
//
// KOL:
//   - 邀请人, 可得 cli 端 5% 佣金, 移动端 10%-15%-20% 的对应等级比例佣金
//   - 受邀人, 5%-7%-10% 的津贴
func SumUserDailyReward(ctx context.Context) error {
	log.Info("start to sum user daily reward")
	start := time.Now()
	defer func() {
		log.Infof("sum user daily reward done, cost: %v", time.Since(start))
	}()

	userRewardSum, err := dao.SumAllUsersReward(ctx)
	if err != nil {
		log.Errorf("SumAllUsersReward: %v", err)
		return err
	}

	kolLevels, err := dao.GetAllKOLLevels(ctx)
	if err != nil {
		log.Errorf("GetAllKOLLevels: %v", err)
		return err
	}

	referrerUserIdInUser, err := dao.GetAllUserReferrerUserId(ctx)
	if err != nil {
		log.Errorf("GetAllUserReferrerUserId: %v", err)
		return err
	}
	//
	//beforeDate := carbon.Yesterday().EndOfDay().String()
	//maxUserRewardBefore, err := dao.GetAllUsersRewardBefore(ctx, beforeDate)
	//if err != nil {
	//	log.Errorf("GetAllUsersRewardBefore: %v", err)
	//	return err
	//}

	referralReward := make(map[string]float64)

	toLevelUpKOLs := make(map[string]*kolLevelRef)

	var updateUserRewards []*model.UserRewardDaily
	for _, userReward := range userRewardSum {
		userReward.UpdatedAt = start
		userReward.CreatedAt = start
		userReward.Time = carbon.Now().StartOfDay().StdTime()

		if _, ok := kolLevels[userReward.UserId]; ok {
			userReward.IsKOL = 1
		}

		if v, ok := referrerUserIdInUser[userReward.UserId]; ok {
			userReward.ReferrerUserId = v
		}

		//if before, ok := maxUserRewardBefore[userReward.UserId]; ok {
		//	userReward.Reward = unSizeVal(userReward.CumulativeReward - before.CumulativeReward)
		//	userReward.AppReward = unSizeVal(userReward.AppReward - before.AppReward)
		//	userReward.CliReward = unSizeVal(userReward.CliReward - before.CliReward)
		//} else {
		//	userReward.Reward = userReward.CumulativeReward
		//}

		if kol, ok := kolLevels[userReward.ReferrerUserId]; ok {
			userReward.IsReferrerKOL = 1
			userReward.KOLBonus = userReward.Reward * float64(kol.ChildrenBonusPercent) / float64(100)
			userReward.CommissionPercent = int64(kol.ParentCommissionPercent)
			userReward.KOLBonusPercent = int64(kol.ChildrenBonusPercent)

			cliReward := userReward.CliReward * DefaultCommissionPercent / 100
			appReward := userReward.AppReward * float64(kol.ParentCommissionPercent) / float64(100)

			userReward.ReferrerReward = cliReward + appReward
			referralReward[userReward.ReferrerUserId] += cliReward + appReward

			if _, exist := toLevelUpKOLs[userReward.ReferrerUserId]; !exist {
				toLevelUpKOLs[userReward.ReferrerUserId] = &kolLevelRef{
					CurrentLevel:               kol.Level,
					LevelUpNodeCountsThreshold: kol.DeviceThreshold,
				}
			}

			toLevelUpKOLs[userReward.ReferrerUserId].ReferralNodeCount += userReward.DeviceOnlineCount
		} else {
			userReward.IsReferrerKOL = 0
			reward := userReward.Reward * DefaultCommissionPercent / 100
			userReward.ReferrerReward = reward
			referralReward[userReward.ReferrerUserId] += reward
		}

		updateUserRewards = append(updateUserRewards, userReward)
	}

	var todos []*model.UserRewardDaily

	for i, u := range updateUserRewards {
		if reward, ok := referralReward[u.UserId]; ok {
			u.ReferralReward = reward
		}

		todos = append(todos, u)

		// Perform bulk insert when todos reaches a multiple of batchSize or at the end of updateUserRewards
		if len(todos)%batchSize == 0 || i == len(updateUserRewards)-1 {
			if err := dao.BulkAddUserRewardDaily(context.Background(), todos); err != nil {
				return errs.Wrap(err, "BulkAddUserRewardDaily")
			}
			todos = todos[:0] // Reset todos slice without reallocating memory
		}
	}

	if err = checkingKOLLevelUp(ctx, toLevelUpKOLs); err != nil {
		log.Errorf("checkingKOLLevelUp: %v", err)
	}

	return nil
}

type kolLevelRef struct {
	CurrentLevel               int
	ReferralNodeCount          int64
	LevelUpNodeCountsThreshold int64
}

func checkingKOLLevelUp(ctx context.Context, nodeCountInKOL map[string]*kolLevelRef) error {
	// Fetch all KOL levels once
	kolLevels, _, err := dao.GetKolLevelConfig(ctx, dao.QueryOption{})
	if err != nil {
		return errs.Wrap(err, "GetAllKOLLevels")
	}

	for kolUserId, kolLevel := range nodeCountInKOL {
		if kolLevel.ReferralNodeCount < kolLevel.LevelUpNodeCountsThreshold {
			log.Infof("Level up conditions not met for user %s, %d, %d", kolUserId, kolLevel.ReferralNodeCount, kolLevel.LevelUpNodeCountsThreshold)
			continue
		}

		nextLevel := kolLevel.CurrentLevel + 1

		// Check if next level exists
		nextLevelExists := false
		for _, level := range kolLevels {
			if level.Level == nextLevel {
				nextLevelExists = true
				break
			}
		}

		if !nextLevelExists {
			log.Infof("MAX KOL level reached for user %s", kolUserId)
			continue
		}

		log.Infof("KOL LEVEL UP: %s, before Level: %d, after Level: %d", kolUserId, kolLevel.CurrentLevel, nextLevel)

		record := &model.KOLLevelUPRecord{
			UserId:             kolUserId,
			BeforeLevel:        int64(kolLevel.CurrentLevel),
			AfterLevel:         int64(nextLevel),
			ReferralNodesCount: kolLevel.ReferralNodeCount,
			CreatedAt:          time.Now(),
		}

		// Update KOL level
		if err := dao.UpdateKOLLevel(ctx, kolUserId, nextLevel); err != nil {
			return errs.Wrap(err, "UpdateKOLLevel")
		}

		// Prepare KOL level-up record
		if err = dao.AddKOLLevelUPRecord(ctx, record); err != nil {
			return errs.Wrap(err, "AddKOLLevelUPRecord")
		}
	}

	return nil
}

func unSizeVal(val float64) float64 {
	if val < 0 {
		return 0
	}
	return val
}

func SumAllNodes() error {
	log.Info("start to sum all nodes")
	start := time.Now()
	defer func() {
		log.Infof("sum all nodes done, cost: %v", time.Since(start))
	}()

	ctx := context.Background()
	fullNodeInfo, err := dao.SumFullNodeInfoFromDeviceInfo(ctx)
	if err != nil {
		log.Errorf("count full node: %v", err)
		return err
	}

	// todo NextElectionTime
	systemInfo, err := dao.SumSystemInfo(ctx)
	if err != nil {
		log.Errorf("sum system info: %v", err)
		return err
	}
	AssetCount := config.GNodesInfo.AssetCount
	if AssetCount == 0 {
		AssetCount = 1
	}
	fullNodeInfo.TAverageReplica = formatter.ToFixed(float64(fullNodeInfo.TUpstreamFileCount)/float64(AssetCount), 2)
	fullNodeInfo.TotalCarfile = systemInfo.CarFileCount
	fullNodeInfo.RetrievalCount = systemInfo.DownloadCount
	fullNodeInfo.NextElectionTime = systemInfo.NextElectionTime
	fullNodeInfo.Time = time.Now()
	fullNodeInfo.CreatedAt = time.Now()

	stats, err := dao.CountStorageStats(ctx)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Errorf("CountStorageStats: %v", err)
	}
	if stats != nil {
		fullNodeInfo.FBackupsFromTitan = stats.TotalSize
	} else {
		sum, err := dao.SumFilStorage(ctx)
		if err != nil {
			log.Errorf("CountStorageStats: %v", err)
		}
		fullNodeInfo.FBackupsFromTitan = float64(sum)
	}

	err = dao.CacheFullNodeInfo(ctx, fullNodeInfo)
	if err != nil {
		log.Errorf("cache full node info: %v", err)
		return err
	}

	fTime := fullNodeInfo.Time
	fullNodeInfo.Time = time.Date(fTime.Year(), fTime.Month(), fTime.Day(), 0, 0, 0, 0, time.Local)
	if err = dao.UpsertFullNodeInfo(ctx, fullNodeInfo); err != nil {
		log.Errorf("upsert full node: %v", err)
		return err
	}

	return nil
}
