// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0
// source: query.sql

package model

import (
	"context"
)

const getDeviceInfo = `-- name: GetDeviceInfo :one
SELECT id, created_at, updated_at, deleted_at, device_id, scheduler_id, secret, node_type, ` + "`" + `rank` + "`" + `, device_name, user_id, sn_code, operator, network_type, system_version, product_type, network_info, external_ip, internal_ip, ip_location, ip_country, ip_city, mac_location, nat_type, upnp, pkg_loss_ratio, nat_ratio, latency, cpu_usage, cpu_cores, memory_usage, memory, disk_usage, disk_space, work_status, device_status, disk_type, io_system, online_time, today_online_time, today_profit, yesterday_profit, seven_days_profit, month_profit, cumu_profit, bandwidth_up, bandwidth_down FROM ` + "`" + `device_info` + "`" + ` WHERE device_id = ? LIMIT 1
`

func (q *Queries) GetDeviceInfo(ctx context.Context, db DBTX, deviceID string) (DeviceInfo, error) {
	row := db.QueryRowContext(ctx, getDeviceInfo, deviceID)
	var i DeviceInfo
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
		&i.DeviceID,
		&i.SchedulerID,
		&i.Secret,
		&i.NodeType,
		&i.Rank,
		&i.DeviceName,
		&i.UserID,
		&i.SnCode,
		&i.Operator,
		&i.NetworkType,
		&i.SystemVersion,
		&i.ProductType,
		&i.NetworkInfo,
		&i.ExternalIp,
		&i.InternalIp,
		&i.IpLocation,
		&i.IpCountry,
		&i.IpCity,
		&i.MacLocation,
		&i.NatType,
		&i.Upnp,
		&i.PkgLossRatio,
		&i.NatRatio,
		&i.Latency,
		&i.CpuUsage,
		&i.CpuCores,
		&i.MemoryUsage,
		&i.Memory,
		&i.DiskUsage,
		&i.DiskSpace,
		&i.WorkStatus,
		&i.DeviceStatus,
		&i.DiskType,
		&i.IoSystem,
		&i.OnlineTime,
		&i.TodayOnlineTime,
		&i.TodayProfit,
		&i.YesterdayProfit,
		&i.SevenDaysProfit,
		&i.MonthProfit,
		&i.CumuProfit,
		&i.BandwidthUp,
		&i.BandwidthDown,
	)
	return i, err
}
