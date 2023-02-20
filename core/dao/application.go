package dao

import (
	"context"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/jmoiron/sqlx"
)

var (
	tableNameApplication       = "application"
	tableNameApplicationResult = "application_result"
)

const (
	ApplicationStatusCreated = iota + 1
	ApplicationStatusFailed
	ApplicationStatusSendEmailFailed
	ApplicationStatusFinished
)

func AddApplication(ctx context.Context, application *model.Application) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`INSERT INTO %s (user_id, email, ip, ip_country, ip_city, amount, node_type, upstream_bandwidth, disk_space, status, created_at, updated_at) 
			VALUES (:user_id, :email, :ip, :ip_country, :ip_city, :amount, :node_type, :upstream_bandwidth, :disk_space, :status, :created_at, :updated_at);`, tableNameApplication),
		application)
	return err
}

func UpdateApplicationStatus(ctx context.Context, id int64, status int) error {
	_, err := DB.ExecContext(ctx, fmt.Sprintf(
		`UPDATE %s SET status = ? where id = ?`, tableNameApplication), status, id)
	return err
}

func GetApplicationList(ctx context.Context, status []int) ([]*model.Application, error) {
	var out []*model.Application
	query, args, err := sqlx.In(fmt.Sprintf(
		`SELECT * FROM %s WHERE status IN (?) LIMIT 50`, tableNameApplication), status)
	if err != nil {
		return nil, err
	}
	query = DB.Rebind(query)
	err = DB.SelectContext(ctx, &out, query, args...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func GetApplicationResults(ctx context.Context, applicationID int64) ([]*model.ApplicationResult, error) {
	var out []*model.ApplicationResult
	err := DB.SelectContext(ctx, &out, fmt.Sprintf(
		`SELECT * FROM %s WHERE application_id = ? `, tableNameApplicationResult), applicationID)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func GetApplicationsByPage(ctx context.Context, option QueryOption) ([]*model.Application, int64, error) {
	var args []interface{}
	where := "WHERE 1=1"

	if option.UserID != "" {
		where += ` AND user_id = ?`
		args = append(args, option.UserID)
	}

	if option.Order != "" && option.OrderField != "" {
		where += fmt.Sprintf(` ORDER BY %s %s`, option.OrderField, option.Order)
	}

	limit := option.PageSize
	offset := option.Page
	if option.PageSize <= 0 {
		limit = 50
	}
	if option.Page > 0 {
		offset = limit * (option.Page - 1)
	}

	var total int64
	var out []*model.Application

	err := DB.GetContext(ctx, &total, fmt.Sprintf(
		`SELECT count(*) FROM %s %s`, tableNameApplication, where,
	), args...)
	if err != nil {
		return nil, 0, err
	}

	err = DB.SelectContext(ctx, &out, fmt.Sprintf(
		`SELECT * FROM %s %s LIMIT %d OFFSET %d`, tableNameApplication, where, limit, offset), args...)
	if err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func AddApplicationResult(ctx context.Context, result []*model.ApplicationResult) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`INSERT INTO %s (user_id, application_id, device_id, secret, node_type, created_at, updated_at) 
			VALUES (:user_id, :application_id, :device_id, :secret, :node_type, :created_at, :updated_at);`, tableNameApplicationResult),
		result)
	return err
}
