package dao

import (
	"context"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

var tableNameloginLog = "login_log"

func AddLoginLog(ctx context.Context, log *model.LoginLog) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`INSERT INTO %s ("login_username", "ipaddr", "login_location", "browser", "os", "status", "msg", "created_at", "updated_at")
			VALUES (:login_username, :ipaddr, :login_location, :browser, :os, :status, :msg, :created_at, :updated_at);`, tableNameloginLog,
	), log)
	return err
}

func ListLoginLog(ctx context.Context, offset, limit int) ([]*model.LoginLog, int64, error) {
	var args []interface{}
	var total int64
	var out []*model.LoginLog

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
