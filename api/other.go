package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/errors"
)

// GetCountryCount 获取在线国家数量
func GetCountryCount(c *gin.Context) {
	count, err := dao.GetCountryCount(c.Request.Context())
	if err != nil {
		log.Errorf("dao GetCountryCount: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.NotFound, c))
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"count": count,
	}))
	return
}
