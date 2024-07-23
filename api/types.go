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
