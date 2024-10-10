package opasynq

import "time"

const (
	// TypeAssetGroupID 文件唯一的组id
	TypeAssetGroupID = "trans:tc_id"

	//
	TaskTypeAssetUploadedNotify = "task:asset:upload:notify"
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

		AssetName   string
		AssetCID    string
		AssetType   string
		AssetSize   int64
		GroupID     int64
		CreatedTime time.Time

		NotifyUrl  string
		RetryCount int
	}
)
