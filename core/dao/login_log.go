package dao

import (
	"context"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

var tableNameloginLog = "login_log"

func AddLoginLog(ctx context.Context, log *model.LoginLog) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`INSERT INTO %s (login_username, ip_address, login_location, browser, os, status, msg, created_at) VALUES 
		(:login_username, :ip_address, :login_location, :browser, :os, :status, :msg, :created_at);`, tableNameloginLog,
	), log)
	return err
}

func ListLoginLog(ctx context.Context, option QueryOption) ([]*model.LoginLog, int64, error) {
	var args []interface{}
	var total int64
	var out []*model.LoginLog

	limit := option.PageSize
	offset := option.Page
	if option.PageSize <= 0 {
		limit = 50
	}
	if option.Page > 0 {
		offset = limit * (option.Page - 1)
	}

	err := DB.GetContext(ctx, &total, fmt.Sprintf(
		`SELECT count(*) FROM %s`, tableNameDeviceInfo,
	), args)
	if err != nil {
		return nil, 0, err
	}

	err = DB.SelectContext(ctx, &out, fmt.Sprintf(
		`SELECT * FROM %s LIMIT %d OFFSET %d`, tableNameloginLog, limit, offset,
	), args...)
	if err != nil {
		return nil, 0, err
	}

	return out, total, err
}
