package opasynq

const (
	// TypeAssetGroupID 文件唯一的组id
	TypeAssetGroupID = "trans:tc_id"
)

type (
	// AssetGroupPayload 文件组载体
	AssetGroupPayload struct {
		UserID  string  `json:"user_id"`
		GroupID []int64 `json:"group_id"`
	}
)
