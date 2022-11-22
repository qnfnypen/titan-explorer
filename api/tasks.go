package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/errors"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"net/http"
	"strconv"
)

func CreateTaskHandler(c *gin.Context) {
	taskInfo := &model.TaskInfo{}
	taskInfo.UserID = c.Query("userId")
	taskInfo.Cid = c.Query("cid")
	taskInfo.BandwidthUp = c.Query("bandwidth_up")
	taskInfo.BandwidthDown = c.Query("bandwidth_down")
	taskInfo.Status = "new"
	taskInfo.TimeNeed = c.Query("time_need")
	Price := c.Query("price")
	if taskInfo.Cid == "" || taskInfo.TimeNeed == "" || Price == "" {
		c.JSON(http.StatusBadRequest, respError(errors.ErrInvalidParams))
		return
	}
	taskInfo.Price = StrToFloat(Price)

	err := dao.UpsertTaskInfo(c.Request.Context(), taskInfo)
	if err != nil {
		log.Errorf("upsert task info: %v", err)
		c.JSON(http.StatusBadRequest, respError(errors.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, respJSON(nil))
}

func GetTaskInfoHandler(c *gin.Context) {
	taskInfo := &model.TaskInfo{}
	taskInfo.UserID = c.Query("userId")
	taskInfo.Status = c.Query("status")
	if taskInfo.Status == "All" {
		taskInfo.Status = ""
	}
	if taskInfo.UserID == "" {
		c.JSON(http.StatusBadRequest, respError(errors.ErrInvalidParams))
		return
	}

	list, total, err := dao.GetTaskInfoList(c.Request.Context(), taskInfo, dao.QueryOption{})
	if err != nil {
		log.Errorf("get task info: %v", err)
		c.JSON(http.StatusBadRequest, respError(errors.ErrInvalidParams))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  list,
		"count": total,
	}))
}

func GetTaskListHandler(c *gin.Context) {
	var TaskInfoSearch TaskSearch
	TaskInfoSearch.DeviceID = c.Query("device_id")
	TaskInfoSearch.Status = c.Query("status")
	if TaskInfoSearch.DeviceID == "" {
		c.JSON(http.StatusBadRequest, respError(errors.ErrInvalidParams))
		return
	}
	sqlClause := fmt.Sprintf("select date_format(time, '%%Y-%%m-%%d') as date, count(1) as num, sum(file_size) as file_size,sum(price) as price from task_info "+
		"where device_id='%s' and status in ('已完成','已连接') group by date", TaskInfoSearch.DeviceID)
	fmt.Println(sqlClause)
	datas, err := dao.GetQueryDataList(sqlClause)
	if err != nil {
		log.Errorf("QueryClickData error[%v] sqlClause[%s]", err, sqlClause)
		c.JSON(http.StatusBadRequest, respError(errors.ErrInternalServer))
		return
	}
	sqlClause = fmt.Sprintf("select count(1) as num_all, sum(file_size) as file_size_all from task_info "+
		"where device_id='%s' and status in ('已完成','已连接')", TaskInfoSearch.DeviceID)
	fmt.Println(sqlClause)
	count_all, err := dao.GetQueryDataList(sqlClause)
	if err != nil {
		log.Errorf("QueryClickData error[%v] sqlClause[%s]", err, sqlClause)
		c.JSON(http.StatusBadRequest, respError(errors.ErrInternalServer))
		return
	}
	resp := make(map[string]interface{})
	resp["tot_num"] = count_all
	resp["data_list"] = datas

	c.JSON(http.StatusOK, respJSON(resp))
}

func GetTaskDetailHandler(c *gin.Context) {
	var TaskInfoSearch TaskSearch
	TaskInfoSearch.DeviceID = c.Query("device_id")
	date := c.Query("date")
	if TaskInfoSearch.DeviceID == "" {
		c.JSON(http.StatusBadRequest, respError(errors.ErrInvalidParams))
		return
	}
	beginTime := ""
	endTime := ""
	if len(date) == 10 {
		beginTime = date + " 00:00:00"
		endTime = date + " 23:59:59"
	}
	sqlClause := fmt.Sprintf("select date_format(time, '%%Y-%%m-%%d') as date,cid,file_name,file_size,bandwidth_up,bandwidth_down,ip_address,created_at,status from task_info "+
		"where device_id='%s' and status in ('已完成','已连接') and time>='%s' and  time<='%s'", TaskInfoSearch.DeviceID, beginTime, endTime)
	fmt.Println(sqlClause)
	datas, err := dao.GetQueryDataList(sqlClause)
	if err != nil {
		log.Errorf("QueryClickData error[%v] sqlClause[%s]", err, sqlClause)
		c.JSON(http.StatusBadRequest, respError(errors.ErrInternalServer))
		return
	}
	sqlClause = fmt.Sprintf("select count(1) as num from task_info "+
		"where device_id='%s' and status in ('已完成','已连接') and created_at>='%s' and  created_at<='%s'", TaskInfoSearch.DeviceID, beginTime, endTime)
	fmt.Println(sqlClause)
	countAll, err := dao.GetQueryDataList(sqlClause)
	if err != nil {
		log.Errorf("QueryClickData error[%v] sqlClause[%s]", err, sqlClause)
		c.JSON(http.StatusBadRequest, respError(errors.ErrInternalServer))
		return
	}
	resp := make(map[string]interface{})
	resp["tot_num"] = countAll
	resp["data_list"] = datas

	c.JSON(http.StatusOK, respJSON(resp))
}

type TaskSearch struct {
	model.TaskInfo
	PageInfo
}

func Str2Float64(s string) float64 {
	ret, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Error(err.Error())
		return 0.00
	}
	return ret
}

func StrToFloat(str string) float64 {
	v, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return float64(0)
	}
	return v
}

func Str2Int(s string) int {
	ret, err := strconv.Atoi(s)
	if err != nil {
		log.Error(err.Error())
		return 0
	}
	return ret
}
