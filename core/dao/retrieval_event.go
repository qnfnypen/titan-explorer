package dao

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"time"
)

const tableNameRetrievalEvent = "retrieval_event"

func GetLastRetrievalEvent(ctx context.Context) (*model.RetrievalEvent, error) {
	var out model.RetrievalEvent
	query := fmt.Sprintf(`SELECT * FROM %s ORDER BY time DESC LIMIT 1;`, tableNameRetrievalEvent)
	err := DB.QueryRowxContext(ctx, query).StructScan(&out)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func GenerateRetrievalEvents(ctx context.Context, startTime, endTime time.Time) ([]*model.RetrievalEvent, error) {
	queryStatement := fmt.Sprintf(`select * from (
	select device_id, max(retrieval_count) - min(retrieval_count) as blocks,
			FROM_UNIXTIME(FLOOR(UNIX_TIMESTAMP(time)/300)*300) as time, 
			max(upstream_traffic) - min(upstream_traffic) as upstream_bandwidth
	from %s where time >= ? and time < ? GROUP BY device_id ,
			FROM_UNIXTIME(FLOOR(UNIX_TIMESTAMP(time)/300)*300)
) a where blocks > 0 `, tableNameDeviceInfoHour)

	var out []*model.RetrievalEvent
	err := DB.SelectContext(ctx, &out, queryStatement, startTime, endTime)
	return out, err
}

func CreateRetrievalEvents(ctx context.Context, events []*model.RetrievalEvent) error {
	statement := fmt.Sprintf(`INSERT INTO %s(device_id, blocks, time, upstream_bandwidth) VALUES (:device_id, :blocks, :time, :upstream_bandwidth) 
	ON DUPLICATE KEY UPDATE blocks = VALUES(blocks), upstream_bandwidth = VALUES(upstream_bandwidth)`, tableNameRetrievalEvent)
	_, err := DB.NamedExecContext(ctx, statement, events)
	return err
}

func GetRetrievalEventsByPage(ctx context.Context, cond *model.RetrievalEvent, option QueryOption) ([]*model.RetrievalEvent, int64, error) {
	var args []interface{}
	where := `WHERE 1=1`
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
