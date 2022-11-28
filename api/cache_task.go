package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/errors"
	"net/http"
	"strconv"
)

func AddCacheTaskHandler(c *gin.Context) {
	carFileCID := c.Query("carfile_cid")
	reliability, _ := strconv.ParseInt(c.Query("reliability"), 10, 64)
	expiration := 365 * 24
	err := schedulerClient.AddCacheTask(c.Request.Context(), carFileCID, int(reliability), expiration)
	if err != nil {
		log.Errorf("add cahce task: %v", err)
		c.JSON(http.StatusBadRequest, respError(errors.ErrInternalServer))
		return
	}
	c.JSON(http.StatusOK, respJSON(nil))
}

func GetCacheTaskInfoHandler(c *gin.Context) {
	carFileCID := c.Query("carfile_cid")
	cacheInfo, err := schedulerClient.GetCacheTaskInfo(c.Request.Context(), carFileCID)
	if err != nil {
		log.Errorf("get cahce task info: %v", err)
		c.JSON(http.StatusBadRequest, respError(errors.ErrInternalServer))
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"cache_task_info": cacheInfo,
	}))
}

func CancelCacheTaskHandler(c *gin.Context) {
	carFileCID := c.Query("carfile_cid")
	err := schedulerClient.CancelCacheTask(c.Request.Context(), carFileCID)
	if err != nil {
		log.Errorf("cancel cahce task: %v", err)
		c.JSON(http.StatusBadRequest, respError(errors.ErrInternalServer))
		return
	}
	c.JSON(http.StatusOK, respJSON(nil))
}

func GetCacheTaskListHandler(c *gin.Context) {
	page, _ := strconv.ParseInt(c.Query("page"), 10, 64)
	size, _ := strconv.ParseInt(c.Query("size"), 10, 64)
	var limit, offset int
	if size <= 0 {
		limit = 50
	}
	if page > 0 {
		offset = int(size * (page - 1))
	}
	resp, err := schedulerClient.ListCacheTasks(c.Request.Context(), offset, limit)
	if err != nil {
		log.Errorf("list cache tasks: %v", err)
		c.JSON(http.StatusBadRequest, respError(errors.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, respJSON(resp))
}

func GetCarFileInfoHandler(c *gin.Context) {
	carFileCID := c.Query("carfile_cid")
	fileInfo, err := schedulerClient.GetCarfileByCID(c.Request.Context(), carFileCID)
	if err != nil {
		log.Errorf("get carfile info: %v", err)
		c.JSON(http.StatusBadRequest, respError(errors.ErrInternalServer))
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"carfile_info": fileInfo,
	}))
}

func RemoveCacheHandler(c *gin.Context) {
	carFileCID := c.Query("carfile_cid")
	err := schedulerClient.RemoveCarfile(c.Request.Context(), carFileCID)
	if err != nil {
		log.Errorf("remove cahce task: %v", err)
		c.JSON(http.StatusBadRequest, respError(errors.ErrInternalServer))
		return
	}
	c.JSON(http.StatusOK, respJSON(nil))
}
