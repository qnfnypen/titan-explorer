package dao

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/pkg/formatter"
	"github.com/golang-module/carbon/v2"
	logging "github.com/ipfs/go-log/v2"
	"time"
)

var (
	tableNameDeviceInfoHour  = "device_info_hour"
	tableNameDeviceInfoDaily = "device_info_daily"
	log                      = logging.Logger("device_info")
)

func BulkUpsertDeviceInfoHours(ctx context.Context, hourInfos []*model.DeviceInfoHour) error {
	upsertStatement := fmt.Sprintf(`INSERT INTO %s (created_at, updated_at, hour_income, user_id, device_id,
				online_time, pkg_loss_ratio, latency, nat_ratio, disk_usage,disk_space,bandwidth_up,bandwidth_down, upstream_traffic, downstream_traffic, retrieval_count, block_count, time)
			VALUES (:created_at, :updated_at, :hour_income, :user_id, :device_id, :online_time, :pkg_loss_ratio, :latency, 
			    :nat_ratio, :disk_usage, :disk_space,:bandwidth_up,:bandwidth_down, :upstream_traffic, :downstream_traffic, :retrieval_count, :block_count, :time) 
			 ON DUPLICATE KEY UPDATE updated_at = now(), hour_income = VALUES(hour_income), user_id = VALUES(user_id),
			online_time = VALUES(online_time), pkg_loss_ratio = VALUES(pkg_loss_ratio), latency = VALUES(latency),
			upstream_traffic = VALUES(upstream_traffic), downstream_traffic = VALUES(downstream_traffic), retrieval_count = VALUES(retrieval_count), block_count = VALUES(block_count),
			nat_ratio = VALUES(nat_ratio), disk_usage = VALUES(disk_usage),disk_space = VALUES(disk_space),bandwidth_up = VALUES(bandwidth_up),bandwidth_down = VALUES(bandwidth_down)`, tableNameDeviceInfoHour)
	_, err := DB.NamedExecContext(ctx, upsertStatement, hourInfos)
	return err
}

func BulkUpsertDeviceInfoDaily(ctx context.Context, dailyInfos []*model.DeviceInfoDaily) error {
	upsertStatement := fmt.Sprintf(`INSERT INTO %s (created_at, updated_at, income, user_id, device_id,
				online_time, pkg_loss_ratio, latency, nat_ratio, disk_usage, disk_space,bandwidth_up,bandwidth_down, upstream_traffic, downstream_traffic, retrieval_count, block_count, time)
			VALUES (:created_at, :updated_at, :income, :user_id, :device_id,
				:online_time, :pkg_loss_ratio, :latency, :nat_ratio, :disk_usage, :disk_space,:bandwidth_up,:bandwidth_down, :upstream_traffic, :downstream_traffic, :retrieval_count, :block_count, :time) 
			 ON DUPLICATE KEY UPDATE updated_at = now(), income = VALUES(income),user_id = VALUES(user_id),
			online_time = VALUES(online_time), pkg_loss_ratio = VALUES(pkg_loss_ratio), latency = VALUES(latency),
			upstream_traffic = VALUES(upstream_traffic), downstream_traffic = VALUES(downstream_traffic), retrieval_count = VALUES(retrieval_count), block_count = VALUES(block_count),
			nat_ratio = VALUES(nat_ratio), disk_usage = VALUES(disk_usage), disk_space = VALUES(disk_space),bandwidth_up = VALUES(bandwidth_up),bandwidth_down = VALUES(bandwidth_down)`, tableNameDeviceInfoDaily)

	_, err := DB.NamedExecContext(ctx, upsertStatement, dailyInfos)
	return err
}

type DeviceStatistics struct {
	Date              string  `json:"date" db:"date"`
	NatRatio          float64 `json:"nat_ratio" db:"nat_ratio"`
	DiskUsage         float64 `json:"disk_usage" db:"disk_usage"`
	DiskSpace         float64 `json:"disk_space" db:"disk_space"`
	Latency           float64 `json:"latency" db:"latency"`
	PkgLossRatio      float64 `json:"pkg_loss_ratio" db:"pkg_loss_ratio"`
	Income            float64 `json:"income" db:"income"`
	OnlineTime        float64 `json:"online_time" db:"online_time"`
	BandwidthUp       float64 `db:"bandwidth_up" json:"bandwidth_up"`
	BandwidthDown     float64 `db:"bandwidth_down" json:"bandwidth_down"`
	UpstreamTraffic   float64 `json:"upstream_traffic" db:"upstream_traffic"`
	DownstreamTraffic float64 `json:"downstream_traffic" db:"downstream_traffic"`
	RetrievalCount    float64 `json:"retrieval_count" db:"retrieval_count"`
	BlockCount        float64 `json:"block_count" db:"block_count"`
	NodeCount         float64 `json:"node_count" db:"node_count"`
}

func GetLatestDeviceStat(ctx context.Context, deviceId string, start string) (DeviceStatistics, error) {
	var ds DeviceStatistics
	query := fmt.Sprintf(`select 
    	hour_income as income, 
    	upstream_traffic, downstream_traffic, block_count, retrieval_count, 
    	online_time from device_info_hour where device_id = '%s' and time < '%s' order by time desc limit 1`, deviceId, start)

	err := DB.GetContext(ctx, &ds, query)
	if err == sql.ErrNoRows {
		//return DeviceStatistics{}, nil

		query2 := fmt.Sprintf(`select 
    	hour_income as income, 
    	upstream_traffic, downstream_traffic, block_count, retrieval_count, 
    	online_time from device_info_hour where device_id = '%s' order by time limit 1`, deviceId)

		err = DB.GetContext(ctx, &ds, query2)
		if err == sql.ErrNoRows {
			return DeviceStatistics{}, nil
		}
	}

	if err != nil {
		return DeviceStatistics{}, err
	}

	return ds, nil
}

func GetDeviceInfoHourList(ctx context.Context, cond *model.DeviceInfoHour, option QueryOption) ([]*DeviceStatistics, error) {
	sqlClause := fmt.Sprintf(`select date_format(time, '%%Y-%%m-%%d %%H') as date, 
	max(online_time) as online_time,
	max(hour_income) as income,
	max(upstream_traffic)  as upstream_traffic, 
	max(downstream_traffic)  as downstream_traffic,
	max(block_count)  as block_count,
	max(retrieval_count) as retrieval_count
	from %s where device_id='%s' and time>='%s' and time<='%s' group by date order by date`, tableNameDeviceInfoHour, cond.DeviceID, option.StartTime, option.EndTime)
	var out []*DeviceStatistics
	err := DB.SelectContext(ctx, &out, sqlClause)
	if err != nil {
		return nil, err
	}

	if len(out) > 0 {
		firstOneInRange, err := GetLatestDeviceStat(ctx, cond.DeviceID, option.StartTime)
		if err != nil {
			return nil, err
		}

		for _, ds := range out {
			tmp := *ds
			ds.OnlineTime -= firstOneInRange.OnlineTime
			if ds.OnlineTime > 60 {
				ds.OnlineTime = 60
			}

			if ds.OnlineTime < 0 {
				ds.OnlineTime = 0
			}

			ds.Income -= firstOneInRange.Income
			if ds.Income < 0 {
				ds.Income = 0
			}
			ds.UpstreamTraffic -= firstOneInRange.UpstreamTraffic
			ds.DownstreamTraffic -= firstOneInRange.DownstreamTraffic
			ds.BlockCount -= firstOneInRange.BlockCount
			ds.RetrievalCount -= firstOneInRange.RetrievalCount
			firstOneInRange = tmp
		}

		return append24HoursData(out), err
	}

	return append24HoursData(out), err
}

func GetDeviceInfoDailyListAppendDays(ctx context.Context, cond *model.DeviceInfoDaily, option QueryOption) ([]*DeviceStatistics, error) {
	var args []interface{}
	where := `WHERE 1=1`
	if cond.DeviceID != "" {
		where += ` AND device_id = ?`
		args = append(args, cond.DeviceID)
	}
	if option.StartTime != "" {
		where += ` AND time >= ?`
		args = append(args, option.StartTime)
	}
	if option.EndTime != "" {
		where += ` AND time <= ?`
		args = append(args, option.EndTime)
	}

	var result []*DeviceStatistics
	err := DB.SelectContext(ctx, &result, fmt.Sprintf(
		`SELECT DATE_FORMAT(time, '%%Y-%%m-%%d') as date, nat_ratio, disk_usage,disk_space,bandwidth_up,bandwidth_down, latency, pkg_loss_ratio, income, online_time, upstream_traffic, 
    	downstream_traffic, retrieval_count, block_count FROM %s %s`, tableNameDeviceInfoDaily, where), args...)
	if err != nil {
		return nil, err
	}

	return handleDailyList(result, option.StartTime, option.EndTime), err
}

func GetDeviceInfoDailyList(ctx context.Context, cond *model.DeviceInfoDaily, option QueryOption) ([]*DeviceStatistics, error) {
	var args []interface{}
	where := `WHERE 1=1`
	if cond.DeviceID != "" {
		where += ` AND device_id = ?`
		args = append(args, cond.DeviceID)
	}
	if option.StartTime != "" {
		where += ` AND time >= ?`
		args = append(args, option.StartTime)
	}
	if option.EndTime != "" {
		where += ` AND time <= ?`
		args = append(args, option.EndTime)
	}

	var result []*DeviceStatistics
	err := DB.SelectContext(ctx, &result, fmt.Sprintf(
		`SELECT DATE_FORMAT(time, '%%Y-%%m-%%d') as date, nat_ratio, disk_usage,disk_space,bandwidth_up,bandwidth_down, latency, pkg_loss_ratio, income, online_time, upstream_traffic, 
    	downstream_traffic, retrieval_count, block_count FROM %s %s`, tableNameDeviceInfoDaily, where), args...)
	if err != nil {
		return nil, err
	}

	return handleDailyListOld(result), err
}

func GetNodesInfoDailyList(ctx context.Context, cond *model.DeviceInfoDaily, option QueryOption) ([]*DeviceStatistics, error) {
	var args []interface{}
	where := `WHERE 1=1`
	if cond.UserID != "" || option.NotBound == "1" {
		where += ` AND user_id = ?`
		args = append(args, cond.UserID)
	}
	if option.StartTime != "" {
		where += ` AND time >= ?`
		args = append(args, option.StartTime)
	}
	if option.EndTime != "" {
		where += ` AND time <= ?`
		args = append(args, option.EndTime)
	}

	var result []*DeviceStatistics
	err := DB.SelectContext(ctx, &result, fmt.Sprintf(
		`SELECT DATE_FORMAT(time, '%%Y-%%m-%%d') as date, COUNT(DISTINCT(device_id)) as node_count,nat_ratio, ROUND(sum(disk_usage*disk_space/100),4) as disk_usage,ROUND(sum(disk_space),4) as disk_space,  ROUND(sum(income),2) as income, ROUND(sum(upstream_traffic),2) as upstream_traffic, 
    	ROUND(sum(bandwidth_up),2) as bandwidth_up,ROUND(sum(bandwidth_down),2) as bandwidth_down,
    	ROUND(sum(downstream_traffic),2) as downstream_traffic, ROUND(sum(retrieval_count),2) as retrieval_count,ROUND(sum(block_count),2) as block_count FROM %s %s group by date`, tableNameDeviceInfoDaily, where), args...)
	if err != nil {
		return nil, err
	}

	return handleDailyList(result, option.StartTime, option.EndTime), err
}

func dateKey(t time.Time) string {
	return t.Format(formatter.TimeFormatMD)
}

func handleDailyList(deviceStat []*DeviceStatistics, start, end string) []*DeviceStatistics {
	startTime, endTime := carbon.Parse(start), carbon.Parse(end)
	deviceInDate := make(map[string]*DeviceStatistics)

	for _, data := range deviceStat {
		deviceInDate[dateKey(carbon.Parse(data.Date).StdTime())] = data
	}

	var out []*DeviceStatistics
	for st := startTime.StartOfDay(); st.Lte(endTime.StartOfDay()); st = st.AddDay() {
		if val, ok := deviceInDate[dateKey(st.StdTime())]; ok {
			out = append(out, val)
			continue
		}
		out = append(out, &DeviceStatistics{
			Date: dateKey(st.StdTime()),
		})
	}

	return out

}

func handleDailyListOld(deviceStat []*DeviceStatistics) []*DeviceStatistics {
	now := time.Now()
	startTime, endTime := now, now
	oneDay := 24 * time.Hour
	deviceInDate := make(map[string]*DeviceStatistics)

	for _, data := range deviceStat {
		t, _ := time.Parse(time.DateOnly, data.Date)

		if t.Before(startTime) {
			startTime = t
		}

		if t.After(endTime) {
			endTime = t
		}

		deviceInDate[data.Date] = data
	}

	var out []*DeviceStatistics
	for startTime.Before(endTime) || startTime.Equal(endTime) {
		key := startTime.Format(formatter.TimeFormatDateOnly)
		val, ok := deviceInDate[key]
		if !ok {
			val = &DeviceStatistics{}
		}
		val.Date = startTime.Format(formatter.TimeFormatMD)
		out = append(out, val)
		startTime = startTime.Add(oneDay)
	}

	return out

}

func append24HoursData(in []*DeviceStatistics) []*DeviceStatistics {
	now := time.Now()
	oneHour := time.Hour
	startTime := now.Add(-23 * oneHour)
	endTime := now
	deviceInDate := make(map[string]*DeviceStatistics)
	var out []*DeviceStatistics
	for _, data := range in {
		deviceInDate[data.Date] = data
	}
	for startTime.Before(endTime) || startTime.Equal(endTime) {
		key := startTime.Format(formatter.TimeFormatYMDH)
		val, ok := deviceInDate[key]
		if !ok {
			val = &DeviceStatistics{}
		}
		val.Date = fmt.Sprintf("%d:00", startTime.Hour())
		out = append(out, val)
		startTime = startTime.Add(oneHour)
	}

	return out
}

func GetUserIncome(cond *model.DeviceInfo, option QueryOption) (map[string]map[string]interface{}, error) {
	sqlClause := fmt.Sprintf(`select date_format(time, '%%Y-%%m-%%d') as date, sum(income) as income from device_info_daily where user_id = '%s' and  time >='%s' and time <= '%s'  GROUP BY date`, cond.UserID, option.StartTime, option.EndTime)
	dataS, err := GetQueryDataList(sqlClause)
	if err != nil {
		return nil, err
	}
	out := make(map[string]map[string]interface{})
	for _, data := range dataS {
		_, ok := out[data["date"]]
		if !ok {
			out[data["date"]] = make(map[string]interface{})
		}
		out[data["date"]]["income"] = formatter.StrToFloat(data["income"])
	}
	return out, nil
}

func GetDeviceInfoDailyByPage(ctx context.Context, cond *model.DeviceInfoDaily, option QueryOption) ([]*model.DeviceInfoDaily, int64, error) {
	var args []interface{}
	where := `WHERE 1=1`
	if cond.DeviceID != "" {
		where += ` AND device_id = ?`
		args = append(args, cond.DeviceID)
	}

	if option.Order != "" && option.OrderField != "" {
		where += fmt.Sprintf(` ORDER BY %s %s`, option.OrderField, option.Order)
	}

	limit := option.PageSize
	offset := option.Page
	if option.PageSize <= 0 {
		limit = 50
	}
	if option.Page > 0 {
		offset = limit * (option.Page - 1)
	}

	var total int64
	var out []*model.DeviceInfoDaily

	err := DB.GetContext(ctx, &total, fmt.Sprintf(
		`SELECT count(*) FROM %s %s`, tableNameDeviceInfo, where,
	), args...)
	if err != nil {
		return nil, 0, err
	}
	err = DB.SelectContext(ctx, &out, fmt.Sprintf(
		`SELECT * FROM %s %s LIMIT %d OFFSET %d`, tableNameDeviceInfoDaily, where, limit, offset), args...)
	if err != nil {
		return nil, 0, err
	}

	return out, total, err
}

func GetAllUsersRewardBefore(ctx context.Context, before string) (map[string]*model.UserRewardDaily, error) {
	query := `select 
	user_id, 
	referrer_user_id,
	max(cumulative_reward) as cumulative_reward,
	sum(app_reward) as app_reward,
	sum(cli_reward) as cli_reward,
	max(total_device_count) as total_device_count,
	max(device_online_count) as device_online_count,
	sum(referral_reward) as referral_reward,
	sum(referrer_reward) as referrer_reward,
	sum(kol_bonus) as kol_bonus
	from user_reward_daily where time < ? group by user_id`

	rows, err := DB.QueryxContext(ctx, query, before)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]*model.UserRewardDaily)

	for rows.Next() {
		var userRewardDaily model.UserRewardDaily
		if err := rows.StructScan(&userRewardDaily); err != nil {
			log.Errorf("struct scan: %v", err)
			continue
		}

		out[userRewardDaily.UserId] = &userRewardDaily
	}

	return out, nil
}

func BulkAddUserRewardDaily(ctx context.Context, userRewards []*model.UserRewardDaily) error {
	update := `INSERT INTO user_reward_daily (user_id, cumulative_reward, reward, app_reward, cli_reward, kol_bonus, referral_reward, device_online_count, total_device_count,
                               referrer_user_id, is_kol, is_referrer_kol, commission_percent, kol_bonus_percent, referrer_reward, time, created_at, updated_at)
			VALUES (:user_id, :cumulative_reward, :reward, :app_reward, :cli_reward, :kol_bonus, :referral_reward, :device_online_count, :total_device_count,
			        :referrer_user_id, :is_kol, :is_referrer_kol, :commission_percent, :kol_bonus_percent, :referrer_reward, :time, :created_at, :updated_at) 
			 ON DUPLICATE KEY UPDATE cumulative_reward = VALUES(cumulative_reward), reward = VALUES(reward), app_reward = VALUES(app_reward), cli_reward = VALUES(cli_reward), kol_bonus = VALUES(kol_bonus), 
			  referral_reward = VALUES(referral_reward), referrer_user_id = VALUES(referrer_user_id), is_kol = VALUES(is_kol), is_referrer_kol = VALUES(is_referrer_kol), commission_percent = VALUES(commission_percent),
			  kol_bonus_percent = VALUES(kol_bonus_percent), referrer_reward = VALUES(referrer_reward), device_online_count = VALUES(device_online_count), total_device_count = VALUES(total_device_count), updated_at = now()`

	_, err := DB.NamedExecContext(ctx, update, userRewards)
	return err
}

func GetUserReferralReward(ctx context.Context, option QueryOption) ([]*model.UserReferralRecord, int64, error) {
	var args []interface{}
	where := `WHERE referrer_user_id <> '' `

	if option.UserID != "" {
		args = append(args, option.UserID)
		where += ` and user_id = ?`
	}

	if option.StartTime != "" {
		args = append(args, option.StartTime)
		where += ` and time >= ?`
	}

	if option.EndTime != "" {
		args = append(args, option.EndTime)
		where += ` and time < ?`
	}

	if option.Order != "" && option.OrderField != "" {
		where += fmt.Sprintf(` ORDER BY %s %s`, option.OrderField, option.Order)
	} else {
		where += fmt.Sprintf(" ORDER BY time desc")
	}

	limit := option.PageSize
	offset := option.Page
	if option.PageSize <= 0 {
		limit = 50
	}
	if option.Page > 0 {
		offset = limit * (option.Page - 1)
	}

	var total int64
	var out []*model.UserReferralRecord

	err := DB.GetContext(ctx, &total, fmt.Sprintf(
		`SELECT count(*) FROM user_reward_daily %s`, where,
	), args...)
	if err != nil {
		return nil, 0, err
	}

	err = DB.SelectContext(ctx, &out, fmt.Sprintf(
		`SELECT user_id, referrer_user_id, device_online_count, reward, referrer_reward, time as updated_at FROM user_reward_daily %s LIMIT %d OFFSET %d`, where, limit, offset), args...)
	if err != nil {
		return nil, 0, err
	}

	return out, total, err
}

func LoadAllUserReferralReward(ctx context.Context, option QueryOption) ([]*model.UserReferralRecord, error) {
	var args []interface{}
	where := `WHERE referrer_user_id <> '' `

	if option.UserID != "" {
		args = append(args, option.UserID)
		where += ` and user_id = ?`
	}

	if option.StartTime != "" {
		args = append(args, option.StartTime)
		where += ` and time >= ?`
	}

	if option.EndTime != "" {
		args = append(args, option.EndTime)
		where += ` and time < ?`
	}

	if option.Order != "" && option.OrderField != "" {
		where += fmt.Sprintf(` ORDER BY %s %s`, option.OrderField, option.Order)
	}

	var out []*model.UserReferralRecord

	err := DB.SelectContext(ctx, &out, fmt.Sprintf(`SELECT user_id, referrer_user_id, device_online_count, reward, referrer_reward, time as updated_at FROM user_reward_daily %s`, where), args...)
	if err != nil {
		return nil, err
	}

	return out, err
}

func BulkUpdateUserReward(ctx context.Context, users []*model.User) error {
	query := `INSERT INTO users (username, reward, referral_reward, referrer_commission_reward, from_kol_bonus_reward, device_count, device_online_count, updated_at) 
	VALUES (:username, :reward, :referral_reward, :referrer_commission_reward, :from_kol_bonus_reward, :device_count, :device_online_count, :updated_at) ON DUPLICATE KEY UPDATE reward = VALUES(reward), 
         referral_reward = VALUES(referral_reward), referrer_commission_reward = VALUES(referrer_commission_reward), from_kol_bonus_reward = VALUES(from_kol_bonus_reward),
        device_count = VALUES(device_count), device_online_count = VALUES(device_online_count), updated_at  = now()`
	_, err := DB.NamedExecContext(ctx, query, users)
	return err
}

func BulkUpdateUserReferralReward(ctx context.Context, users []*model.User) error {
	query := `INSERT INTO users (username, referral_reward, updated_at) 
	VALUES (:username, :referral_reward, :updated_at) ON DUPLICATE KEY UPDATE referral_reward = VALUES(referral_reward), updated_at  = now()`
	_, err := DB.NamedExecContext(ctx, query, users)
	return err
}
