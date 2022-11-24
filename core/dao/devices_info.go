package dao

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

var tableNameDeviceInfo = "device_info"

func GetDeviceInfoList(ctx context.Context, cond *model.DeviceInfo, option QueryOption) ([]*model.DeviceInfo, int64, error) {
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
	if cond.DeviceStatus != "" && cond.DeviceStatus != "allDevices" {
		where += ` AND device_status = ?`
		args = append(args, cond.DeviceStatus)
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
	var out []*model.DeviceInfo

	err := DB.GetContext(ctx, &total, fmt.Sprintf(
		`SELECT count(*) FROM %s %s`, tableNameDeviceInfo, where,
	), args...)
	if err != nil {
		return nil, 0, err
	}

	err = DB.SelectContext(ctx, &out, fmt.Sprintf(
		`SELECT * FROM %s %s LIMIT %d OFFSET %d`, tableNameDeviceInfo, where, limit, offset,
	), args...)
	if err != nil {
		return nil, 0, err
	}

	return out, total, err
}

func GetDeviceInfoByID(ctx context.Context, deviceID string) (*model.DeviceInfo, error) {
	var out model.DeviceInfo
	if err := DB.QueryRowxContext(ctx, fmt.Sprintf(
		`SELECT * FROM %s WHERE device_id = ?`, tableNameDeviceInfo), deviceID,
	).StructScan(&out); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}

func UpdateDeviceInfo(ctx context.Context, deviceInfo *model.DeviceInfo) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`UPDATE %s SET  node_type = :node_type,  device_name = :device_name,
				sn_code = :sn_code,  operator = :operator, network_type = :network_type,
				today_income = :today_income,  yesterday_income = :yesterday_income,  cumu_profit = :cumu_profit,
				system_version = :system_version,  product_type = :product_type, network_info = :network_info,
				external_ip = :external_ip,  internal_ip = :internal_ip,  ip_location = :ip_location,  
				mac_location = :mac_location,  nat_type = :nat_type,  upnp = :upnp, pkg_loss_ratio = :pkg_loss_ratio,  
				nat_ratio = :nat_ratio,  latency = :latency,  cpu_usage = :cpu_usage,  memory_usage = :memory_usage,
				disk_usage = :disk_usage,  work_status = :work_status, device_status = :device_status,  disk_type = :disk_type,
				io_system = :io_system, online_time = :online_time, today_online_time = :today_online_time,  today_profit = :today_profit,
				seven_days_profit = :seven_days_profit, month_profit = :month_profit,  bandwidth_up = :bandwidth_up,  
				bandwidth_down = :bandwidth_down, updated_at = :updated_at WHERE device_id = :device_id`, tableNameDeviceInfo),
		deviceInfo)
	return err
}

func CreateDeviceInfo(ctx context.Context, deviceInfo *model.DeviceInfo) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`INSERT INTO %s (device_id, secret, node_type, device_name, user_id, sn_code, operator,
				network_type, today_income, yesterday_income, cumu_profit, system_version, product_type,
				network_info, external_ip, internal_ip, ip_location, mac_location, nat_type, upnp,
				pkg_loss_ratio, nat_ratio, latency, cpu_usage, memory_usage, disk_usage, work_status,
				device_status, disk_type, io_system, online_time, today_online_time, today_profit, seven_days_profit,
				month_profit, bandwidth_up, bandwidth_down, created_at, updated_at, deleted_at)
			VALUES (:device_id, :secret, :node_type, :device_name, :user_id, :sn_code, :operator,
			    :network_type, :today_income, :yesterday_income, :cumu_profit, :system_version, :product_type, 
			    :network_info, :external_ip, :internal_ip, :ip_location, :mac_location, :nat_type, :upnp, 
			    :pkg_loss_ratio, :nat_ratio, :latency, :cpu_usage, :memory_usage, :disk_usage, :work_status, 
			    :device_status, :disk_type, :io_system, :online_time, :today_online_time, :today_profit, :seven_days_profit, 
			    :month_profit, :bandwidth_up, :bandwidth_down, :created_at, :updated_at, :deleted_at);`, tableNameDeviceInfo,
	), deviceInfo)
	return err
}
