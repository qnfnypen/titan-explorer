package api

import (
	ge "errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/errors"
)

func checkHealth(c *gin.Context) {
	// tws platform check
	err := mDB.Health()
	if err != nil {
		c.JSON(http.StatusBadRequest, respError(errors.InternalServer, ge.New("数据库连接失败")))
		return
	}
	err = kubMgr.Health()
	if err != nil {
		c.JSON(http.StatusBadRequest, respError(errors.InternalServer, ge.New("kub token过期")))
		return
	}
	err = orderMgr.CheckChainMgr()
	if err != nil {
		c.JSON(http.StatusBadRequest, respError(errors.InternalServer, ge.New("chain connect error")))
		return
	}

	c.JSON(http.StatusOK, respJSON(gin.H{
		"msg": "success",
	}))
}
