// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

import (
	"time"

	"gorm.io/gorm"
)

const TableNameTaskInfo = "task_info"

// TaskInfo mapped from table <task_info>
type TaskInfo struct {
	ID             int64          `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	CreatedAt      time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at" json:"deleted_at"`
	UserID         string         `gorm:"column:user_id;not null" json:"user_id"`
	MinerID        string         `gorm:"column:miner_id;not null" json:"miner_id"`
	DeviceID       string         `gorm:"column:device_id;not null" json:"device_id"`
	FileName       string         `gorm:"column:file_name;not null" json:"file_name"`
	IPAddress      string         `gorm:"column:ip_address;not null" json:"ip_address"`
	Cid            string         `gorm:"column:cid;not null" json:"cid"`
	BandwidthUp    string         `gorm:"column:bandwidth_up;not null" json:"bandwidth_up"`
	BandwidthDown  string         `gorm:"column:bandwidth_down;not null" json:"bandwidth_down"`
	TimeNeed       string         `gorm:"column:time_need;not null" json:"time_need"`
	Time           time.Time      `gorm:"column:time" json:"time"`
	ServiceCountry string         `gorm:"column:service_country;not null" json:"service_country"`
	Region         string         `gorm:"column:region;not null" json:"region"`
	Status         string         `gorm:"column:status;not null" json:"status"`
	Price          float64        `gorm:"column:price;not null" json:"price"`
	FileSize       float64        `gorm:"column:file_size;not null" json:"file_size"`
	DownloadURL    string         `gorm:"column:download_url;not null" json:"download_url"`
}

// TableName TaskInfo's table name
func (*TaskInfo) TableName() string {
	return TableNameTaskInfo
}
