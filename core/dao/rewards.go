package dao

import (
	"context"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/utils"
	logging "github.com/ipfs/go-log/v2"
	"strings"
)

var (
	tableNameDeviceInfoHour  = "device_info_hour"
	tableNameDeviceInfoDaily = "device_info_daily"
	log                      = logging.Logger("statistics")
)

func BulkUpsertDeviceInfoHours(ctx context.Context, hourInfos []*model.DeviceInfoHour) error {
	upsertStatement := fmt.Sprintf(`INSERT INTO %s (created_at, updated_at, hour_income, user_id, device_id,
				online_time, pkg_loss_ratio, latency, nat_ratio, disk_usage, upstream_traffic, downstream_traffic, retrieve_count, block_count, time)
			VALUES (:created_at, :updated_at, :hour_income, :user_id, :device_id, :online_time, :pkg_loss_ratio, :latency, 
			    :nat_ratio, :disk_usage, :upstream_traffic, :downstream_traffic, :retrieve_count, :block_count, :time) 
			 ON DUPLICATE KEY UPDATE updated_at = now(), hour_income = :hour_income,
			online_time = :online_time, pkg_loss_ratio = :pkg_loss_ratio, latency = :latency,
			upstream_traffic = :upstream_traffic, downstream_traffic = :downstream_traffic, retrieve_count = :retrieve_count, block_count = :block_count,
			nat_ratio = :nat_ratio, disk_usage = :disk_usage`, tableNameDeviceInfoHour)
	tx := DB.MustBegin()
	defer tx.Rollback()

	for _, hourInfo := range hourInfos {
		_, err := tx.NamedExecContext(ctx, upsertStatement, hourInfo)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func BulkUpsertDeviceInfoDaily(ctx context.Context, dailyInfos []*model.DeviceInfoDaily) error {
	upsertStatement := fmt.Sprintf(`INSERT INTO %s (created_at, updated_at, income, user_id, device_id,
				online_time, pkg_loss_ratio, latency, nat_ratio, disk_usage, upstream_traffic, downstream_traffic, retrieve_count, block_count, time)
			VALUES (:created_at, :updated_at, :income, :user_id, :device_id,
				:online_time, :pkg_loss_ratio, :latency, :nat_ratio, :disk_usage, :upstream_traffic, :downstream_traffic, :retrieve_count, :block_count, :time) 
			 ON DUPLICATE KEY UPDATE updated_at = now(), income = :income,
			online_time = :online_time, pkg_loss_ratio = :pkg_loss_ratio, latency = :latency,
			upstream_traffic = :upstream_traffic, downstream_traffic = :downstream_traffic, retrieve_count = :retrieve_count, block_count = :block_count,
			nat_ratio = :nat_ratio, disk_usage = :disk_usage`, tableNameDeviceInfoDaily)
	tx := DB.MustBegin()
	defer tx.Rollback()

	for _, dailyInfo := range dailyInfos {
		_, err := tx.NamedExecContext(ctx, upsertStatement, dailyInfo)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func GetDeviceInfoDailyHourList(ctx context.Context, cond *model.DeviceInfoHour, option QueryOption) ([]map[string]string, error) {
	sqlClause := fmt.Sprintf(`select user_id,date_format(time, '%%Y-%%m-%%d %%H') as date, avg(nat_ratio) as nat_ratio, 
	avg(disk_usage) as disk_usage, avg(latency) as latency, avg(pkg_loss_ratio) as pkg_loss_ratio, 
	max(hour_income) as hour_income_max, min(hour_income) as hour_income_min,
	max(online_time) as online_time_max,min(online_time) as online_time_min,
	max(upstream_traffic) as upstream_traffic_max, min(upstream_traffic) as upstream_traffic_min,
	max(downstream_traffic) as downstream_traffic_max, min(downstream_traffic) as downstream_traffic_min,
	max(retrieve_count) as retrieve_count_max, min(retrieve_count) as retrieve_count_min,
	from device_info_hour where device_id='%s' and time>='%s' and time<='%s' group by date`, cond.DeviceID, option.StartTime, option.EndTime)
	dataS, err := GetQueryDataList(sqlClause)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	for _, data := range dataS {
		onlineTime := fmt.Sprintf("%.2f", utils.Str2Float64(data["online_time_max"])-utils.Str2Float64(data["online_time_min"]))
		hourIncome := fmt.Sprintf("%.2f", utils.Str2Float64(data["hour_income_max"])-utils.Str2Float64(data["hour_income_min"]))
		upstreamTraffic := fmt.Sprintf("%.2f", utils.Str2Float64(data["upstream_traffic_max"])-utils.Str2Float64(data["upstream_traffic_min"]))
		downstreamTraffic := fmt.Sprintf("%.2f", utils.Str2Float64(data["downstream_traffic_max"])-utils.Str2Float64(data["downstream_traffic_min"]))
		RetrieveCount := fmt.Sprintf("%d", utils.Str2Int64(data["retrieve_count_max"])-utils.Str2Int64(data["retrieve_count_min"]))
		data["online_time"] = onlineTime
		data["hour_income"] = hourIncome
		data["upstream_traffic"] = upstreamTraffic
		data["downstream_traffic"] = downstreamTraffic
		data["retrieve_count"] = RetrieveCount
		delete(data, "online_time_max")
		delete(data, "online_time_min")
		delete(data, "hour_income_max")
		delete(data, "hour_income_min")
		delete(data, "upstream_traffic_min")
		delete(data, "upstream_traffic_max")
		delete(data, "downstream_traffic_min")
		delete(data, "downstream_traffic_max")
		delete(data, "retrieve_count_min")
		delete(data, "retrieve_count_max")
		data["date"] = strings.Split(data["date"], " ")[1] + ":00"
	}
	return dataS, err
}

func GetDeviceInfoDailyList(ctx context.Context, cond *model.DeviceInfoDaily, option QueryOption) ([]*model.DeviceInfoDaily, error) {
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

	var out []*model.DeviceInfoDaily
	err := DB.SelectContext(ctx, &out, fmt.Sprintf(
		`SELECT * FROM %s %s`, tableNameDeviceInfoDaily, where), args...)
	if err != nil {
		return nil, err
	}

	return out, err
}

func GetIncomeAllList(ctx context.Context, cond *model.DeviceInfoDaily, option QueryOption) []map[string]interface{} {
	sqlClause := fmt.Sprintf("select date_format(time, '%%Y-%%m-%%d') as date, , sum(income) as income from device_info_daily "+
		"where device_id='%s' and time>='%s' and time<='%s' group by date", cond.DeviceID, option.StartTime, option.EndTime)
	if cond.UserID != "" {
		sqlClause = fmt.Sprintf("select date_format(time, '%%Y-%%m-%%d') as date, sum(income) as income from device_info_daily "+
			"where user_id='%s' and time>='%s' and time<='%s' group by date", cond.UserID, option.StartTime, option.EndTime)
	}
	dataS, err := GetQueryDataList(sqlClause)
	if err != nil {
		return nil
	}
	var mapIncomeList []map[string]interface{}
	for _, data := range dataS {
		mapIncome := make(map[string]interface{})
		mapIncome["date"] = data["date"]
		mapIncome["income"] = utils.StrToFloat(data["income"])
		mapIncomeList = append(mapIncomeList, mapIncome)
	}
	return mapIncomeList
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
		`SELECT * FROM %s %s LIMIT %d OFFSET %s`, tableNameDeviceInfoDaily, where, limit, offset), args...)
	if err != nil {
		return nil, 0, err
	}

	return out, total, err
}
