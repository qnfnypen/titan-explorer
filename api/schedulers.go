package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/errors"
	"net/http"
)

func GetSchedulersHandler(c *gin.Context) {
	schedulers, err := dao.GetSchedulers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NotFound, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"schedulers": schedulers,
	}))
	return
}
