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
	tableNameHourDaily   = "hour_daily"
	tableNameIncomeDaily = "income_daily"
)

func GetHourDailyByTime(ctx context.Context, deviceID string, time time.Time) (*model.HourDaily, error) {
	var out model.HourDaily
	if err := DB.QueryRowxContext(ctx, fmt.Sprintf(
		`SELECT * FROM %s WHERE device_id = ? AND time = ?`, tableNameHourDaily),
		deviceID, time,
	).StructScan(&out); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}

func UpdateHourDaily(ctx context.Context, hourDaily *model.HourDaily) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`UPDATE %s SET updated_at = :updated_at, hour_income = :hour_income,
			online_time = :online_time, pkg_loss_ratio = :pkg_loss_ratio, latency = :latency,
			nat_ratio = :nat_ratio, disk_usage = :disk_usage, time = :time WHERE id = :id`, tableNameHourDaily),
		hourDaily)
	return err
}

func CreateHourDaily(ctx context.Context, hourDaily *model.HourDaily) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`INSERT INTO %s (created_at, updated_at, hour_income, user_id, device_id,
				online_time, pkg_loss_ratio, latency, nat_ratio, disk_usage, time)
			VALUES (:created_at, :updated_at, :hour_income, :user_id, :device_id, :online_time, :pkg_loss_ratio, :latency, 
			    :nat_ratio, :disk_usage, :time);`,
		tableNameHourDaily,
	), hourDaily)
	return err
}

func GetIncomeDailyByTime(ctx context.Context, deviceID string, time time.Time) (*model.IncomeDaily, error) {
	var out model.IncomeDaily
	if err := DB.QueryRowxContext(ctx, fmt.Sprintf(
		`SELECT * FROM %s WHERE device_id =? AND time = ?`, tableNameIncomeDaily),
		deviceID, time,
	).StructScan(&out); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}

func CreateIncomeDaily(ctx context.Context, daily *model.IncomeDaily) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`INSERT INTO %s (created_at, updated_at, income, user_id, device_id,
				online_time, pkg_loss_ratio, latency, nat_ratio, disk_usage, time)
			VALUES (:created_at, :updated_at, :income, :user_id, :device_id,
				:online_time, :pkg_loss_ratio, :latency, :nat_ratio, :disk_usage, :time);`,
		tableNameIncomeDaily,
	), daily)
	return err
}

func UpdateIncomeDaily(ctx context.Context, daily *model.IncomeDaily) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`UPDATE %s SET updated_at = :updated_at, deleted_at = :deleted_at, income = :income,
			online_time = :online_time, pkg_loss_ratio = :pkg_loss_ratio, latency = :latency,
			nat_ratio = :nat_ratio, disk_usage = :disk_usage, time = :time WHERE id = :id`, tableNameIncomeDaily),
		daily)
	return err
}

func GetIncomeDailyHourList(ctx context.Context, cond *model.HourDaily, option QueryOption) ([]*model.HourDaily, int64, error) {
	var args []interface{}
	where := `WHERE 1=1`
	if cond.DeviceID != "" {
		where += ` AND device_id = ?`
		args = append(args, cond.DeviceID)
	}
	if cond.UserID != "" {
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

	limit := option.PageSize
	offset := option.Page
	if option.PageSize <= 0 {
		limit = 50
	}
	if option.Page > 0 {
		offset = limit * (option.Page - 1)
	}

	var total int64
	var out []*model.HourDaily

	err := DB.GetContext(ctx, &total, fmt.Sprintf(
		`SELECT count(*) FROM %s %s`, tableNameHourDaily, where,
	), args...)
	if err != nil {
		return nil, 0, err
	}

	err = DB.SelectContext(ctx, &out, fmt.Sprintf(
		`SELECT * FROM %s %s LIMIT %d OFFSET %d`, tableNameHourDaily, where, limit, offset,
	), args...)
	if err != nil {
		return nil, 0, err
	}

	return out, total, err
}

func GetIncomeDailyList(ctx context.Context, cond *model.IncomeDaily, option QueryOption) ([]*model.IncomeDaily, int64, error) {
	var args []interface{}
	where := `WHERE 1=1`
	if cond.DeviceID != "" {
		where += ` AND device_id = ?`
		args = append(args, cond.DeviceID)
	}
	if cond.UserID != "" {
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

	limit := option.PageSize
	offset := option.Page
	if option.PageSize <= 0 {
		limit = 50
	}
	if option.Page > 0 {
		offset = limit * (option.Page - 1)
	}

	var total int64
	var out []*model.IncomeDaily

	err := DB.GetContext(ctx, &total, fmt.Sprintf(
		`SELECT count(*) FROM %s %s`, tableNameIncomeDaily, where,
	), args...)
	if err != nil {
		return nil, 0, err
	}

	err = DB.SelectContext(ctx, &out, fmt.Sprintf(
		`SELECT * FROM %s %s LIMIT %d OFFSET %d`, tableNameIncomeDaily, where, limit, offset,
	), args...)
	if err != nil {
		return nil, 0, err
	}

	return out, total, err
}

func GetIncomeAllList(ctx context.Context, cond *model.IncomeDaily, option QueryOption) []map[string]interface{} {
	sqlClause := fmt.Sprintf("select date_format(time, '%%Y-%%m-%%d') as date, , sum(income) as income from income_daily "+
		"where device_id='%s' and time>='%s' and time<='%s' group by date", cond.DeviceID, option.StartTime, option.EndTime)
	if cond.UserID != "" {
		sqlClause = fmt.Sprintf("select date_format(time, '%%Y-%%m-%%d') as date, sum(income) as income from income_daily "+
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
