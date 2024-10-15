package opasynq

import "time"

const (
	// TypeAssetGroupID 文件唯一的组id
	TypeAssetGroupID = "trans:tc_id"

	//
	TaskTypeAssetUploadedNotify = "task:asset:upload:notify"

	//
	TaskTypeAssetDeleteNotify = "task:asset:delete:notify"

	// TypeDeleteAssetOperation 从调度器删除文件操作
	TypeDeleteAssetOperation = "operation:delete:asset"
)

const (
	TaskQueueExplorer = "explorer"

	TaskQueueTenant = "tenant"
)

type (
	// AssetGroupPayload 文件组载体
	AssetGroupPayload struct {
		UserID  string  `json:"user_id"`
		GroupID []int64 `json:"group_id"`
	}

	// AssetUploadNotifyPayload 上传文件完成通知
	AssetUploadNotifyPayload struct {
		ExtraID  string // 外部文件ID
		TenantID string // 租户ID
		UserID   string // 上传者ID

		AssetName      string
		AssetCID       string
		AssetType      string
		AssetSize      int64
		GroupID        int64
		CreatedTime    time.Time
		AssetDirectUrl string // 上传完成的直接地址

		NotifyUrl  string
		RetryCount int
	}

	AssetDeleteNotifyPayload struct {
		ExtraID  string // 外部文件ID
		TenantID string // 租户ID
		UserID   string // 上传者ID

		AssetCID string
	}

	// DeleteAssetPayload 删除
	DeleteAssetPayload struct {
		CID    string `json:"cid"`
		AreaID string `json:"area_id"`
	}
)
