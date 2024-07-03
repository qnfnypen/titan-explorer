package dao

import (
	"context"
	"database/sql"

	"github.com/gnasnik/titan-explorer/core/generated/model"
)

func AcmeRecord(ctx context.Context) (*model.Acme, error) {
	var acme model.Acme
	query := "select * from acme order by expire_at desc limit 1"
	err := DB.GetContext(ctx, &acme, query)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &acme, nil
}

func AcmeAdd(ctx context.Context, acme *model.Acme) error {
	_, err := DB.NamedExecContext(ctx, `
		insert into acme (certificate, private_key, created_at, expire_at) values (:certificate, :private_key, :created_at, :expire_at)
	`, acme)
	return err
}
