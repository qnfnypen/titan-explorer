package dao

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/utils"
	"time"
)

const tableNameRetrievalEvent = "retrieval_event"
const tableNameValidateEvent = "validation_event"

func GetLastRetrievalEvent(ctx context.Context) (*model.RetrievalEvent, error) {
	var out model.RetrievalEvent
	query := fmt.Sprintf(`SELECT * FROM %s ORDER BY updated_at DESC LIMIT 1 OFFSET 1;`, tableNameRetrievalEvent)
	err := DB.QueryRowxContext(ctx, query).StructScan(&out)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &out, nil
}

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

func GenerateRetrievalEvents(ctx context.Context, startTime, endTime time.Time) ([]*model.RetrievalEvent, error) {
	queryStatement := fmt.Sprintf(`
	select device_id, max(retrieval_count) as blocks,
			FROM_UNIXTIME(CEIL(UNIX_TIMESTAMP(time)/300)*300) as time, 
			max(upstream_traffic) as upstream_bandwidth
	from %s where time >= ? and time < ? GROUP BY device_id ,
			FROM_UNIXTIME(CEIL(UNIX_TIMESTAMP(time)/300)*300) ORDER BY time`, tableNameDeviceInfoHour)

	var events []*model.RetrievalEvent
	err := DB.SelectContext(ctx, &events, queryStatement, startTime, endTime)
	if err != nil {
		return nil, err
	}

	var out []*model.RetrievalEvent
	eventInDate := make(map[string]*model.RetrievalEvent)
	for _, event := range events {
		last, ok := eventInDate[event.DeviceID]
		if !ok {
			eventInDate[event.DeviceID] = event
			continue
		}

		if event.Blocks-last.Blocks <= 0 {
			eventInDate[event.DeviceID] = event
			continue
		}

		out = append(out, &model.RetrievalEvent{
			DeviceID:          event.DeviceID,
			Time:              event.Time,
			Blocks:            event.Blocks - last.Blocks,
			UpstreamBandwidth: utils.ToFixed(event.UpstreamBandwidth-last.UpstreamBandwidth, 2),
		})
		eventInDate[event.DeviceID] = event
	}

	return out, err
}

func CreateRetrievalEvents(ctx context.Context, events []*model.RetrievalEvent) error {
	statement := fmt.Sprintf(`INSERT INTO %s(device_id, blocks, carfile_cid, token_id, client_id, block_size, status, time, upstream_bandwidth, start_time, end_time) 
	VALUES (:device_id, :blocks, :carfile_cid, :token_id, :client_id, :block_size, :status, :time, :upstream_bandwidth, :start_time, :end_time) 
	ON DUPLICATE KEY UPDATE blocks = VALUES(blocks), status = VALUES(status), end_time = VALUES(end_time), upstream_bandwidth = VALUES(upstream_bandwidth)`, tableNameRetrievalEvent)
	_, err := DB.NamedExecContext(ctx, statement, events)
	return err
}

func GetRetrievalEventsByPage(ctx context.Context, cond *model.RetrievalEvent, option QueryOption) ([]*model.RetrievalEvent, int64, error) {
	var args []interface{}
	where := `WHERE 1=1 AND status = 1`
	if cond.DeviceID != "" {
		where += ` AND device_id = ?`
		args = append(args, cond.DeviceID)
	}

	if option.Order != "" && option.OrderField != "" {
		where += fmt.Sprintf(` ORDER BY %s %s`, option.OrderField, option.Order)
	} else {
		where += " ORDER BY time DESC"
	}

	limit := option.PageSize
	offset := option.Page
	if option.PageSize <= 0 {
		limit = 50
	}
	if option.Page > 0 {
		offset = limit * (option.Page - 1)
	}

	var total int64
	var out []*model.RetrievalEvent

	err := DB.GetContext(ctx, &total, fmt.Sprintf(
		`SELECT count(*) FROM %s %s`, tableNameRetrievalEvent, where,
	), args...)
	if err != nil {
		return nil, 0, err
	}

	err = DB.SelectContext(ctx, &out, fmt.Sprintf(
		`SELECT * FROM %s %s LIMIT %d OFFSET %d`, tableNameRetrievalEvent, where, limit, offset,
	), args...)
	if err != nil {
		return nil, 0, err
	}

	return out, total, err
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
	out.TotalDownload += out2.TotalDownload
	out.DeviceID = nodeId
	err = UpdateTotalDownload(ctx, &out)
	if err != nil {
		return err
	}
	return nil
}
