package dao

import (
	"context"
	"fmt"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"strconv"
)

func GetIncomeDailyHourList(ctx context.Context, cond *model.HourDaily, option QueryOption) ([]*model.HourDaily, int64, error) {
	limit := option.PageSize
	offset := option.PageSize * (option.Page - 1)

	db := DB.Model(&model.HourDaily{}).WithContext(ctx)
	var hd []*model.HourDaily
	var total int64
	if cond.DeviceID != "" {
		db = db.Where("device_id = ?", cond.DeviceID)
	}
	if cond.UserID != "" {
		db = db.Where("user_id = ?", cond.UserID)
	}
	if !option.StartTime.IsZero() {
		db = db.Where("time >= ?", option.StartTime)
	}
	if !option.EndTime.IsZero() {
		db = db.Where("time <= ?", option.EndTime)
	}
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Limit(limit).Offset(offset).Find(&hd).Error
	if err != nil {
		return nil, 0, err
	}

	return hd, total, nil
}

func GetIncomeDailyList(ctx context.Context, cond *model.IncomeDaily, option QueryOption) ([]*model.IncomeDaily, int64, error) {
	limit := option.PageSize
	offset := option.PageSize * (option.Page - 1)

	db := DB.Model(&model.IncomeDaily{}).WithContext(ctx)
	var incomeDailies []*model.IncomeDaily
	var total int64
	if cond.DeviceID != "" {
		db = db.Where("device_id = ?", cond.DeviceID)
	}
	if cond.UserID != "" {
		db = db.Where("user_id = ?", cond.UserID)
	}
	if !option.StartTime.IsZero() {
		db = db.Where("time >= ?", option.StartTime)
	}
	if !option.EndTime.IsZero() {
		db = db.Where("time <= ?", option.EndTime)
	}
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Limit(limit).Offset(offset).Find(&incomeDailies).Error
	if err != nil {
		return nil, 0, err
	}

	return incomeDailies, total, nil
}

func GetIncomeAllList(ctx context.Context, cond *model.IncomeDaily, option QueryOption) []map[string]interface{} {
	sqlClause := fmt.Sprintf("select date_format(time, '%%Y-%%m-%%d') as date, , sum(income) as income from income_daily "+
		"where device_id='%s' and time>='%s' and time<='%s' group by date", cond.DeviceID, option.StartTime, option.EndTime)
	if cond.UserID != "" {
		sqlClause = fmt.Sprintf("select date_format(time, '%%Y-%%m-%%d') as date, sum(income) as income from income_daily "+
			"where user_id='%s' and time>='%s' and time<='%s' group by date", cond.UserID, option.StartTime, option.EndTime)
	}
	datas, err := GetQueryDataList(sqlClause)
	if err != nil {
		return nil
	}
	var mapIncomeList []map[string]interface{}
	for _, data := range datas {
		mapIncome := make(map[string]interface{})
		mapIncome["date"] = data["date"]
		mapIncome["income"], _ = strconv.ParseFloat(data["income"], 10)
		mapIncomeList = append(mapIncomeList, mapIncome)
	}
	return mapIncomeList
}
