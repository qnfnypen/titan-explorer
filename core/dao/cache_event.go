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
	query := fmt.Sprintf(`SELECT * FROM %s ORDER BY time DESC LIMIT 1;`, tableNameCacheEvent)
	err := DB.QueryRowxContext(ctx, query).StructScan(&out)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return &out, nil
}
