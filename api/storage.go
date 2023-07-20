package api

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/utils"
	"github.com/golang-module/carbon/v2"
	"net/http"
	"time"
)

func GetStorageHourHandler(c *gin.Context) {
	userId := c.Query("user_id")
	//date := c.Query("date")
	start := c.Query("from")
	end := c.Query("to")
	m := queryStorageHourly(userId, start, end)
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"series_data": m,
	}))
}

func GetStorageDailyHandler(c *gin.Context) {
	userId := c.Query("user_id")
	start := c.Query("from")
	end := c.Query("to")
	m := QueryStorageDaily(userId, start, end)
	c.JSON(http.StatusOK, respJSON(JsonObject{
		"series_data": m,
	}))
}

func queryStorageHourly(userId, startTime, endTime string) []*dao.UserInfoRes {
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
		end, _ := time.Parse(utils.TimeFormatDateOnly, endTime)
		end = end.Add(1 * time.Hour).Add(-time.Second)
		option.EndTime = end.Format(utils.TimeFormatDatetime)
	}

	list, err := dao.GetStorageInfoHourList(context.Background(), userId, option)
	if err != nil {
		log.Errorf("queryStorageHourly GetStorageInfoHourList: %v", err)
		return nil
	}
	return list
}

func QueryStorageDaily(userId, startTime, endTime string) []*dao.UserInfoRes {
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
		end, _ := time.Parse(utils.TimeFormatDateOnly, endTime)
		end = end.Add(24 * time.Hour).Add(-time.Second)
		option.EndTime = end.Format(utils.TimeFormatDatetime)
	}
	list, err := dao.GetStorageInfoDaysList(context.Background(), userId, option)
	if err != nil {
		log.Errorf("QueryStorageDaily GetStorageInfoDaysList: %v", err)
		return nil
	}

	return list
}
