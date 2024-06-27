package dao

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

const (
	BugStateWaiting = 1
	BugStateDone    = 2
)

func BugsCountByBuilderCtx(ctx context.Context, cb squirrel.SelectBuilder) (int64, error) {
	// count builder
	query, args, err := cb.Column("Count(*) as count").From("bugs").ToSql()
	if err != nil {
		return 0, err
	}

	var total int64
	if err := DB.GetContext(ctx, &total, query, args...); err != nil {
		return 0, err
	}

	return total, err
}

func BugsAddCtx(ctx context.Context, b *model.Bug) error {
	query := "insert into bugs (username, code, node_id, email, telegram_id, description, feedback_type, feedback, pics, log, platform, version, state, reward, reward_type, operator, created_at, updated_at) " +
		"values (:username, :code, :node_id, :email, :telegram_id, :description, :feedback_type, :feedback, :pics, :log, :platform, :version, :state, :reward, :reward_type, :operator, :created_at, :updated_at)"
	_, err := DB.NamedExecContext(ctx, query, b)
	return err
}

func BugsListPageCtx(ctx context.Context, page, size int, sb squirrel.SelectBuilder, sel string) ([]*model.Bug, int64, error) {
	var out = make([]*model.Bug, 0)
	var count int64

	if page < 1 {
		page = 1
	}

	query, args, err := sb.Columns(sel).Offset(uint64((page - 1) * size)).Limit(uint64(size)).ToSql()
	if err != nil {
		return nil, 0, err
	}

	if err := DB.SelectContext(ctx, &out, query, args...); err != nil {
		return nil, 0, err
	}

	query, args, err = sb.Column("Count(*) as count").ToSql()
	if err != nil {
		return nil, 0, err
	}

	if err := DB.GetContext(ctx, &count, query, args...); err != nil {
		return nil, 0, err
	}

	return out, count, nil
}

func BugUpdateCtx(ctx context.Context, b *model.Bug) error {
	query := "update bugs set username=:username, node_id=:node_id, email=:email, telegram_id=:telegram_id, description=:description, feedback_type=:feedback_type, " +
		"feedback=:feedback, pics=:pics, log=:log, platform=:platform, version=:version, state=:state, reward=:reward, reward_type=:reward_type, operator=:operator where id=:id"
	_, err := DB.NamedExecContext(ctx, query, b)
	return err
}
