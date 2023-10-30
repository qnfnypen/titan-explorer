package dao

import (
	"context"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

var tableNameStorageStats = "storage_stats"

func AddStorageStats(ctx context.Context, stats []*model.StorageStats) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(`
		INSERT INTO %s (s_rank, project_id, project_name, total_size, user_count, provider_count, storage_change_24h, storage_change_percentage_24h, time, expiration, gas, pledge, locations, created_at, updated_at)
			VALUES (:s_rank, :project_id, :project_name, :total_size, :user_count, :provider_count, :storage_change_24h, :storage_change_percentage_24h, :time, :expiration, :gas, :pledge, :locations, :created_at, :updated_at)
		 ON DUPLICATE KEY UPDATE s_rank=VALUES(s_rank), total_size=VALUES(total_size), user_count=VALUES(user_count), provider_count=VALUES(provider_count), storage_change_24h=VALUES(storage_change_24h), storage_change_percentage_24h=VALUES(storage_change_percentage_24h),
		expiration=VALUES(expiration), gas=VALUES(gas), pledge=VALUES(pledge), locations=VALUES(locations)`, tableNameStorageStats,
	), stats)
	return err
}

func CountStorageStats(ctx context.Context) (*model.StorageSummary, error) {
	var (
		users       int64
		providers   int64
		lastOneTime string
		out         model.StorageSummary
	)

	if err := DB.GetContext(ctx, &lastOneTime, fmt.Sprintf(`select time from %s order by id desc limit 1`, tableNameStorageStats)); err != nil {
		return nil, err
	}

	queryStatement := fmt.Sprintf(`select count(DISTINCT project_id) projects, sum(total_size) as total_size, sum(provider_count) as providers, sum(user_count) as users, 
    	sum(pledge) as pledges, sum(gas) as gases from %s where time=?;`, tableNameStorageStats)

	err := DB.GetContext(ctx, &out, queryStatement, lastOneTime)
	if err != nil {
		return nil, err
	}

	if err := DB.GetContext(ctx, &users, fmt.Sprintf(`select count(distinct user_id) from %s`, tableNameAsset)); err != nil {
		return nil, err
	}

	if err := DB.GetContext(ctx, &providers, fmt.Sprintf(`select count(distinct provider) from %s`, tableNameFilStorage)); err != nil {
		return nil, err
	}

	out.Providers = providers
	out.Users = users
	out.LatestUpdateTime = lastOneTime

	return &out, nil
}

func GetLastStorageStats(ctx context.Context) (*model.StorageStats, error) {
	var out model.StorageStats
	if err := DB.GetContext(ctx, &out, fmt.Sprintf(`select * from %s order by time desc limit 1`, tableNameStorageStats)); err != nil {
		return nil, err
	}
	return &out, nil
}

func ListStorageStats(ctx context.Context, projectId int64, opts QueryOption) ([]*model.StorageStats, int64, error) {
	var args []interface{}
	conditionStatement := ` where 1=1`
	if projectId >= 0 {
		conditionStatement += ` and project_id = ?`
		args = append(args, projectId)
	}

	if opts.StartTime != "" {
		conditionStatement += ` and time >= ?`
		args = append(args, opts.StartTime)
	}

	if opts.EndTime != "" {
		conditionStatement += ` and time < ?`
		args = append(args, opts.EndTime)
	}

	var total int64
	err := DB.GetContext(ctx, &total, fmt.Sprintf(`select count(1) from (select * from %s %s group by project_id ) as a`, tableNameStorageStats, conditionStatement), args...)
	if err != nil {
		return nil, 0, err
	}

	queryStatement := `select * from (select * from %s %s group by project_id) a %s limit ? offset ?`

	orderStatement := ""
	if opts.OrderField == "rank" {
		opts.OrderField = "s_rank"
	}

	if opts.Order != "" && opts.OrderField != "" {
		orderStatement = fmt.Sprintf(` ORDER BY %s %s`, opts.OrderField, opts.Order)
	}

	limit := opts.PageSize
	offset := opts.Page
	if opts.PageSize <= 0 {
		limit = 50
	}
	if opts.Page > 0 {
		offset = limit * (opts.Page - 1)
	}

	args = append(args, limit, offset)

	var out []*model.StorageStats
	err = DB.SelectContext(ctx, &out, fmt.Sprintf(queryStatement, tableNameStorageStats, conditionStatement, orderStatement), args...)

	if err != nil {
		return nil, 0, err
	}

	return out, total, err
}
