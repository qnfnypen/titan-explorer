package dao

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

const tableNameValidationEvent = "validation_event"

func GetLastValidationEvent(ctx context.Context) (*model.ValidationEvent, error) {
	var out model.ValidationEvent
	query := fmt.Sprintf(`SELECT * FROM %s ORDER BY time DESC LIMIT 1;`, tableNameValidationEvent)
	err := DB.QueryRowxContext(ctx, query).StructScan(&out)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func CreateValidationEvent(ctx context.Context, events []*model.ValidationEvent) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`INSERT INTO %s (device_id, validator_id, blocks, status, time, duration, upstream_traffic)
			VALUES (:device_id, :validator_id, :blocks, :status, :time, :duration, :upstream_traffic);`, tableNameValidationEvent,
	), events)
	return err
}

func GetValidationEventsByPage(ctx context.Context, cond *model.ValidationEvent, option QueryOption) ([]*model.ValidationEvent, int64, error) {
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
	var out []*model.ValidationEvent

	err := DB.GetContext(ctx, &total, fmt.Sprintf(
		`SELECT count(*) FROM %s %s`, tableNameValidationEvent, where,
	), args...)
	if err != nil {
		return nil, 0, err
	}

	err = DB.SelectContext(ctx, &out, fmt.Sprintf(
		`SELECT * FROM %s %s ORDER BY time DESC LIMIT %d OFFSET %d`, tableNameValidationEvent, where, limit, offset,
	), args...)
	if err != nil {
		return nil, 0, err
	}

	return out, total, err
}
