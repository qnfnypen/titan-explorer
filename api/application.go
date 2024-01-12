package api

import (
	"context"
	"database/sql"
	"github.com/Filecoin-Titan/titan/api"
	"github.com/Filecoin-Titan/titan/api/types"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/errors"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/pkg/iptool"
	"net/http"
	"strconv"
	"time"
)

func CreateApplicationHandler(c *gin.Context) {
	params := model.Application{}
	if err := c.BindJSON(&params); err != nil {
		log.Errorf("create application: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	if params.UserID == "" {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	if params.Amount <= 0 {
		params.Amount = 1
	}

	if params.Amount > 10 {
		c.JSON(http.StatusOK, respErrorCode(errors.AmountLimitExceeded, c))
		return
	}

	params.CreatedAt = time.Now()
	params.UpdatedAt = time.Now()
	params.NodeType = 1
	params.Status = dao.ApplicationStatusCreated
	params.Ip = iptool.GetClientIP(c.Request)

	schedulerClient, err := getSchedulerClient(c.Request.Context(), params.AreaID)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.NoSchedulerFound, c))
		return
	}

	err = handleApplication(c.Request.Context(), schedulerClient, &params, int(params.Amount))
	if err != nil {
		log.Errorf("handleApplication %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(nil))
}

func handleApplication(ctx context.Context, schedulerClient api.Scheduler, application *model.Application, amount int) error {
	registration, err := schedulerClient.RequestActivationCodes(ctx, types.NodeType(1), amount)
	if err != nil {
		log.Errorf("register node: %v", err)
		return err
	}
	var results []*model.ApplicationResult
	var deviceInfos []*model.DeviceInfo
	for _, deviceInfo := range registration {
		results = append(results, &model.ApplicationResult{
			UserID:        application.UserID,
			DeviceID:      deviceInfo.NodeID,
			NodeType:      1,
			ApplicationID: application.ID,
			Secret:        deviceInfo.ActivationCode,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		})
		deviceInfos = append(deviceInfos, &model.DeviceInfo{
			UserID:       application.UserID,
			IpLocation:   application.AreaID,
			DeviceID:     deviceInfo.NodeID,
			NodeType:     1,
			DeviceName:   "",
			BindStatus:   "binding",
			ActiveStatus: 0,
			DeviceStatus: "offline",
		})
	}

	if len(results) == 0 {
		log.Infof("update application status: %v", registration)
		return nil
	}

	e := dao.AddApplicationResult(ctx, results)
	if e != nil {
		log.Errorf("create application result: %v", err)
		return e
	}

	err = dao.BulkUpsertDeviceInfo(ctx, deviceInfos)
	if err != nil {
		log.Errorf("add device info: %v", err)
	}

	return nil
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
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  applications,
		"total": total,
	}))
}

func GetApplicationAmountHandler(c *gin.Context) {
	userID := c.Query("user_id")
	option := dao.QueryOption{
		UserID: userID,
	}
	total, err := dao.GetApplicationAmount(c.Request.Context(), option)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusOK, respJSON(JsonObject{
			"total": total,
		}))
		return
	}
	if err != nil {
		log.Errorf("get applications: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"total": total,
	}))
}
