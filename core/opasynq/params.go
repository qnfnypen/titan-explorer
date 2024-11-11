package opasynq

import (
	"time"

	"github.com/gnasnik/titan-explorer/core/generated/model"
)

const (
	// TypeAssetGroupID 文件唯一的组id
	TypeAssetGroupID = "trans:tc_id"

	//
	TaskTypeAssetUploadedNotify = "task:asset:upload:notify"

	//
	TaskTypeAssetDeleteNotify = "task:asset:delete:notify"

	// TypeDeleteAssetOperation 从调度器删除文件操作
	TypeDeleteAssetOperation = "operation:delete:asset"

	// TypeSyncIPFSRecord 同步ipfs文件记录
	TypeSyncIPFSRecord = "sync:ipfs"
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
		Area           string // 用于获取AssetDirectUrl的区域

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

	// IPFSRecordPayload ipfs文件记录
	IPFSRecordPayload struct {
		AreaID string          `json:"area_id"`
		Info   model.UserAsset `json:"info"`
	}
)
