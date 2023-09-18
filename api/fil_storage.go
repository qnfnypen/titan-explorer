package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/errors"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"net/http"
	"strconv"
	"time"
)

const (
	FilcoinMainnetResetTimestamp       = 1602773040
	FilcoinMainnetStartBlock           = 148888
	FilcoinMainnetEpochDurationSeconds = 30
)

func CreateFilStorageHandler(c *gin.Context) {
	var params []*model.FilStorage
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	for i := 0; i < len(params); i++ {
		params[i].StartTime = time.Unix(getTimestampByHeight(params[i].StartHeight), 0)
		params[i].EndTime = time.Unix(getTimestampByHeight(params[i].EndHeight), 0)
		params[i].CreatedAt = time.Now()
		params[i].UpdatedAt = time.Now()
	}

	if err := dao.AddFilStorages(c.Request.Context(), params); err != nil {
		log.Errorf("add fil storage: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(nil))
}

func GetFilStorageListHandler(c *gin.Context) {
	cid := c.Query("cid")
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))

	asset, err := dao.GetAssetByCID(c.Request.Context(), cid)
	if err != nil {
		log.Errorf("GetAssetByCID: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	option := dao.QueryOption{
		Page:     page,
		PageSize: pageSize,
	}

	list, total, err := dao.ListFilStorages(c.Request.Context(), asset.Path, option)
	if err != nil {
		log.Errorf("ListFilStorages: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  list,
		"total": total,
	}))
}

func getTimestampByHeight(height int64) int64 {
	height = height - FilcoinMainnetStartBlock
	if height < 0 {
		return 0
	}

	return FilcoinMainnetResetTimestamp + FilcoinMainnetEpochDurationSeconds*height
}
