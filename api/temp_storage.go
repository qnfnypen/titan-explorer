package api

import (
	"net/http"

	"github.com/Filecoin-Titan/titan/api"
	"github.com/Filecoin-Titan/titan/api/types"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/errors"
)

type (
	// UploadTempFileReq 上传临时文件
	UploadTempFileReq struct {
		AssetName string   `json:"asset_name" binding:"required"`
		AssetCID  string   `json:"asset_cid" binding:"required"`
		NodeID    string   `json:"node_id"`
		AssetSize int64    `json:"asset_size" binding:"required"`
		AreaIDs   []string `json:"area_ids" binding:"required"`
	}
)

// UploadTmepFile 未登陆用户受限制上传文件
func UploadTmepFile(c *gin.Context) {
	var req UploadTempFileReq

	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}
	if len(req.AreaIDs) > 3 {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	schCli, err := getSchedulerClient(c.Request.Context(), req.AreaIDs[0])
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	createAssetRsp, err := schCli.CreateAsset(c.Request.Context(), &types.CreateAssetReq{
		AssetCID:      req.AssetCID,
		AssetSize:     req.AssetSize,
		NodeID:        req.NodeID,
		ReplicaCount:  20,
		ExpirationDay: 1,
	})
	if err != nil {
		log.Errorf("CreateAssetHandler CreateAsset error: %v", err)
		if webErr, ok := err.(*api.ErrWeb); ok {
			c.JSON(http.StatusOK, respErrorCode(webErr.Code, c))
			return
		}
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	rsp := make([]JsonObject, len(createAssetRsp.List))
	for i, v := range createAssetRsp.List {
		rsp[i] = JsonObject{"CandidateAddr": v.UploadURL, "Token": v.Token}
	}

	c.JSON(http.StatusOK, respJSON(rsp))
}
