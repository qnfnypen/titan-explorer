package api

import (
	"context"
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/errors"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/pkg/formatter"
	"github.com/golang-module/carbon/v2"
	errs "github.com/pkg/errors"
	"net/http"
	"strconv"
	"time"
)

func GetStorageHourHandler(c *gin.Context) {
	userId := c.Query("user_id")
	start := c.Query("from")
	end := c.Query("to")
	startTime := time.Now()
	areaId := dao.GetAreaID(c.Request.Context(), userId)
	schedulerClient := GetNewScheduler(c.Request.Context(), areaId)
	Info, err := schedulerClient.GetUserInfo(c.Request.Context(), userId)
	if err != nil {
		log.Errorf("api GetUserInfo: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}
	var infos []*model.UserInfo
	var userInfo model.UserInfo
	userInfo.UserId = userId
	userInfo.TotalSize = Info.TotalSize
	userInfo.UsedSize = Info.UsedSize
	userInfo.TotalBandwidth = Info.TotalTraffic
	userInfo.PeakBandwidth = Info.PeakBandwidth
	userInfo.DownloadCount = Info.DownloadCount
	userInfo.Time = startTime
	userInfo.CreatedAt = time.Now()
	userInfo.UpdatedAt = time.Now()
	infos = append(infos, &userInfo)
	e := dao.BulkUpsertStorageHours(c.Request.Context(), infos)
	if err != nil {
		log.Errorf("create user info hour: %v", e)
	}
	m := queryStorageHourly(c.Request.Context(), userId, start, end)
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"series_data": m,
	}))
}

func GetStorageDailyHandler(c *gin.Context) {
	userId := c.Query("user_id")
	start := c.Query("from")
	end := c.Query("to")
	m := QueryStorageDaily(c.Request.Context(), userId, start, end)
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"series_data": m,
	}))
}

func queryStorageHourly(ctx context.Context, userId, startTime, endTime string) []*dao.UserInfoRes {
	option := dao.QueryOption{
		StartTime: startTime,
		EndTime:   endTime,
	}
	if option.StartTime == "" {
		option.StartTime = carbon.Now().StartOfHour().SubHours(25).String()
	}
	if option.EndTime == "" {
		option.EndTime = carbon.Now().String()
	} else {
		end, _ := time.Parse(formatter.TimeFormatDateOnly, endTime)
		end = end.Add(1 * time.Hour).Add(-time.Second)
		option.EndTime = end.Format(formatter.TimeFormatDatetime)
	}

	list, err := dao.GetStorageInfoHourList(ctx, userId, option)
	if err != nil {
		log.Errorf("queryStorageHourly GetStorageInfoHourList: %v", err)
		return nil
	}
	return list
}

func QueryStorageDaily(ctx context.Context, userId, startTime, endTime string) []*dao.UserInfoRes {
	option := dao.QueryOption{
		StartTime: startTime,
		EndTime:   endTime,
	}
	if startTime == "" {
		option.StartTime = carbon.Now().SubDays(14).StartOfDay().String()
	}
	if endTime == "" {
		option.EndTime = carbon.Now().EndOfDay().String()
	} else {
		end, _ := time.Parse(formatter.TimeFormatDateOnly, endTime)
		end = end.Add(24 * time.Hour).Add(-time.Second)
		option.EndTime = end.Format(formatter.TimeFormatDatetime)
	}
	list, err := dao.GetStorageInfoDaysList(ctx, userId, option)
	if err != nil {
		log.Errorf("QueryStorageDaily GetStorageInfoDaysList: %v", err)
		return nil
	}

	return list
}

func ListStorageStats(c *gin.Context) {
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))
	order := c.Query("order")
	orderField := c.Query("order_field")
	option := dao.QueryOption{
		Page:       page,
		PageSize:   pageSize,
		Order:      order,
		OrderField: orderField,
		StartTime:  carbon.Now().SubMinutes(5).String(),
		EndTime:    carbon.Now().String(),
	}

	lastSs, err := dao.GetLastStorageStats(c.Request.Context())
	if err == nil && lastSs != nil {
		option.StartTime = lastSs.Time
	}

	list, count, err := dao.ListStorageStats(c.Request.Context(), -1, option)
	if errs.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusOK, respJSON(JsonObject{
			"storage": model.StorageSummary{},
			"list":    nil,
			"total":   0,
		}))
		return
	}

	if err != nil {
		log.Errorf("ListStorageStats: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	var summary model.StorageSummary
	stats, err := dao.CountStorageStats(c.Request.Context())
	if err != nil {
		log.Errorf("CountStorageStats: %v", err)
	}

	if stats != nil {
		summary = *stats
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"storage": summary,
		"list":    list,
		"total":   count,
	}))
}
