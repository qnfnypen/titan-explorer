package dao

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

const tableNameTenants = "tenants"

var (
	TenantStateActive   = "active"
	TenantStateInactive = "inactive"
)

func GetTenantByBuilder(ctx context.Context, sb squirrel.SelectBuilder) (*model.Tenant, error) {
	var tenant model.Tenant
	query, args, err := sb.From(tableNameTenants).Limit(1).ToSql()
	if err != nil {
		return nil, err
	}

	err = DB.SelectContext(ctx, &tenant, query, args...)
	return &tenant, err
}

// func LoadTenantApiKeyPair(ctx context.Context, tenantID string) (*model.Tenant, string, string, error) {
// 	var tenant model.Tenant
// 	query, args, err := squirrel.Select("*").From(tableNameTenants).Where("tanant_id = ?", tenantID).Limit(1).ToSql()
// 	if err != nil {
// 		return nil, "", "", err
// 	}

// 	err = DB.GetContext(ctx, &tenant, query, args...)
// 	if err != nil {
// 		return nil, "", "", err
// 	}

// }
