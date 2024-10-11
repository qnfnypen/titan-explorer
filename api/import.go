package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/errors"
	"github.com/gnasnik/titan-explorer/pkg/opfie"
)

var (
	ipfsCli, _ = opfie.NewIPFSClient("")
)

type (
	// GetIPFSInfoByCIDSReq 获取ipfs信息的请求
	GetIPFSInfoByCIDSReq struct {
		CIDs string `json:"cids" binding:"required"`
	}
)

// GetIPFSInfoByCIDs 通过cid获取ipfs的信息
// @Summary 导入ipfs文件
// @Description 导入ipfs文件
// @Security ApiKeyAuth
// @Tags import
// @Param req body GetIPFSInfoByCIDSReq true
// @Success 200 {object} JsonObject "{[]{CandidateAddr:"",Token:""}}"
// @Router /api/v1/storage/ipfs_info [post]
func GetIPFSInfoByCIDs(c *gin.Context) {
	var (
		req      GetIPFSInfoByCIDSReq
		nameMaps = make(map[string]string)
	)

	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	// 处理cids
	cidList := strings.Split(req.CIDs, "\n")
	for _, v := range cidList {
		v = strings.TrimSpace(v)
		links, _, _ := ipfsCli.GetInfoByCID(c.Request.Context(), v)
		if len(links) > 0 {
			nameMaps[v] = links[0].Name
		} else {
			nameMaps[v] = ""
		}
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list": nameMaps,
	})) 
}
