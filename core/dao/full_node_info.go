package dao

import (
	"context"
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
                total_download_bandwidth, car_file_count, total_file_size, file_download_count, time, created_at) 
		VALUES (:validator_count, :candidate_count, :edge_count, :total_storage, :total_uplink_bandwidth, 
                :total_download_bandwidth, :car_file_count, :total_file_size, :file_download_count, :time, :created_at)`, tableNameFullNodeInfoHours),
		fullNodeInfoHour)
	return err
}
