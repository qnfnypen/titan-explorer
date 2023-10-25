package dao

import (
	"context"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"time"
)

var tableNameStorageStats = "storage_stats"

func AddStorageStats(ctx context.Context, stats []*model.StorageStats) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(`
		INSERT INTO %s ( project_id, project_name, total_size, user_count, provider_count, storage_change_24h, storage_change_percentage_24h, time, expiration, gas, pledge, locations, created_at, updated_at)
			VALUES ( :project_id, :project_name, :total_size, :user_count, :provider_count, :storage_change_24h, :storage_change_percentage_24h, :time, :expiration, :gas, :pledge, :locations, :created_at, :updated_at)`, tableNameStorageStats,
	), stats)
	return err
}

func CountStorageStats(ctx context.Context) (*model.StorageSummary, error) {
	var lastOneTime time.Time
	if err := DB.GetContext(ctx, &lastOneTime, fmt.Sprintf(`select time from %s order by id desc limit 1`, tableNameStorageStats)); err != nil {
		return nil, err
	}

	queryStatement := `select count(DISTINCT project_id) projects, sum(total_size) as storage_size, sum(provider_count) as providers, sum(user_count) as users, 
    	sum(pledge) as pledges, sum(gas) as gases from %s where time=?;`

	var out model.StorageSummary
	err := DB.GetContext(ctx, &out, queryStatement, lastOneTime)
	if err != nil {
		return nil, err
	}

	return &out, nil
}

func ListStorageStats(ctx context.Context, opts QueryOption) ([]*model.StorageStats, int64, error) {
	countStatement := `select count(1) from %s where time > '%s'  and time < '%s' group by project_id `

	var total int64
	err := DB.GetContext(ctx, &total, fmt.Sprintf(countStatement, tableNameStorageStats, opts.StartTime, opts.EndTime))
	if err != nil {
		return nil, 0, err
	}

	queryStatement := `select project_id, project_name, total_size, user_count, provider_count, expiration, gas, pledge, max(total_size) - min(total_size) as storage_change_24h, 
       	(max(total_size) - min(total_size))/total_size as storage_change_percentage_24h  from %s where time > '%s' and time < '%s' group by project_id limit ? offset ?`

	limit := opts.PageSize
	offset := opts.Page
	if opts.PageSize <= 0 {
		limit = 50
	}
	if opts.Page > 0 {
		offset = limit * (opts.Page - 1)
	}

	var out []*model.StorageStats
	err = DB.SelectContext(ctx, &out, fmt.Sprintf(queryStatement, tableNameStorageStats, opts.StartTime, opts.EndTime), limit, offset)

	if err != nil {
		return nil, 0, err
	}

	return out, total, err
}
