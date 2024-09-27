package dao

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

func GetTenantByBuilder(ctx context.Context, sb squirrel.SelectBuilder) (*model.Tenant, error) {
	var tenant model.Tenant
	query, args, err := sb.From(tableNameUser).Limit(1).ToSql()
	if err != nil {
		return nil, err
	}

	err = DB.SelectContext(ctx, &tenant, query, args...)
	return &tenant, err
}
