package dao

import (
	"context"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"time"
)

const tableNameRetrieveEvent = "retrieve_event"

func CreateRetrieveEvent(ctx context.Context, events []*model.LoginLog) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`INSERT INTO %s (device_id, blocks, time, upstream_bandwidth)
			VALUES (:device_id, :blocks, :time, :upstream_bandwidth);`, tableNameRetrieveEvent,
	), events)
	return err
}

func groupDevicesAndInsert(ctx context.Context, startTime, endTime time.Time) error {
	queryStatement := fmt.Sprintf(`
INSERT INTO %s(device_id, carfile_cid, block_size, blocks, time)
select device_id, created_at, (c.retrieve_count - c.pre_retrieve_count) as retrieve_count, (c.upstream_traffic - c.pre_upstream_traffic) as upstream_traffic  
from (
	select device_id, retrieve_count , upstream_traffic , created_at, 
	@a.retrieve_count AS pre_retrieve_count,
	@a.upstream_traffic AS pre_upstream_traffic,
	@a.retrieve_count := a.retrieve_count, 
	@a.upstream_traffic := a.upstream_traffic  
	from %s a ,
	(SELECT @a.retrieve_count := 0, @a.upstream_traffic := 0 ) b %s 
) c where (c.retrieve_count - c.pre_retrieve_count) > 0 `, tableNameRetrieveEvent, tableNameDeviceInfoHour)

	_, err := DB.ExecContext(ctx, queryStatement, startTime, endTime)
	return err
}

func GetRetrieveEventsByPage(ctx context.Context, cond *model.CacheEvent, option QueryOption) ([]*model.RetrieveEvent, int64, error) {
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
	var out []*model.RetrieveEvent

	err := DB.GetContext(ctx, &total, fmt.Sprintf(
		`SELECT count(*) FROM %s %s`, tableNameRetrieveEvent, where,
	), args...)
	if err != nil {
		return nil, 0, err
	}

	err = DB.SelectContext(ctx, &out, fmt.Sprintf(
		`SELECT * FROM %s %s LIMIT %d OFFSET %d`, tableNameRetrieveEvent, where, limit, offset,
	), args...)
	if err != nil {
		return nil, 0, err
	}

	return out, total, err
}
