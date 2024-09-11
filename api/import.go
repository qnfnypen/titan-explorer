package api

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
// func GetIPFSInfoByCIDs(c *gin.Context) {
// 	var (
// 		req     GetIPFSInfoByCIDSReq
// 		cidList []string
// 	)

// 	err := c.ShouldBindJSON(&req)
// 	if err != nil {
// 		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
// 		return
// 	}

// 	// 处理cids
// 	cidList = strings.Split(req.CIDs, "\n")
// 	_ = cidList
// }
