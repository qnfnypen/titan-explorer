package dao

import (
	"context"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"strconv"
)

var (
	tableNameDeviceInfoHour  = "device_info_hour"
	tableNameDeviceInfoDaily = "device_info_daily"
)

func BulkUpsertDeviceInfoHours(ctx context.Context, hourInfos []*model.DeviceInfoHour) error {
	upsertStatement := fmt.Sprintf(`INSERT INTO %s (created_at, updated_at, hour_income, user_id, device_id,
				online_time, pkg_loss_ratio, latency, nat_ratio, disk_usage, time)
			VALUES (:created_at, :updated_at, :hour_income, :user_id, :device_id, :online_time, :pkg_loss_ratio, :latency, 
			    :nat_ratio, :disk_usage, :time) 
			 ON DUPLICATE KEY UPDATE updated_at = now(), hour_income = :hour_income,
			online_time = :online_time, pkg_loss_ratio = :pkg_loss_ratio, latency = :latency,
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
				online_time, pkg_loss_ratio, latency, nat_ratio, disk_usage, time)
			VALUES (:created_at, :updated_at, :income, :user_id, :device_id,
				:online_time, :pkg_loss_ratio, :latency, :nat_ratio, :disk_usage, :time) 
			 ON DUPLICATE KEY UPDATE updated_at = now(), income = :income,
			online_time = :online_time, pkg_loss_ratio = :pkg_loss_ratio, latency = :latency,
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

func GetDeviceInfoDailyHourList(ctx context.Context, cond *model.DeviceInfoHour, option QueryOption) ([]*model.DeviceInfoHour, error) {
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

	var out []*model.DeviceInfoHour
	err := DB.SelectContext(ctx, &out, fmt.Sprintf(
		`SELECT * FROM %s %s`, tableNameDeviceInfoHour, where), args...)
	if err != nil {
		return nil, err
	}

	return out, err
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
	datas, err := GetQueryDataList(sqlClause)
	if err != nil {
		return nil
	}
	var mapIncomeList []map[string]interface{}
	for _, data := range datas {
		mapIncome := make(map[string]interface{})
		mapIncome["date"] = data["date"]
		mapIncome["income"], _ = strconv.ParseFloat(data["income"], 10)
		mapIncomeList = append(mapIncomeList, mapIncome)
	}
	return mapIncomeList
}
