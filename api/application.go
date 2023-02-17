package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/errors"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/utils"
	"net/http"
	"net/mail"
	"strconv"
	"time"
)

func CreateApplicationHandler(c *gin.Context) {
	params := model.Application{}
	if err := c.BindJSON(&params); err != nil {
		log.Errorf("create application: %v", err)
		c.JSON(http.StatusOK, respError(errors.ErrInvalidParams))
		return
	}

	_, err := mail.ParseAddress(params.Email)
	if err != nil || params.UserID == "" || params.NodeType == 0 {
		c.JSON(http.StatusOK, respError(errors.ErrInvalidParams))
		return
	}

	if params.Amount <= 0 {
		params.Amount = 1
	}

	if params.Amount > 500 {
		c.JSON(http.StatusOK, respError(errors.ErrAmountLimitExceeded))
		return
	}

	params.CreatedAt = time.Now()
	params.UpdatedAt = time.Now()
	params.Status = dao.ApplicationStatusCreated
	params.Ip = utils.GetClientIP(c.Request)
	if err := dao.AddApplication(c.Request.Context(), &params); err != nil {
		log.Errorf("add application: %v", err)
		c.JSON(http.StatusOK, respError(errors.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, respJSON(nil))
}

func GetApplicationsHandler(c *gin.Context) {
	userID := c.Query("user_id")
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))
	order := c.Query("order")
	orderField := c.Query("order_field")
	option := dao.QueryOption{
		Page:       page,
		PageSize:   pageSize,
		UserID:     userID,
		Order:      order,
		OrderField: orderField,
	}
	applications, total, err := dao.GetApplicationsByPage(c.Request.Context(), option)
	if err != nil {
		log.Errorf("get applications: %v", err)
		c.JSON(http.StatusOK, respError(errors.ErrInternalServer))
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  applications,
		"total": total,
	}))
}
