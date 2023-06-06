package dao

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

const tableNameCacheEvent = "cache_event"

func CreateCacheEvents(ctx context.Context, events []*model.CacheEvent) error {
	query := fmt.Sprintf(`INSERT INTO %s(device_id, carfile_cid,block_size,status, blocks, replicaInfos ,time) 
			VALUES(:device_id, :carfile_cid, :block_size, :status, :blocks, :replicaInfos, :time) ON DUPLICATE KEY UPDATE blocks = VALUES(blocks),
	block_size = VALUES(block_size),status = VALUES(status),replicaInfos = VALUES(replicaInfos),time = VALUES(time),updated_at = now()`, tableNameCacheEvent)

	_, err := DB.NamedExecContext(ctx, query, events)
	return err
}

func ResetCacheEvents(ctx context.Context, carFileCid string) error {
	_, err := DB.DB.ExecContext(ctx, fmt.Sprintf(
		`UPDATE %s SET status = 4, updated_at = now() WHERE carfile_cid = '%s'`, tableNameCacheEvent, carFileCid))
	return err
}

func GetReplicaInfo(ctx context.Context) model.FullNodeInfo {
	var ReplicaInfo model.FullNodeInfo
	query := fmt.Sprintf("SELECT count(*) as t_upstream_file_count,ROUND(count(*)/count(distinct(carfile_cid)),2) AS t_average_replica FROM %s where status = 3 ", tableNameCacheEvent)
	err := DB.QueryRowxContext(ctx, query).StructScan(&ReplicaInfo)
	if err != nil {
		log.Errorf("getDeviceInfo %v", err)
	}
	return ReplicaInfo
}

func GetLastCacheEvent(ctx context.Context) (*model.CacheEvent, error) {
	var out model.CacheEvent
	query := fmt.Sprintf(`SELECT * FROM %s ORDER BY time DESC LIMIT 1;`, tableNameCacheEvent)
	err := DB.QueryRowxContext(ctx, query).StructScan(&out)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func CountCacheEvent(ctx context.Context, nodeId string) error {
	var out model.DeviceInfo
	query := fmt.Sprintf(`SELECT device_id,count(DISTINCT(carfile_cid)) as block_count FROM %s where status = 3 and device_id = '%s' GROUP BY device_id;`, tableNameCacheEvent, nodeId)
	err := DB.QueryRowxContext(ctx, query).StructScan(&out)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	err = UpdateCarFileCount(ctx, &out)
	if err != nil {
		return err
	}
	return nil
}

func GetCacheEventsByPage(ctx context.Context, cond *model.CacheEvent, option QueryOption) ([]*model.CacheEvent, int64, error) {
	var args []interface{}
	where := `WHERE 1=1`
	if cond.DeviceID != "" {
		where += ` AND device_id = ?`
		args = append(args, cond.DeviceID)
	}

	if option.Order != "" && option.OrderField != "" {
		where += fmt.Sprintf(` ORDER BY %s %s`, option.OrderField, option.Order)
	}

	if option.Order != "" && option.OrderField != "" {
		where += fmt.Sprintf(` ORDER BY %s %s`, option.OrderField, option.Order)
	} else {
		where += " ORDER BY time DESC"
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
	var out []*model.CacheEvent

	err := DB.GetContext(ctx, &total, fmt.Sprintf(
		`SELECT count(*) FROM %s %s`, tableNameCacheEvent, where,
	), args...)
	if err != nil {
		return nil, 0, err
	}

	err = DB.SelectContext(ctx, &out, fmt.Sprintf(
		`SELECT * FROM %s %s LIMIT %d OFFSET %d`, tableNameCacheEvent, where, limit, offset,
	), args...)
	if err != nil {
		return nil, 0, err
	}

	return out, total, err
}
