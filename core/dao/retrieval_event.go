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

func GroupDevicesAndCreateRetrievalEvents(ctx context.Context, startTime, endTime time.Time) error {
	queryStatement := fmt.Sprintf(`
INSERT INTO %s(device_id, blocks, time, upstream_bandwidth)
select device_id, (c.retrieval_count - c.pre_retrieval_count) as blocks, time, (c.upstream_traffic - c.pre_upstream_traffic) as upstream_bandwidth  
from (
	select device_id, retrieval_count, upstream_traffic , time, 
	@a.retrieval_count AS pre_retrieval_count,
	@a.upstream_traffic AS pre_upstream_traffic,
	@a.retrieval_count := a.retrieval_count, 
	@a.upstream_traffic := a.upstream_traffic  lo
	from %s a ,
	(SELECT @a.retrieval_count := 0, @a.upstream_traffic := 0 ) b where time > ? and time < ?  and retrieval_count <> @a.retrieval_count
) c where (c.retrieval_count - c.pre_retrieval_count) > 0 `, tableNameRetrievalEvent, tableNameDeviceInfoHour)

	_, err := DB.ExecContext(ctx, queryStatement, startTime, endTime)
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
