package dao

import (
	"context"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

var tableNameFilStorage = "fil_storage"

func AddFilStorages(ctx context.Context, storages []*model.FilStorage) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`INSERT INTO %s ( provider, sector_num, cost, message_cid, piece_cid, payload_cid, deal_id, path, f_index, piece_size, start_height, end_height, start_time, end_time, created_at, updated_at)
			VALUES ( :provider, :sector_num, :cost, :message_cid, :piece_cid, :payload_cid, :deal_id, :path, :f_index, :piece_size, :start_height, :end_height, :start_time, :end_time, :created_at, :updated_at)
			ON DUPLICATE KEY UPDATE  provider = VALUES(provider), sector_num = VALUES(sector_num), cost = VALUES(cost), message_cid = VALUES(message_cid), 
			piece_cid = VALUES(piece_cid), payload_cid = VALUES(payload_cid), deal_id = VALUES(deal_id), piece_size = VALUES(piece_size), 
			start_height = VALUES(start_height), end_height = VALUES(end_height), start_time = VALUES(start_time), end_time = VALUES(end_time);`, tableNameFilStorage,
	), storages)
	return err
}

func CountFilStorage(ctx context.Context, cid string) (int64, error) {
	var total int64
	err := DB.GetContext(ctx, &total, fmt.Sprintf(
		`select count(*) from %s f left join %s a on f.path = a.path where a.path <> '' and a.cid = ?`, tableNameFilStorage, tableNameAsset,
	), cid)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func ListFilStorages(ctx context.Context, path string, option QueryOption) ([]*model.FilStorage, int64, error) {
	var args []interface{}
	var total int64
	var out []*model.FilStorage

	limit := option.PageSize
	offset := option.Page
	if option.PageSize <= 0 {
		limit = 50
	}
	if option.Page > 0 {
		offset = limit * (option.Page - 1)
	}

	args = append(args, path)

	err := DB.GetContext(ctx, &total, fmt.Sprintf(
		`SELECT count(*) FROM %s WHERE path = ?`, tableNameFilStorage,
	), args...)
	if err != nil {
		return nil, 0, err
	}

	err = DB.SelectContext(ctx, &out, fmt.Sprintf(
		`SELECT * FROM %s WHERE path = ? LIMIT %d OFFSET %d`, tableNameFilStorage, limit, offset,
	), args...)
	if err != nil {
		return nil, 0, err
	}

	return out, total, err
}

func SumFilStorage(ctx context.Context) (int64, error) {
	var total int64
	err := DB.GetContext(ctx, &total, fmt.Sprintf(`select sum(piece_size) from %s`, tableNameFilStorage))
	if err != nil {
		return 0, err
	}
	return total, nil
}
