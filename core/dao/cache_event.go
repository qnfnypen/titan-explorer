package dao

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

const tableNameCacheEvent = "cache_event"

func GetLastCacheEvent(ctx context.Context) (*model.CacheEvent, error) {
	var out model.CacheEvent
	query := fmt.Sprintf(`SELECT * FROM %s ORDER BY max_created_time DESC LIMIT 1;`, tableNameCacheEvent)
	err := DB.QueryRowxContext(ctx, query).StructScan(&out)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &out, nil
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
		`SELECT * FROM %s %s LIMIT %d OFFSET %d ORDER BY time DESC`, tableNameCacheEvent, where, limit, offset,
	), args...)
	if err != nil {
		return nil, 0, err
	}

	return out, total, err
}
