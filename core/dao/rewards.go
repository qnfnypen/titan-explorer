package dao

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"strconv"
	"time"
)

var (
	tableNameDeviceInfoHour  = "device_info_hour"
	tableNameDeviceInfoDaily = "device_info_daily"
)

func GetDeviceInfoHourByTime(ctx context.Context, deviceID string, time time.Time) (*model.DeviceInfoHour, error) {
	var out model.DeviceInfoHour
	if err := DB.QueryRowxContext(ctx, fmt.Sprintf(
		`SELECT * FROM %s WHERE device_id = ? AND time = ?`, tableNameDeviceInfoHour),
		deviceID, time,
	).StructScan(&out); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}

func UpdateDeviceInfoHour(ctx context.Context, deviceInfoHour *model.DeviceInfoHour) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`UPDATE %s SET updated_at = :updated_at, hour_income = :hour_income,
			online_time = :online_time, pkg_loss_ratio = :pkg_loss_ratio, latency = :latency,
			nat_ratio = :nat_ratio, disk_usage = :disk_usage, time = :time WHERE id = :id`, tableNameDeviceInfoHour),
		deviceInfoHour)
	return err
}

func CreateDeviceInfoHour(ctx context.Context, deviceInfoHour *model.DeviceInfoHour) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`INSERT INTO %s (created_at, updated_at, hour_income, user_id, device_id,
				online_time, pkg_loss_ratio, latency, nat_ratio, disk_usage, time)
			VALUES (:created_at, :updated_at, :hour_income, :user_id, :device_id, :online_time, :pkg_loss_ratio, :latency, 
			    :nat_ratio, :disk_usage, :time);`,
		tableNameDeviceInfoHour,
	), deviceInfoHour)
	return err
}

func GetDeviceInfoDailyByTime(ctx context.Context, deviceID string, time time.Time) (*model.DeviceInfoDaily, error) {
	var out model.DeviceInfoDaily
	if err := DB.QueryRowxContext(ctx, fmt.Sprintf(
		`SELECT * FROM %s WHERE device_id =? AND time = ?`, tableNameDeviceInfoDaily),
		deviceID, time,
	).StructScan(&out); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}

func CreateDeviceInfoDaily(ctx context.Context, daily *model.DeviceInfoDaily) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`INSERT INTO %s (created_at, updated_at, income, user_id, device_id,
				online_time, pkg_loss_ratio, latency, nat_ratio, disk_usage, time)
			VALUES (:created_at, :updated_at, :income, :user_id, :device_id,
				:online_time, :pkg_loss_ratio, :latency, :nat_ratio, :disk_usage, :time);`,
		tableNameDeviceInfoDaily,
	), daily)
	return err
}

func UpdateDeviceInfoDaily(ctx context.Context, daily *model.DeviceInfoDaily) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`UPDATE %s SET updated_at = :updated_at, deleted_at = :deleted_at, income = :income,
			online_time = :online_time, pkg_loss_ratio = :pkg_loss_ratio, latency = :latency,
			nat_ratio = :nat_ratio, disk_usage = :disk_usage, time = :time WHERE id = :id`, tableNameDeviceInfoDaily),
		daily)
	return err
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
