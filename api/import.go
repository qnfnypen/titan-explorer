package api

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Filecoin-Titan/titan/api"
	"github.com/Filecoin-Titan/titan/api/terrors"
	"github.com/Filecoin-Titan/titan/api/types"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/errors"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/pkg/opfie"
)

var (
	ipfsCli, _ = opfie.NewIPFSClient("")
)

type (
	// GetIPFSInfoByCIDSReq 获取ipfs信息的请求
	GetIPFSInfoByCIDSReq struct {
		CIDs    string   `json:"cids" binding:"required"`
		AreaID  []string `json:"area_id"`
		GroupID int64    `json:"group_id"`
	}
)

// SyncIPFSInfoByCIDs 通过cid获取ipfs的信息
// @Summary 导入ipfs文件
// @Description 导入ipfs文件
// @Security ApiKeyAuth
// @Tags import
// @Param req body GetIPFSInfoByCIDSReq true
// @Success 200 {object} JsonObject "{[]{CandidateAddr:"",Token:""}}"
// @Router /api/v1/storage/sync_ipfs [post]
func SyncIPFSInfoByCIDs(c *gin.Context) {
	var (
		req      GetIPFSInfoByCIDSReq
		claims   = jwt.ExtractClaims(c)
		username = claims[identityKey].(string)
		irs      []model.SyncIPFSRecord
		tnow     = time.Now().Unix()

		cids      []string
		totalSize int64
	)

	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	areaIds := getAreaIDsByArea(c, req.AreaID)
	if len(areaIds) == 0 {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}
	// 判断用户是否存在
	user, err := dao.GetUserByUsername(c.Request.Context(), username)
	switch err {
	case sql.ErrNoRows:
		c.JSON(http.StatusOK, respErrorCode(errors.UserNotFound, c))
		return
	case nil:
	default:
		log.Errorf("CreateAssetHandler dao.GetUserByUsername() error: %+v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	// 处理cids，筛选掉用户已经存在的数据
	cidList := strings.Split(req.CIDs, "\n")
	for _, v := range cidList {
		v = strings.TrimSpace(v)
		cids = append(cids, v)
	}
	ncids, err := dao.GetNoExistCIDs(c.Request.Context(), username, cids)
	if err != nil {
		log.Errorf("GetNoExistCIDs error: %v", err)
		c.JSON(http.StatusOK, respErrorCode(int(terrors.UserStorageSizeNotEnough), c))
		return
	}
	for _, v := range ncids {
		name := v
		// 获取此次同步的文件全部大小
		links, size, _ := ipfsCli.GetInfoByCID(c.Request.Context(), v)
		if len(links) > 0 && links[0].Name != "" {
			name = links[0].Name
		}
		totalSize += int64(size)
		irs = append(irs, model.SyncIPFSRecord{Username: username, CID: v, Timestamp: tnow, AreaID: areaIds[0],
			GroupID: req.GroupID, Size: int64(size), Name: name})
	}
	// 判断用户存储空间是否够用
	if user.TotalStorageSize-user.UsedStorageSize < totalSize {
		c.JSON(http.StatusOK, respErrorCode(int(terrors.UserStorageSizeNotEnough), c))
		return
	}

	// 调用调度器同步ipfs文件
	sc, err := getSchedulerClient(c.Request.Context(), areaIds[0])
	if err != nil {
		log.Errorf("CreateAssetHandler getSchedulerClient error: %v", err)
		if webErr, ok := err.(*api.ErrWeb); ok {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"err":  webErr.Code,
				"msg":  webErr.Message,
				"Log":  areaIds[0],
			})
			return
		}
	}
	err = sc.PullAsset(c.Request.Context(), &types.PullAssetReq{CIDs: cids})
	if err != nil {
		log.Errorf("PullAssetReq error: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	// 将数据增加到sync_ipfs_record表中
	err = dao.AddIPFSRecords(c.Request.Context(), irs)
	if err != nil {
		log.Errorf("AddIPFSRecords error: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"msg": "success",
	}))
}

// GetIPFSRecords 获取用户上传ipfs的记录表
// @Summary 获取用户上传ipfs的记录表
// @Description 获取用户上传ipfs的记录表
// @Security ApiKeyAuth
// @Tags import
// @Param page query int true
// @Param size query int true
// @Success 200 {object} JsonObject "{"msg":""}"
// @Router /api/v1/storage/sync_ipfs [get]
func GetIPFSRecords(c *gin.Context) {
	var (
		claims   = jwt.ExtractClaims(c)
		username = claims[identityKey].(string)
	)

	page, _ := strconv.Atoi(c.Query("page"))
	size, _ := strconv.Atoi(c.Query("size"))
	if page == 0 {
		page = 1
	}
	if size == 0 {
		size = 10
	}

	total, list, err := dao.GetIPFSRecordsByUsername(c.Request.Context(), username, page, size)
	if err != nil {
		log.Errorf("GetIPFSRecordsByUsername error: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"total": total,
		"list":  list,
	}))
}
