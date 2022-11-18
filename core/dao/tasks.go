package dao

import (
	"context"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

func UpsertTaskInfo(ctx context.Context, task *model.TaskInfo) error {
	var ti model.TaskInfo
	result := DB.Where("cid = ?", task.Cid).First(&ti)
	if result.RowsAffected <= 0 {
		err := DB.Create(&task).Error
		if err != nil {
			return err
		}
	} else {
		ti.BandwidthUp = task.BandwidthUp
		ti.BandwidthDown = task.BandwidthDown
		ti.TimeNeed = task.TimeNeed
		ti.Price = task.Price
		err := DB.Save(&ti).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func GetTaskInfoList(ctx context.Context, cond *model.TaskInfo, option QueryOption) ([]*model.TaskInfo, int64, error) {
	limit := option.PageSize
	offset := option.PageSize * (option.Page - 1)

	db := DB.Model(&model.TaskInfo{}).WithContext(ctx)
	var ti []*model.TaskInfo
	var total int64
	if cond.Status != "" && cond.Status != "All" {
		db = db.Where("status = ?", cond.Status)
	}
	if cond.Cid != "" {
		db = db.Where("cid = ?", cond.Cid)
	}
	if cond.UserID != "" {
		db = db.Where("user_id = ?", cond.UserID)
	}
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = db.Limit(limit).Offset(offset).Find(&ti).Error
	if err != nil {
		return nil, 0, err
	}

	return ti, total, nil
}
