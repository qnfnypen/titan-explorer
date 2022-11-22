package dao

import (
	"context"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

var tableNameScheduler = "schedulers"

func GetSchedulers(ctx context.Context) ([]*model.Scheduler, error) {
	var out []*model.Scheduler
	err := DB.SelectContext(ctx, &out, fmt.Sprintf(
		`SELECT * FROM %s`, tableNameScheduler,
	))
	if err != nil {
		return nil, err
	}
	return out, nil
}
