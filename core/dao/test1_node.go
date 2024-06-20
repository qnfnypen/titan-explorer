package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

const (
	test1NodeTable = "device_info"
	zeroTime       = "0000-00-00 00:00:00.000"
)

// GetTest1Nodes 获取test1节点信息
func GetTest1Nodes(ctx context.Context, statusCode int64, page, size uint64) (int64, []model.Test1NodeInfo, error) {
	// device_status_code 1-在线 2-故障 3-离线
	var (
		totalBuilder squirrel.SelectBuilder
		infoBuilder  squirrel.SelectBuilder

		total int64
		infos = make([]model.Test1NodeInfo, 0)
	)

	if statusCode <= 0 || statusCode > 4 {
		return 0, nil, errors.New("param error")
	}

	// 获取删除节点
	if statusCode == 4 {
		totalBuilder = squirrel.Select("COUNT(device_id)").From(test1NodeTable).Where("deleted_at <> ?", 0)
		infoBuilder = squirrel.Select("device_name,external_ip,system_version,device_id,ip_location,cumulative_profit").From(test1NodeTable).
			Where("deleted_at <> ?", 0).Offset((page - 1) * size).Limit(size)
	} else {
		totalBuilder = squirrel.Select("COUNT(device_id)").From(test1NodeTable).Where("device_status_code = ?", statusCode)
		infoBuilder = squirrel.Select("device_name,external_ip,system_version,device_id,ip_location,cumulative_profit").From(test1NodeTable).
			Where("device_status_code = ?", statusCode).Offset((page - 1) * size).Limit(size)
	}

	query, args, err := totalBuilder.ToSql()
	if err != nil {
		return 0, nil, fmt.Errorf("generate sql of get total of device's info error:%w", err)
	}
	err = DB.GetContext(ctx, &total, query, args...)
	if err != nil {
		return 0, nil, fmt.Errorf("get total of device's info error:%w", err)
	}

	query, args, err = infoBuilder.ToSql()
	if err != nil {
		return 0, nil, fmt.Errorf("generate sql of get device's info error:%w", err)
	}
	err = DB.SelectContext(ctx, &infos, query, args...)
	if err != nil {
		return 0, nil, fmt.Errorf(" get device's info error:%w", err)
	}

	return total, infos, nil
}

// GetNodeNums 获取节点总数
func GetNodeNums(ctx context.Context, usedID string) (int64, int64, int64, int64, error) {
	var online, abnormal, offline, deleted int64
	// device_status_code 1-在线 2-故障 3-离线
	// 在线
	query, args, err := squirrel.Select("COUNT(device_id)").From(test1NodeTable).Where("device_status_code = 1 AND user_id = ? AND deleted_at = ?", usedID, zeroTime).ToSql()
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("get total of device's info error:%w", err)
	}
	err = DB.GetContext(ctx, &online, query, args...)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("get total of device's info error:%w", err)
	}
	// 故障
	query, args, err = squirrel.Select("COUNT(device_id)").From(test1NodeTable).Where("device_status_code = 2 AND user_id = ? AND deleted_at = ?", usedID, zeroTime).ToSql()
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("get total of device's info error:%w", err)
	}
	err = DB.GetContext(ctx, &abnormal, query, args...)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("get total of device's info error:%w", err)
	}
	// 离线
	query, args, err = squirrel.Select("COUNT(device_id)").From(test1NodeTable).Where("device_status_code = 3 AND user_id = ? AND deleted_at = ?", usedID, zeroTime).ToSql()
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("get total of device's info error:%w", err)
	}
	err = DB.GetContext(ctx, &offline, query, args...)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("get total of device's info error:%w", err)
	}
	// 删除
	query, args, err = squirrel.Select("COUNT(device_id)").From(test1NodeTable).Where("deleted_at <> 0 AND user_id = ?", usedID).ToSql()
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("get total of device's info error:%w", err)
	}
	err = DB.GetContext(ctx, &deleted, query, args...)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("get total of device's info error:%w", err)
	}

	return online, abnormal, offline, deleted, nil
}

// UpdateTest1DeviceName 编辑节点设备备注
func UpdateTest1DeviceName(ctx context.Context, id, name string) error {
	query, args, err := squirrel.Update(test1NodeTable).Set("device_name", name).Where("device_id = ?", id).ToSql()
	if err != nil {
		return fmt.Errorf("generate update device's name error:%w", err)
	}

	_, err = DB.DB.ExecContext(ctx, query, args...)
	switch err {
	case sql.ErrNoRows:
	case nil:
	default:
		return fmt.Errorf("update device's name error:%w", err)
	}

	return nil
}

// DeleteOfflineDevice 删除离线设备
func DeleteOfflineDevice(ctx context.Context, ids []string, usedID string) error {
	query, args, err := squirrel.Update(test1NodeTable).Set("deleted_at", time.Now()).Where(squirrel.Eq{
		"device_id":          ids,
		"user_id":            usedID,
		"device_status_code": 3,
		"deleted_at":         zeroTime,
	}).ToSql()
	if err != nil {
		return fmt.Errorf("generate delete offline's device error:%w", err)
	}

	_, err = DB.DB.ExecContext(ctx, query, args...)
	switch err {
	case sql.ErrNoRows:
	case nil:
	default:
		return fmt.Errorf("delete offline's device error:%w", err)
	}

	return nil
}

// MoveBackDeletedDevice 移回删除的设备
func MoveBackDeletedDevice(ctx context.Context, ids []string, usedID string) error {
	query, args, err := squirrel.Update(test1NodeTable).Set("deleted_at", zeroTime).Where("deleted_at <> 0").Where(squirrel.Eq{
		"device_id": ids,
		"user_id":   usedID,
	}).ToSql()
	if err != nil {
		return fmt.Errorf("generate move back deleted device error:%w", err)
	}

	_, err = DB.DB.ExecContext(ctx, query, args...)
	switch err {
	case sql.ErrNoRows:
	case nil:
	default:
		return fmt.Errorf("move back deleted device error:%w", err)
	}

	return nil
}
