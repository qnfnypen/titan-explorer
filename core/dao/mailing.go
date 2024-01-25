package dao

import (
	"context"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

const tableNameSubscription = "subscription"

func AddSubscription(ctx context.Context, subscription *model.Subscription) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`INSERT INTO %s (company, name, email, telegram, wechat, location, storage, calculation, bandwidth, join_testnet, idle_resource_percentages, subscribe, source, created_at, updated_at) 
			VALUES (:company, :name, :email, :telegram, :wechat, :location, :storage, :calculation, :bandwidth, :join_testnet, :idle_resource_percentages, :subscribe, :source, now(), now());`, tableNameSubscription),
		subscription)
	return err
}
