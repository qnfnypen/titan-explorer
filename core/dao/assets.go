package dao

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/gnasnik/titan-explorer/core/generated/model"
)

const (
	tableNameAsset = "assets"
)

func AddAssets(ctx context.Context, assets []*model.Asset) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`INSERT INTO %s ( node_id, backup_result, cid, hash, total_size, end_time, expiration, user_id, edge_replicas, candidate_replicas, total_blocks, created_time, area_id, state, note, bandwidth, source, retry_count, replenish_replicas, failed_count, succeeded_count,
                client_ip)
			VALUES ( :node_id, :backup_result, :cid, :hash, :total_size, :end_time, :expiration, :user_id, :edge_replicas, :candidate_replicas, :total_blocks, :created_time, :area_id, :state, :note, :bandwidth, :source, :retry_count, :replenish_replicas, :failed_count, :succeeded_count,
				:client_ip) 
			ON DUPLICATE KEY UPDATE  backup_result = VALUES(backup_result), end_time = VALUES(end_time), expiration = VALUES(expiration), user_id = VALUES(user_id), edge_replicas = VALUES(edge_replicas), candidate_replicas = VALUES(candidate_replicas), 
			total_blocks = values(total_blocks), area_id = values(area_id), state = values(state), note = values(note), bandwidth = values(bandwidth), source = values(source), retry_count = values(retry_count), replenish_replicas = values(replenish_replicas), 
			failed_count = values(failed_count), succeeded_count = values(succeeded_count);`, tableNameAsset,
	), assets)
	return err
}

func UpdateAssetPath(ctx context.Context, cid string, path string) error {
	_, err := DB.ExecContext(ctx, fmt.Sprintf(
		`UPDATE %s SET path = ? where cid = ?`, tableNameAsset), path, cid)
	return err
}

func UpdateAssetBackupResult(ctx context.Context, cid string, backupResult int) error {
	_, err := DB.ExecContext(ctx, fmt.Sprintf(
		`UPDATE %s SET backup_result = ? where cid = ?`, tableNameAsset), backupResult, cid)
	return err
}

func GetLatestAsset(ctx context.Context) (*model.Asset, error) {
	var asset model.Asset
	err := DB.GetContext(ctx, &asset, fmt.Sprintf(
		`SELECT * from %s ORDER BY end_time DESC LIMIT 1`, tableNameAsset))
	if err != nil {
		return nil, err
	}
	return &asset, err
}

func CountAssets(ctx context.Context) (int64, error) {
	var count int64
	err := DB.GetContext(ctx, &count, fmt.Sprintf(
		`select count(cid) from %s`, tableNameAsset))
	if err != nil {
		return 0, err
	}
	return count, err
}

func GetAssetsByEmptyPath(ctx context.Context) ([]*model.Asset, int64, error) {
	l1CompletedState := []string{"EdgesSelect", "EdgesPulling", "Servicing", "EdgesFailed"}

	query := fmt.Sprintf(`SELECT * FROM %s WHERE backup_result = 1 AND path = '' and state in (?)`, tableNameAsset)
	queryIn, queryArgs, err := sqlx.In(query, l1CompletedState)
	if err != nil {
		return nil, 0, err
	}

	var out []*model.Asset
	err = DB.SelectContext(ctx, &out, queryIn, queryArgs...)

	if err != nil {
		return nil, 0, err
	}

	count := fmt.Sprintf(`SELECT count(*) FROM %s WHERE backup_result = 1 AND path = '' and state in (?)`, tableNameAsset)
	countIn, countArgs, err := sqlx.In(count, l1CompletedState)
	if err != nil {
		return nil, 0, err
	}

	var total int64
	err = DB.GetContext(ctx, &total, countIn, countArgs...)
	if err != nil {
		return nil, 0, err
	}

	return out, total, err
}

func GetAssetByCID(ctx context.Context, cid string) (*model.Asset, error) {
	var asset model.Asset
	err := DB.GetContext(ctx, &asset, fmt.Sprintf(
		`SELECT * from %s WHERE cid = ?`, tableNameAsset), cid)
	if err != nil {
		return nil, err
	}
	return &asset, err
}

func GetAssetsByCID(ctx context.Context, cid string, aids []string) ([]*model.Asset, error) {
	var assets []*model.Asset

	// query := fmt.Sprintf(`SELECT * FROM %s WHERE backup_result = 1 AND path = '' and state in (?)`, tableNameAsset)
	// queryIn, queryArgs, err := sqlx.In(query, l1CompletedState)
	// if err != nil {
	// 	return nil, 0, err
	// }
	query := fmt.Sprintf(`SELECT * from %s WHERE cid = ? AND area_id IN (?)`, tableNameAsset)
	queryIn, queryArgs, err := sqlx.In(query, cid, aids)
	if err != nil {
		return nil, err
	}
	err = DB.SelectContext(ctx, &assets, queryIn, queryArgs...)
	if err != nil {
		return nil, err
	}
	return assets, err
}

func AllAssets(ctx context.Context) ([]*model.Asset, error) {
	var assets []*model.Asset
	err := DB.SelectContext(ctx, &assets, fmt.Sprintf("SELECT * FROM %s", tableNameAsset))
	if err != nil {
		return nil, err
	}
	return assets, nil
}

func GetAssetsList(ctx context.Context, cid string, nodeId, areaId string, option QueryOption) (int64, []*model.Asset, error) {
	var args []interface{}
	where := `WHERE 1 = 1`

	if cid != "" {
		where += fmt.Sprintf(" AND cid = '%s'", cid)
	}

	if nodeId != "" {
		where += fmt.Sprintf(" AND node_id = '%s'", nodeId)
	}

	if areaId != "" {
		where += fmt.Sprintf(" And area_id = '%s'", areaId)
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
	var out []*model.Asset

	err := DB.GetContext(ctx, &total, fmt.Sprintf(
		`SELECT count(*) FROM %s %s`, tableNameAsset, where,
	), args...)
	if err != nil {
		return 0, nil, err
	}
	err = DB.SelectContext(ctx, &out, fmt.Sprintf(
		`SELECT * FROM %s %s ORDER BY created_time DESC LIMIT %d OFFSET %d`, tableNameAsset, where, limit, offset,
	), args...)
	if err != nil {
		return 0, nil, err
	}

	return total, out, err
}

func GetAssetsListByCIds(ctx context.Context, cids []string) ([]*model.Asset, error) {
	var out []*model.Asset

	query := "SELECT * FROM assets where cid in (?)"

	query, args, err := sqlx.In(query, cids)
	if err != nil {
		return nil, err
	}

	err = DB.SelectContext(ctx, &out, query, args...)
	if err != nil {
		return nil, err
	}

	return out, err
}
