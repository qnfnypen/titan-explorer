package dao

import (
	"context"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/core/generated/query"
)

func AddLoginLog(ctx context.Context, log *model.LoginLog) error {
	return query.LoginLog.WithContext(ctx).Create(log)
}

func ListLoginLog(ctx context.Context, offset, limit int) ([]*model.LoginLog, int64, error) {
	return query.LoginLog.WithContext(ctx).FindByPage(offset, limit)
}
