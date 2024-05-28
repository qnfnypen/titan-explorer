package dao

import (
	"context"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

func AddDataCollection(ctx context.Context, dc *model.DataCollection) error {
	_, err := DB.NamedExecContext(ctx, `INSERT INTO data_collection(event, value, os, url, ip, created_at) VALUES(:event, :value, :os, :url, :ip, :created_at)`, dc)
	return err
}

func CountPageViewByEvent(ctx context.Context, event model.DataCollectionEvent, code string, option QueryOption) (ipCount, pvCount int64, err error) {
	where := ""
	if option.StartTime != "" {
		where += fmt.Sprintf(" AND created_at >= '%s'", option.StartTime)
	}

	if option.EndTime != "" {
		where += fmt.Sprintf(" AND created_at < '%s'", option.EndTime)
	}

	query := `select count(distinct ip) as ip_count, count(1) as pv from data_collection where event = ? and value = ?` + where

	row := DB.QueryRowxContext(ctx, query, event, code)
	err = row.Scan(&ipCount, &pvCount)

	return
}

func GetPageViewIPCountDailyStat(ctx context.Context, event model.DataCollectionEvent, code string, option QueryOption) ([]*model.DateValue, error) {
	var out []*model.DateValue

	query := `select date_format(created_at, '%Y-%m-%d') as date, count(DISTINCT ip) as value from data_collection where event = ? and value = ? and created_at >= ? and created_at <= ? group by date`
	err := DB.SelectContext(ctx, &out, query, event, code, option.StartTime, option.EndTime)
	if err != nil {
		return nil, err
	}

	return appendDataValueList(out, option.StartTime, option.EndTime), nil
}

func GetPageViewCountDailyStat(ctx context.Context, event model.DataCollectionEvent, code string, option QueryOption) ([]*model.DateValue, error) {
	var out []*model.DateValue

	query := `select date_format(created_at, '%Y-%m-%d') as date, count(1) as value from data_collection where event = ? and value = ? and created_at >= ? and created_at <= ? group by date`
	err := DB.SelectContext(ctx, &out, query, event, code, option.StartTime, option.EndTime)
	if err != nil {
		return nil, err
	}

	return appendDataValueList(out, option.StartTime, option.EndTime), nil
}
