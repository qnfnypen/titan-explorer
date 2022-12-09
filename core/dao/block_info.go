package dao

import (
	"context"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/jmoiron/sqlx"
	"time"
)

const tableNameBlockInfo = "block_info"

func CreateBlockInfo(ctx context.Context, blockInfos []*model.BlockInfo) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`INSERT INTO %s (device_id, carfile_cid, carfile_hash, status, size, created_time)
	VALUES (:device_id, :carfile_cid, :carfile_hash, :status, :size, :created_time)`, tableNameBlockInfo,
	), blockInfos)
	return err
}

func groupBlocksAndInsert(ctx context.Context, tx *sqlx.Tx, startTime, endTime time.Time) error {
	queryStatement := fmt.Sprintf(`
INSERT INTO %s(device_id, carfile_cid, block_size, blocks, time)
SELECT device_id, carfile_cid, sum(size) AS block_size, count(*) AS blocks,  
		DATE_FORMAT(
			CONCAT( DATE( created_time ), ' ', HOUR ( created_time ), ':', floor( MINUTE ( created_time ) / 5 ) * 5 ),'%%Y-%%m-%%d %%H:%%i' )
			 AS time FROM %s WHERE created_time >= ? AND created_time < ? 
 GROUP BY device_id, carfile_cid, DATE_FORMAT( time, '%%Y-%%m-%%d %%H:%%i' )`, tableNameCacheEvent, tableNameBlockInfo)

	_, err := tx.ExecContext(ctx, queryStatement, startTime, endTime)
	return err
}

func deleteBlocksByTimeTx(ctx context.Context, tx *sqlx.Tx, startTime, endTime time.Time) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE created_time >= ? AND created_time <= ?`, tableNameBlockInfo)
	_, err := tx.ExecContext(ctx, query, startTime, endTime)
	return err
}

func TxStatisticDeviceBlocks(ctx context.Context, startTime, endTime time.Time) error {
	tx, err := DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = groupBlocksAndInsert(ctx, tx, startTime, endTime)
	if err != nil {
		return err
	}

	err = deleteBlocksByTimeTx(ctx, tx, startTime, endTime)
	if err != nil {
		return err
	}

	return tx.Commit()
}
