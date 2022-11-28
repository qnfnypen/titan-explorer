package dao

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

var (
	tableNameFullNodeInfoHours = "full_node_info_hours"
	tableNameFullNodeInfoDays  = "full_node_info_days"
)

func AddFullNodeInfoHours(ctx context.Context, fullNodeInfoHour *model.FullNodeInfoHour) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`INSERT INTO %s (validator_count, candidate_count, edge_count, total_storage, total_uplink_bandwidth, 
                total_download_bandwidth, total_carfile, total_carfile_size, download_count, time, created_at) 
		VALUES (:validator_count, :candidate_count, :edge_count, :total_storage, :total_uplink_bandwidth, :total_download_bandwidth
		  , :total_carfile, :total_carfile_size, :download_count, :time, :created_at)`, tableNameFullNodeInfoHours),
		fullNodeInfoHour)
	return err
}

func AddFullNodeInfoDays(ctx context.Context, fullNodeInfoDay *model.FullNodeInfoHour) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`INSERT INTO %s (validator_count, candidate_count, edge_count, total_storage, total_uplink_bandwidth, 
                total_download_bandwidth, total_carfile, total_carfile_size, download_count, time, created_at) 
		VALUES (:validator_count, :candidate_count, :edge_count, :total_storage, :total_uplink_bandwidth, :total_download_bandwidth
		  , :total_carfile, :total_carfile_size, :download_count, :time, :created_at)`, tableNameFullNodeInfoDays),
		fullNodeInfoDay)
	return err
}

func GetFullNodeInfo(ctx context.Context) (*model.FullNodeInfoHour, error) {
	var out model.FullNodeInfoHour
	if err := DB.QueryRowxContext(ctx, fmt.Sprintf(
		`SELECT * FROM %s ORDER BY created_at DESC LIMIT 1`, tableNameFullNodeInfoHours)).StructScan(&out); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}
