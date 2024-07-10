package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"time"
)

const (
	EligibleTopOnlineTimePercent = 10
	IncentiveRewardPercent       = 10
)

func IsGeneratedOnlineIncentive(ctx context.Context, date time.Time) (bool, error) {
	query := `select * from device_online_incentive where date >= ? limit 1`

	var incentive model.DeviceOnlineIncentive
	err := DB.GetContext(ctx, &incentive, query, date)

	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func GenerateEligibleOnlineDevices(ctx context.Context) error {
	countQuery := `select count(device_id) from device_info where yesterday_online_time > 0 and node_type = 1`

	var count int64
	err := DB.GetContext(ctx, &count, countQuery)
	if err != nil {
		return err
	}

	limit := count * EligibleTopOnlineTimePercent / 100
	if limit <= 0 {
		limit = 1
	}

	queryOnlineTime := `select ifnull(min(yesterday_online_time), 0) from (
				select yesterday_online_time from device_info where yesterday_online_time > 0 and node_type = 1 order by yesterday_online_time desc limit ? 
		)t`

	var onlineTime int64
	err = DB.GetContext(ctx, &onlineTime, queryOnlineTime, limit)
	if err != nil {
		return err
	}

	tx, err := DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := fmt.Sprintf(`insert into device_online_incentive 
	select device_id, user_id, if(yesterday_profit > 0, yesterday_profit/%d, 0) as reward, yesterday_online_time as online_time, date_sub(curdate(), interval 1 day) as date, now() as created_at 
	from device_info where yesterday_online_time >= ? and node_type = 1`, IncentiveRewardPercent)

	_, err = tx.ExecContext(ctx, query, onlineTime)
	if err != nil {
		return err
	}

	update := fmt.Sprintf(`update device_info set online_incentive_profit = online_incentive_profit +  yesterday_profit/%d where device_id in ( 
            select device_id from device_online_incentive where date >= date_sub(curdate(), interval 1 day))`, IncentiveRewardPercent)

	_, err = tx.ExecContext(ctx, update)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func GetDeviceOnlineIncentiveList(ctx context.Context, deviceId string, option QueryOption) ([]*model.DeviceOnlineIncentive, int64, error) {
	limit := option.PageSize
	offset := option.Page
	if option.PageSize <= 0 {
		limit = 50
	}
	if option.Page > 0 {
		offset = limit * (option.Page - 1)
	}

	var total int64
	var out []*model.DeviceOnlineIncentive

	err := DB.GetContext(ctx, &total, `SELECT count(*) FROM device_online_incentive where device_id = ? `, deviceId)
	if err != nil {
		return nil, 0, err
	}

	err = DB.SelectContext(ctx, &out, `SELECT * FROM  device_online_incentive where device_id = ? ORDER BY date desc LIMIT ? OFFSET ?`, deviceId, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return out, total, err
}
