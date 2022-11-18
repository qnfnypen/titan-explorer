package dao

import (
	"context"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/core/generated/query"
)

func GetSchedulers(ctx context.Context) ([]*model.Scheduler, error) {
	return query.Scheduler.WithContext(ctx).Find()
}

func AddScheduler(ctx context.Context, scheduler *model.Scheduler) error {
	return query.Scheduler.WithContext(ctx).Create(scheduler)
}

func DeleteScheduler(ctx context.Context, id int64) error {
	scheduler, err := query.Scheduler.WithContext(ctx).Select(query.Scheduler.ID.Eq(id)).Take()
	if err != nil {
		return err
	}

	_, err = query.Scheduler.WithContext(ctx).Delete(scheduler)
	return err
}
