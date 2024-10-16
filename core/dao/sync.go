package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

const (
	tableIPFSRecords = "sync_ipfs_record"
)

// AddIPFSRecords 添加ipfs同步记录
func AddIPFSRecords(ctx context.Context, irs []model.SyncIPFSRecord) error {
	sq := squirrel.Insert(tableIPFSRecords).Columns("username,cid,timestamp")

	for _, v := range irs {
		sq = sq.Values(v.Username, v.CID, v.Timestamp)
	}

	query, args, err := sq.Options("IGNORE").ToSql()
	if err != nil {
		return fmt.Errorf("generate sql of add sync_ipfs_records error:%w", err)
	}

	_, err = DB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("add sync_ipfs_records error:%w", err)
	}

	return nil
}

// GetUnSyncIPFSRecords 获取未同步成功的ipfs文件列表且时间不超过一个小时
func GetUnSyncIPFSRecords(ctx context.Context) ([]model.SyncIPFSRecord, error) {
	var irs []model.SyncIPFSRecord

	query, args, err := squirrel.Select("*").From(tableIPFSRecords).Where("status = 0 AND timestamp >= ?", time.Now().Unix()-3600).ToSql()
	if err != nil {
		return nil, fmt.Errorf("generate sql of get unsync ipfs records error:%w", err)
	}

	err = DB.SelectContext(ctx, &irs, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get unsync ipfs records error:%w", err)
	}

	return irs, nil
}

// UpdateIPFSRecordStatus 更新ipfs文件同步状态
func UpdateIPFSRecordStatus(ctx context.Context, cids []string, areaID string) error {
	query, args, err := squirrel.Update(tableIPFSRecords).Set("status", 1).Where(squirrel.Eq{"cid": cids, "area_id": areaID}).ToSql()
	if err != nil {
		return fmt.Errorf("generate sql of update status of ipfs's record error:%w", err)
	}

	_, err = DB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update status of ipfs's record error:%w", err)
	}

	return nil
}

// GetIPFSRecordsByCIDs 根据cid获取ipfs记录列表
func GetIPFSRecordsByCIDs(ctx context.Context, cids []string) ([]model.SyncIPFSRecord, error) {
	var rs []model.SyncIPFSRecord

	query, args, err := squirrel.Select("*").From(tableIPFSRecords).Where(squirrel.Eq{"cid": cids}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("generate sql of get ipfs's record error:%w", err)
	}

	err = DB.SelectContext(ctx, &rs, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get ipfs's record error:%w", err)
	}

	return rs, nil
}

// GetIPFSRecordsByUsername 根据用户名获取ipfs同步记录
func GetIPFSRecordsByUsername(ctx context.Context, un string, page, size int) (int64, []model.SyncIPFSRecord, error) {
	var (
		total int64
		rs    []model.SyncIPFSRecord
	)

	query, args, err := squirrel.Select("COUNT(id)").From(tableIPFSRecords).Where("username = ?", un).ToSql()
	if err != nil {
		return 0, nil, fmt.Errorf("generate sql of get total of ipfs's record error:%w", err)
	}
	err = DB.GetContext(ctx, &total, query, args...)
	if err != nil {
		return 0, nil, fmt.Errorf("get total of ipfs's record error:%w", err)
	}

	query, args, err = squirrel.Select("*").From(tableIPFSRecords).Where("username = ?", un).OrderBy("id DESC").Limit(uint64(size)).Offset(uint64(page-1) * uint64(size)).ToSql()
	if err != nil {
		return 0, nil, fmt.Errorf("generate sql of get ipfs's record error:%w", err)
	}

	err = DB.SelectContext(ctx, &rs, query, args...)
	if err != nil {
		return 0, nil, fmt.Errorf("get ipfs's record error:%w", err)
	}

	return total, rs, nil
}
