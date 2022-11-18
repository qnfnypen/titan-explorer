package dao

import (
	"context"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/core/generated/query"
	"gorm.io/gorm"
)

func GetDeviceInfoList(ctx context.Context, cond *model.DeviceInfo, option QueryOption) ([]*model.DeviceInfo, int64, error) {
	limit := option.PageSize
	offset := option.PageSize * (option.Page - 1)

	db := DB.Model(&model.DeviceInfo{}).WithContext(ctx)
	var di []*model.DeviceInfo
	var total int64
	if cond.DeviceID != "" {
		db = db.Where("device_id = ?", cond.DeviceID)
	}
	if cond.UserID != "" {
		db = db.Where("user_id = ?", cond.UserID)
	}
	if cond.DeviceStatus != "" && cond.DeviceStatus != "allDevices" {
		db = db.Where("device_status = ?", cond.DeviceStatus)
	}
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = db.Limit(limit).Offset(offset).Find(&di).Error
	return di, total, err
}

func CreateDeviceInfo(ctx context.Context, deviceInfo *model.DeviceInfo) error {
	db := query.DeviceInfo.WithContext(ctx)

	take, err := db.Where(query.DeviceInfo.DeviceID.Eq(deviceInfo.DeviceID)).Take()
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	if take != nil {
		return nil
	}

	return db.Create(deviceInfo)
}

func UpsertUserDevice(ctx context.Context, deviceInfo *model.DeviceInfo) error {
	var di model.DeviceInfo
	result := DB.Where("device_id = ?", deviceInfo.DeviceID).First(&di)
	if result.RowsAffected <= 0 {
		err := DB.Create(deviceInfo).Error
		if err != nil {
			return err
		}
	} else {
		result := DB.Where("device_id = ?", deviceInfo.DeviceID).Where("user_id = ?", deviceInfo.UserID).First(di)
		if result.RowsAffected <= 0 {
			di.UserID = deviceInfo.UserID
			err := DB.Save(&di).Error
			if err != nil {
				return err
			}
		}
	}
	return nil
}
