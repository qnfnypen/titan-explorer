package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/errors"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"net/http"
	"strconv"
)

func GetCacheListHandler(c *gin.Context) {
	info := &model.CacheEvent{
		DeviceID: c.Query("device_id"),
	}
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))
	option := dao.QueryOption{
		Page:       page,
		PageSize:   pageSize,
		OrderField: "time",
		EndTime:    "DESC",
	}

	list, total, err := dao.GetCacheEventsByPage(c.Request.Context(), info, option)
	if err != nil {
		log.Errorf("get device info list: %v", err)
		c.JSON(http.StatusBadRequest, respError(errors.ErrNotFound))
		return
	}
	
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  list,
		"total": total,
	}))
}
