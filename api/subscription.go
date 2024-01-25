package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/errors"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"net/http"
)

func SubscribeHandler(c *gin.Context) {
	var param model.Subscription

	if err := c.BindJSON(&param); err != nil {
		log.Errorf("bind json: %v", err)
		c.JSON(http.StatusBadRequest, respErrorCode(errors.InvalidParams, c))
		return
	}

	if param.Name == "" || param.Email == "" {
		c.JSON(http.StatusBadRequest, respErrorCode(errors.InvalidParams, c))
		return
	}

	if err := dao.AddSubscription(c.Request.Context(), &param); err != nil {
		log.Errorf("add subscription: %v", err)
		c.JSON(http.StatusBadRequest, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(nil))
}
