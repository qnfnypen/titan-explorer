package dao

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

const tableNameRetrievalEvent = "retrieval_event"
const tableNameValidateEvent = "validation_event"

func GetUnfinishedEvent(ctx context.Context) ([]string, error) {
	var tokenIds []string
	query := fmt.Sprintf(`SELECT token_id FROM %s WHERE status=0;`, tableNameRetrievalEvent)
	err := DB.SelectContext(ctx, &tokenIds, query)
	if err == sql.ErrNoRows {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return tokenIds, nil
}

func CountRetrievalEvent(ctx context.Context, nodeId string) error {
	var out model.DeviceInfo
	query := fmt.Sprintf(`SELECT device_id,count(DISTINCT(token_id)) as download_count,sum(block_size) as total_upload  FROM %s where status = 1 and device_id = '%s' GROUP BY device_id;`, tableNameRetrievalEvent, nodeId)
	err := DB.QueryRowxContext(ctx, query).StructScan(&out)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	err = UpdateDownloadCount(ctx, &out)
	if err != nil {
		return err
	}
	return nil
}

func CreateRetrievalEvents(ctx context.Context, events []*model.RetrievalEvent) error {
	statement := fmt.Sprintf(`INSERT INTO %s(device_id, blocks, carfile_cid, token_id, client_id, block_size, status, time, upstream_bandwidth, start_time, end_time) 
	VALUES (:device_id, :blocks, :carfile_cid, :token_id, :client_id, :block_size, :status, :time, :upstream_bandwidth, :start_time, :end_time) 
	ON DUPLICATE KEY UPDATE blocks = VALUES(blocks), status = VALUES(status), end_time = VALUES(end_time), upstream_bandwidth = VALUES(upstream_bandwidth)`, tableNameRetrievalEvent)
	_, err := DB.NamedExecContext(ctx, statement, events)
	return err
}

func CountUploadTraffic(ctx context.Context, nodeId string) error {
	if !GetIdIfExit(ctx, nodeId) {
		return nil
	}
	// count retrieval Traffic
	var out model.DeviceInfo
	query := fmt.Sprintf(`SELECT client_id as device_id,sum(block_size) as total_download FROM %s where status = 1 and client_id = '%s' GROUP BY client_id;`, tableNameRetrievalEvent, nodeId)
	_ = DB.QueryRowxContext(ctx, query).StructScan(&out)
	// count validate Traffic
	var out2 model.DeviceInfo
	query = fmt.Sprintf(`SELECT validator_id as device_id,sum(upstream_traffic) as total_download FROM %s where status = 1 and validator_id = '%s' GROUP BY validator_id;`, tableNameValidateEvent, nodeId)
	err := DB.QueryRowxContext(ctx, query).StructScan(&out2)
	if err != nil {
		log.Infof("count validate Traffic:%v", err)
	}
	out.DownloadTraffic += out2.DownloadTraffic
	out.DeviceID = nodeId
	err = UpdateTotalDownload(ctx, &out)
	if err != nil {
		return err
	}
	return nil
}
