package dao

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Filecoin-Titan/titan/api/types"
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

var (
	AssetTransferTypeMap = map[string]types.ServiceType{
		AssetTransferTypeDownload: types.ServiceTypeDownload,
		AssetTransferTypeUpload:   types.ServiceTypeUpload,
	}

	AssetServiceStatusMap = map[int]types.ServiceStatus{
		AssetTransferStateSuccess: types.ServiceTypeSucceed,
		AssetTransferStateFailure: types.ServiceTypeFailed,
	}
)

func InsertOrUpdateAssetTransferLog(ctx context.Context, record *model.AssetTransferLog, details []*model.AssetTrasnferDetail) error {
	tx, err := DB.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		err = tx.Rollback()
		if err != nil && err != sql.ErrTxDone {
			log.Errorf("InsertOrUpdateAssetTransferLog Rollback err:%s", err.Error())
		}
	}()

	recordStatement := `INSERT INTO asset_transfer_log(trace_id, user_id, cid, hash, node_id, rate, cost_ms, total_size, state, transfer_type, log, area, created_at, ip, first_byte_time, available_bandwidth)
	VALUES(:trace_id, :user_id, :cid, :hash, :node_id, :rate, :cost_ms, :total_size, :state, :transfer_type, :log, :area, :created_at, :ip, :first_byte_time, :available_bandwidth)
	ON DUPLICATE KEY UPDATE 
	user_id=VALUES(user_id), cid=VALUES(cid), hash=VALUES(hash), node_id=VALUES(node_id), rate=VALUES(rate), cost_ms=VALUES(cost_ms), 
	total_size=VALUES(total_size), state=VALUES(state), transfer_type=VALUES(transfer_type), log=VALUES(log), area=VALUES(area), ip=VALUES(ip), 
	first_byte_time=VALUES(first_byte_time), available_bandwidth=VALUES(available_bandwidth)`
	_, err = tx.NamedExecContext(ctx, recordStatement, record)
	if err != nil {
		return err
	}

	if len(details) == 0 {
		return tx.Commit()
	}

	detailStatement := `INSERT INTO asset_transfer_detail(trace_id, node_id, state, transfer_type, peek, elasped_time, size, errors, created_at) 
	VALUES(:trace_id, :node_id, :state, :transfer_type, :peek, :elasped_time, :size, :errors, :created_at)`

	_, err = tx.NamedExecContext(ctx, detailStatement, details)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func NewLogTrace(ctx context.Context, uid string, transferType string, area string) (string, error) {
	log := &model.AssetTransferLog{
		TraceId:      uuid.New().String(),
		UserId:       uid,
		CreatedAt:    time.Now(),
		TransferType: transferType,
		Area:         area,
	}
	if err := InsertOrUpdateAssetTransferLog(ctx, log, nil); err != nil {
		return "", err
	}
	return log.TraceId, nil
}

type ComprehensiveStats struct {
	TotalDownloads   int     `db:"total_downloads" json:"total_downloads"`
	TotalUploads     int     `db:"total_uploads" json:"total_uploads"`
	DownloadSuccess  int     `db:"download_success" json:"download_success"`
	UploadSuccess    int     `db:"upload_success" json:"upload_success"`
	DownloadFailure  int     `db:"download_failure" json:"download_failure"`
	UploadFailure    int     `db:"upload_failure" json:"upload_failure"`
	DownloadSize     int     `db:"download_size" json:"download_size"`
	UploadSize       int     `db:"upload_size" json:"upload_size"`
	DownloadAvgSpeed float64 `db:"download_avg_speed" json:"download_avg_speed"`
	UploadAvgSpeed   float64 `db:"upload_avg_speed" json:"upload_avg_speed"`
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

	err := DB.GetContext(ctx, &stats, query, args...)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

type ComprehensiveStatsWithDay struct {
	*ComprehensiveStats
	Day string `db:"day" json:"day"`
}

func GetComprehensiveStatsInPeriodByUserGroupByDay(ctx context.Context, start, end int64, username string) ([]*ComprehensiveStatsWithDay, error) {
	var stats []*ComprehensiveStatsWithDay
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

type AssetTransferDetailError struct {
	StatusCode int64
	Message    string
}

type AssetTransferDetailErrors []*AssetTransferDetailError

func (a AssetTransferDetailErrors) Append(code int64, message string) AssetTransferDetailErrors {
	if code >= http.StatusOK && code <= http.StatusIMUsed {
		return a
	}
	if message == "" {
		return a
	}

	return append(a, &AssetTransferDetailError{StatusCode: code, Message: message})
}

func (a AssetTransferDetailErrors) ToString() string {
	if len(a) == 0 {
		return "[]"
	}
	jsonStr, err := json.Marshal(a)
	if err != nil {
		log.Errorf("AssetTransferDetailErrors ToString err:%s", err.Error())
		return "[]"
	}
	return string(jsonStr)
}
