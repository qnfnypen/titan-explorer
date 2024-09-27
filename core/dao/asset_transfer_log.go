package dao

import (
	"context"
	"strings"
	"time"

	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/google/uuid"
)

func InsertAssetTransferLog(ctx context.Context, log *model.AssetTransferLog) error {
	statement := `INSERT INTO asset_transfer_log(trace_id, user_id, cid, hash, node_id, rate, cost_ms, total_size, state, transfer_type, log, created_at)
	 VALUES(:trace_id, :user_id, :cid, :hash, :node_id, :rate, :cost_ms, :total_size, :state, :transfer_type, log, :created_at)`
	_, err := DB.NamedExecContext(ctx, statement, log)
	return err
}

func NewLogTrace(ctx context.Context, uid string) (string, error) {
	log := &model.AssetTransferLog{
		TraceId:   uuid.New().String(),
		UserId:    uid,
		CreatedAt: time.Now(),
	}
	if err := InsertAssetTransferLog(ctx, log); err != nil {
		return "", err
	}
	return log.TraceId, nil
}

type ComprehensiveStats struct {
	TotalDownloads   int `db:"total_downloads"`
	TotalUploads     int `db:"total_uploads"`
	DownloadSuccess  int `db:"download_success"`
	UploadSuccess    int `db:"upload_success"`
	DownloadFailure  int `db:"download_failure"`
	UploadFailure    int `db:"upload_failure"`
	DownloadSize     int `db:"download_size"`
	UploadSize       int `db:"upload_size"`
	DownloadAvgSpeed int `db:"download_avg_speed"`
	UploadAvgSpeed   int `db:"upload_avg_speed"`
}

// 获取所有统计数据
func GetComprehensiveStatsInPeriod(ctx context.Context, start, end int64) (*ComprehensiveStats, error) {
	var stats ComprehensiveStats
	query := `
		SELECT 
			SUM(CASE WHEN transfer_type = 'download' THEN 1 ELSE 0 END) AS total_downloads,
			SUM(CASE WHEN transfer_type = 'upload' THEN 1 ELSE 0 END) AS total_uploads,
			SUM(CASE WHEN transfer_type = 'download' AND state = 1 THEN 1 ELSE 0 END) AS download_success,
			SUM(CASE WHEN transfer_type = 'upload' AND state = 1 THEN 1 ELSE 0 END) AS upload_success,
			SUM(CASE WHEN transfer_type = 'download' AND state = 2 THEN 1 ELSE 0 END) AS download_failure,
			SUM(CASE WHEN transfer_type = 'upload' AND state = 2 THEN 1 ELSE 0 END) AS upload_failure,
			SUM(CASE WHEN transfer_type = 'download' THEN total_size ELSE 0 END) AS download_size,
			SUM(CASE WHEN transfer_type = 'upload' THEN total_size ELSE 0 END) AS upload_size,
			AVG(CASE WHEN transfer_type = 'download' AND state = 1 THEN rate ELSE 0 END) AS download_avg_speed,
			AVG(CASE WHEN transfer_type = 'upload' AND state = 1 THEN rate ELSE 0 END) AS upload_avg_speed
		FROM asset_transfer_log`

	// 检查是否需要加上时间条件
	var conditions []string
	args := []interface{}{}

	// 如果 start 和 end 都不为 0，则添加时间范围条件
	if start > 0 {
		conditions = append(conditions, "created_at >= FROM_UNIXTIME(?)")
		args = append(args, start)
	}
	if end > 0 {
		conditions = append(conditions, "created_at <= FROM_UNIXTIME(?)")
		args = append(args, end)
	}

	// 如果有条件，则将它们添加到查询中
	if len(conditions) > 0 {
		query += " WHERE " + joinConditions(conditions)
	}

	err := DB.GetContext(ctx, &stats, query, args...)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// joinConditions 将查询条件用 AND 连接
func joinConditions(conditions []string) string {
	return " " + strings.Join(conditions, " AND ")
}
