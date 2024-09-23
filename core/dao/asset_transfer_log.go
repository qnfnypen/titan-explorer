package dao

import (
	"context"

	"github.com/gnasnik/titan-explorer/core/generated/model"
)

func InsertAssetTransferLog(ctx context.Context, log *model.AssetTransferLog) error {
	statement := `INSERT INTO asset_transfer_log(trace_id, user_id, cid, hash, rate, cost_ms, total_size, succeed, transfer_type, created_at)
	 VALUES(:trace_id, :user_id, :cid, :hash, :rate, :cost_ms, :total_size, :succeed, :transfer_type, :created_at)`
	_, err := DB.NamedExecContext(ctx, statement, log)
	return err
}
