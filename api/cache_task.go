package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/errors"
	"github.com/gnasnik/titan-explorer/utils"
	"github.com/linguohua/titan/api"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type CreateCacheParams struct {
	CarfileCid  string `json:"carfile_cid"`
	Reliability int    `json:"reliability"`
	ExpiredTime string `json:"expired_time"`
}

func AddCacheTaskHandler(c *gin.Context) {
	params := &CreateCacheParams{}
	err := c.BindJSON(params)
	if err != nil {
		log.Errorf("bind json: %v", err)
		c.JSON(http.StatusOK, respError(errors.ErrInvalidParams))
		return
	}
	expiredTime, _ := time.Parse(utils.TimeFormatDateOnly, params.ExpiredTime)
	info := &api.CacheCarfileInfo{
		NeedReliability: params.Reliability,
		CarfileCid:      strings.TrimSpace(params.CarfileCid),
		ExpiredTime:     expiredTime,
	}
	err = schedulerAdmin.CacheCarfile(c.Request.Context(), info)
	if err != nil {
		log.Errorf("api AddCacheTask: %v", err)
		c.JSON(http.StatusOK, respError(errors.NewError(err.Error())))
		return
	}
	c.JSON(http.StatusOK, respJSON(nil))
}

func GetCacheTaskInfoHandler(c *gin.Context) {
	carFileCID := c.Query("carfile_cid")
	cacheInfo, err := schedulerAdmin.GetCarfileRecordInfo(c.Request.Context(), carFileCID)
	if err != nil {
		log.Errorf("api GetCarfileRecordInfo: %v", err)
		c.JSON(http.StatusOK, respError(errors.ErrInternalServer))
		return
	}
	c.JSON(http.StatusOK, respJSON(cacheInfo))
}

func DeleteCacheTaskHandler(c *gin.Context) {
	params := &CreateCacheParams{}
	err := c.BindJSON(params)
	if err != nil {
		log.Errorf("bind json: %v", err)
		c.JSON(http.StatusOK, respError(errors.ErrInvalidParams))
		return
	}
	err = schedulerAdmin.RemoveCarfile(c.Request.Context(), params.CarfileCid)
	if err != nil {
		log.Errorf("api RemoveCarfile: %v", err)
		c.JSON(http.StatusOK, respError(errors.ErrInternalServer))
		return
	}
	c.JSON(http.StatusOK, respJSON(nil))
}

func DeleteCacheTaskByDeviceHandler(c *gin.Context) {
	carFileCID := c.Query("carfile_cid")
	deviceID := c.Query("device_id")
	err := schedulerAdmin.RemoveCache(c.Request.Context(), carFileCID, deviceID)
	if err != nil {
		log.Errorf("api RemoveCache: %v", err)
		c.JSON(http.StatusOK, respError(errors.ErrInternalServer))
		return
	}
	c.JSON(http.StatusOK, respJSON(nil))
}

func GetCacheTaskListHandler(c *gin.Context) {
	page, _ := strconv.ParseInt(c.Query("current"), 10, 64)
	//size, _ := strconv.ParseInt(c.Query("size"), 10, 64)
	resp, err := schedulerAdmin.ListCarfileRecords(c.Request.Context(), int(page))
	if err != nil {
		log.Errorf("api ListCarfileRecords: %v", err)
		c.JSON(http.StatusOK, respError(errors.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  resp.CarfileRecords,
		"total": resp.Cids,
	}))
}

func GetCarFileInfoHandler(c *gin.Context) {
	carFileCID := c.Query("carfile_cid")
	fileInfo, err := schedulerAdmin.GetCarfileRecordInfo(c.Request.Context(), carFileCID)
	if err != nil {
		log.Errorf("get carfile info: %v", err)
		c.JSON(http.StatusOK, respError(errors.ErrInternalServer))
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"carfile_info": fileInfo,
	}))
}

func RemoveCacheHandler(c *gin.Context) {
	carFileCID := c.Query("carfile_cid")
	err := schedulerAdmin.RemoveCarfile(c.Request.Context(), carFileCID)
	if err != nil {
		log.Errorf("remove cahce task: %v", err)
		c.JSON(http.StatusOK, respError(errors.ErrInternalServer))
		return
	}
	c.JSON(http.StatusOK, respJSON(nil))
}
