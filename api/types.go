package api

// CreateAssetReq 上传文件请求参数
type CreateAssetReq struct {
	AssetName string
	AssetCID  string
	NodeID    string
	UserID    string
	AssetType string
	AssetSize int64
	GroupID   int
}

// MoveNodeReq 节点迁移请求参数
type MoveNodeReq struct {
	NodeID     string `json:"node_id" binding:"required"`
	FromAreaID string `json:"from_area_id" binding:"required"`
	ToAreaID   string `json:"to_area_id" binding:"required"`
}
