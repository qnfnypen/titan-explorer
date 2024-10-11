package opasynq

const (
	// TypeAssetGroupID 文件唯一的组id
	TypeAssetGroupID = "trans:tc_id"
	// TypeDeleteAssetOperation 从调度器删除文件操作
	TypeDeleteAssetOperation = "operation:delete:asset"
)

type (
	// AssetGroupPayload 文件组载体
	AssetGroupPayload struct {
		UserID  string  `json:"user_id"`
		GroupID []int64 `json:"group_id"`
	}

	// DeleteAssetPayload 删除
	DeleteAssetPayload struct {
		CID    string `json:"cid"`
		AreaID string `json:"area_id"`
	}
)
