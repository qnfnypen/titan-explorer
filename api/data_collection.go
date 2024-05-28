package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/errors"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/pkg/iptool"
	"github.com/mssola/user_agent"
	"net/http"
	"time"
)

func DataCollectionHandler(c *gin.Context) {
	var params model.DataCollection
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	userAgent := c.Request.Header.Get("User-Agent")
	ua := user_agent.New(userAgent)

	params.Os = ua.OS()
	params.IP = iptool.GetClientIP(c.Request)
	params.CreatedAt = time.Now()

	err := dao.AddDataCollection(c.Request.Context(), &params)
	if err != nil {
		log.Errorf("AddDataCollection: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(nil))
}
