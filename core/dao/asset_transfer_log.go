package dao

import (
	"context"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/google/uuid"
)

const (
	AssetTransferTypeDownload = "download"
	AssetTransferTypeUpload   = "upload"

	AssetTransferStateCreated = 0
	AssetTransferStateSuccess = 1
	AssetTransferStateFailure = 2
)

func InsertOrUpdateAssetTransferLog(ctx context.Context, log *model.AssetTransferLog) error {
	statement := `INSERT INTO asset_transfer_log(trace_id, user_id, cid, hash, node_id, rate, cost_ms, total_size, state, transfer_type, log, area, created_at, ip)
	VALUES(:trace_id, :user_id, :cid, :hash, :node_id, :rate, :cost_ms, :total_size, :state, :transfer_type, :log, :area, :created_at, :ip)
	ON DUPLICATE KEY UPDATE 
	user_id=VALUES(user_id), cid=VALUES(cid), hash=VALUES(hash), node_id=VALUES(node_id), rate=VALUES(rate), cost_ms=VALUES(cost_ms), 
	total_size=VALUES(total_size), state=VALUES(state), transfer_type=VALUES(transfer_type), log=VALUES(log), area=VALUES(area), ip=VALUES(ip)`
	_, err := DB.NamedExecContext(ctx, statement, log)
	return err
}

func NewLogTrace(ctx context.Context, uid string, transferType string, area string) (string, error) {
	log := &model.AssetTransferLog{
		TraceId:      uuid.New().String(),
		UserId:       uid,
		CreatedAt:    time.Now(),
		TransferType: transferType,
		Area:         area,
	}
	if err := InsertOrUpdateAssetTransferLog(ctx, log); err != nil {
		return "", err
	}
	return log.TraceId, nil
}

type ComprehensiveStats struct {
	TotalDownloads   int     `db:"total_downloads"`
	TotalUploads     int     `db:"total_uploads"`
	DownloadSuccess  int     `db:"download_success"`
	UploadSuccess    int     `db:"upload_success"`
	DownloadFailure  int     `db:"download_failure"`
	UploadFailure    int     `db:"upload_failure"`
	DownloadSize     int     `db:"download_size"`
	UploadSize       int     `db:"upload_size"`
	DownloadAvgSpeed float64 `db:"download_avg_speed"`
	UploadAvgSpeed   float64 `db:"upload_avg_speed"`
}

// 获取所有统计数据
func GetComprehensiveStatsInPeriod(ctx context.Context, start, end int64, area string) (*ComprehensiveStats, error) {
	var stats ComprehensiveStats
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN transfer_type = 'download' THEN 1 ELSE 0 END), 0) AS total_downloads,
			COALESCE(SUM(CASE WHEN transfer_type = 'upload' THEN 1 ELSE 0 END), 0) AS total_uploads,
			COALESCE(SUM(CASE WHEN transfer_type = 'download' AND state = 1 THEN 1 ELSE 0 END), 0) AS download_success,
			COALESCE(SUM(CASE WHEN transfer_type = 'upload' AND state = 1 THEN 1 ELSE 0 END), 0) AS upload_success,
			COALESCE(SUM(CASE WHEN transfer_type = 'download' AND state = 2 THEN 1 ELSE 0 END), 0) AS download_failure,
			COALESCE(SUM(CASE WHEN transfer_type = 'upload' AND state = 2 THEN 1 ELSE 0 END), 0) AS upload_failure,
			COALESCE(SUM(CASE WHEN transfer_type = 'download' THEN total_size ELSE 0 END), 0) AS download_size,
			COALESCE(SUM(CASE WHEN transfer_type = 'upload' THEN total_size ELSE 0 END), 0) AS upload_size,
			COALESCE(AVG(CASE WHEN transfer_type = 'download' AND state = 1 THEN rate ELSE NULL END), 0) AS download_avg_speed,
			COALESCE(AVG(CASE WHEN transfer_type = 'upload' AND state = 1 THEN rate ELSE NULL END), 0) AS upload_avg_speed
		FROM asset_transfer_log`

	var conditions []string
	args := []interface{}{}

	// 添加时间范围条件
	if start > 0 {
		conditions = append(conditions, "created_at >= FROM_UNIXTIME(?)")
		args = append(args, start)
	}
	if end > 0 {
		conditions = append(conditions, "created_at <= FROM_UNIXTIME(?)")
		args = append(args, end)
	}

	if area != "" {
		conditions = append(conditions, "area = ?")
		args = append(args, area)
	}

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

// 获取用户的统计
func GetComprehensiveStatsInPeriodByUser(ctx context.Context, start, end int64, username string) (*ComprehensiveStats, error) {
	var stats ComprehensiveStats
	query := `
		SELECT 
			DATE(created_at) AS day,
			COALESCE(SUM(CASE WHEN transfer_type = 'download' THEN 1 ELSE 0 END), 0) AS total_downloads,
			COALESCE(SUM(CASE WHEN transfer_type = 'upload' THEN 1 ELSE 0 END), 0) AS total_uploads,
			COALESCE(SUM(CASE WHEN transfer_type = 'download' AND state = 1 THEN 1 ELSE 0 END), 0) AS download_success,
			COALESCE(SUM(CASE WHEN transfer_type = 'upload' AND state = 1 THEN 1 ELSE 0 END), 0) AS upload_success,
			COALESCE(SUM(CASE WHEN transfer_type = 'download' AND state = 2 THEN 1 ELSE 0 END), 0) AS download_failure,
			COALESCE(SUM(CASE WHEN transfer_type = 'upload' AND state = 2 THEN 1 ELSE 0 END), 0) AS upload_failure,
			COALESCE(SUM(CASE WHEN transfer_type = 'download' THEN total_size ELSE 0 END), 0) AS download_size,
			COALESCE(SUM(CASE WHEN transfer_type = 'upload' THEN total_size ELSE 0 END), 0) AS upload_size,
			COALESCE(AVG(CASE WHEN transfer_type = 'download' AND state = 1 THEN rate ELSE NULL END), 0) AS download_avg_speed,
			COALESCE(AVG(CASE WHEN transfer_type = 'upload' AND state = 1 THEN rate ELSE NULL END), 0) AS upload_avg_speed
		FROM asset_transfer_log`

	var conditions []string
	args := []interface{}{}

	// 添加时间范围条件
	if start > 0 {
		conditions = append(conditions, "created_at >= FROM_UNIXTIME(?)")
		args = append(args, start)
	}
	if end > 0 {
		conditions = append(conditions, "created_at <= FROM_UNIXTIME(?)")
		args = append(args, end)
	}

	if username != "" {
		conditions = append(conditions, "user_id = ?")
		args = append(args, username)
	}

	if len(conditions) > 0 {
		query += " WHERE " + joinConditions(conditions)
	}

	query += " GROUP BY DATE(created_at) ORDER BY DATE(created_at) DESC"

	err := DB.GetContext(ctx, &stats, query, args...)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

func GetComprehensiveStatsInPeriodByUserGroupByDay(ctx context.Context, start, end int64, username string) ([]*ComprehensiveStats, error) {
	var stats []*ComprehensiveStats
	query := `
		SELECT 
			DATE(created_at) AS day,
			COALESCE(SUM(CASE WHEN transfer_type = 'download' THEN 1 ELSE 0 END), 0) AS total_downloads,
			COALESCE(SUM(CASE WHEN transfer_type = 'upload' THEN 1 ELSE 0 END), 0) AS total_uploads,
			COALESCE(SUM(CASE WHEN transfer_type = 'download' AND state = 1 THEN 1 ELSE 0 END), 0) AS download_success,
			COALESCE(SUM(CASE WHEN transfer_type = 'upload' AND state = 1 THEN 1 ELSE 0 END), 0) AS upload_success,
			COALESCE(SUM(CASE WHEN transfer_type = 'download' AND state = 2 THEN 1 ELSE 0 END), 0) AS download_failure,
			COALESCE(SUM(CASE WHEN transfer_type = 'upload' AND state = 2 THEN 1 ELSE 0 END), 0) AS upload_failure,
			COALESCE(SUM(CASE WHEN transfer_type = 'download' THEN total_size ELSE 0 END), 0) AS download_size,
			COALESCE(SUM(CASE WHEN transfer_type = 'upload' THEN total_size ELSE 0 END), 0) AS upload_size,
			COALESCE(AVG(CASE WHEN transfer_type = 'download' AND state = 1 THEN rate ELSE NULL END), 0) AS download_avg_speed,
			COALESCE(AVG(CASE WHEN transfer_type = 'upload' AND state = 1 THEN rate ELSE NULL END), 0) AS upload_avg_speed
		FROM asset_transfer_log`

	var conditions []string
	args := []interface{}{}

	// 添加时间范围条件
	if start > 0 {
		conditions = append(conditions, "created_at >= FROM_UNIXTIME(?)")
		args = append(args, start)
	}
	if end > 0 {
		conditions = append(conditions, "created_at <= FROM_UNIXTIME(?)")
		args = append(args, end)
	}

	if username != "" {
		conditions = append(conditions, "user_id = ?")
		args = append(args, username)
	}

	if len(conditions) > 0 {
		query += " WHERE " + joinConditions(conditions)
	}

	query += " GROUP BY DATE(created_at) ORDER BY DATE(created_at) DESC"

	err := DB.SelectContext(ctx, &stats, query, args...)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func ListAssetTransferDetail(ctx context.Context, sb squirrel.SelectBuilder) ([]*model.AssetTransferLog, error) {
	var res []*model.AssetTransferLog
	query, args, err := sb.ToSql()
	if err != nil {
		return nil, err
	}

	if err := DB.SelectContext(ctx, &res, query, args...); err != nil {
		return nil, err
	}
	return res, nil
}
