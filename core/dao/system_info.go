package dao

import (
	"context"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

var tableNameSystemInfo = "system_info"

func UpsertSystemInfo(ctx context.Context, systemInfo *model.SystemInfo) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`INSERT INTO %s (scheduler_uuid, car_file_count, download_count, next_election_time) 
		VALUES (:scheduler_uuid, :car_file_count, :download_count, :next_election_time) ON DUPLICATE KEY UPDATE 
		scheduler_uuid =:scheduler_uuid, car_file_count=:car_file_count, download_count=:download_count, 
		    next_election_time =:next_election_time`, tableNameSystemInfo),
		systemInfo)
	return err
}

func SumSystemInfo(ctx context.Context) (*model.SystemInfo, error) {
	queryStatement := fmt.Sprintf(`SELECT sum(car_file_count) as car_file_count, sum(download_count) as download_count, 
       min(next_election_time) as next_election_time FROM %s;`, tableNameSystemInfo)

	var out model.SystemInfo
	if err := DB.QueryRowxContext(ctx, queryStatement).StructScan(&out); err != nil {
		return nil, err
	}
	return &out, nil
}
