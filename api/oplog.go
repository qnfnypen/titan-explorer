package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/errors"
	"net/http"
	"strconv"
)

func GetLoginLogHandler(c *gin.Context) {
	page, _ := strconv.ParseInt(c.Query("page"), 10, 64)
	size, _ := strconv.ParseInt(c.Query("size"), 10, 64)
	opt := dao.QueryOption{
		Page:     int(page),
		PageSize: int(size),
	}
	list, total, err := dao.ListLoginLog(c.Request.Context(), opt)
	if err != nil {
		c.JSON(http.StatusBadRequest, respError(errors.ErrInternalServer))
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  list,
		"total": total,
	}))
}

func GetOperationLogHandler(c *gin.Context) {
	page, _ := strconv.ParseInt(c.Query("page"), 10, 64)
	size, _ := strconv.ParseInt(c.Query("size"), 10, 64)
	opt := dao.QueryOption{
		Page:     int(page),
		PageSize: int(size),
	}
	list, total, err := dao.ListOperationLog(c.Request.Context(), opt)
	if err != nil {
		c.JSON(http.StatusBadRequest, respError(errors.ErrInternalServer))
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  list,
		"total": total,
	}))
}
