package api

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/errors"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/pkg/formatter"
	"github.com/golang-module/carbon/v2"
	"github.com/tealeg/xlsx/v3"
	"net/http"
	"strconv"
	"time"
)

type NodeDailyTrend struct {
	X string              `json:"x"`
	Y *model.FullNodeInfo `json:"y"`
}

func GetNodeDailyTrendHandler(c *gin.Context) {
	info := &model.FullNodeInfo{}
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	page, _ := strconv.Atoi(c.Query("page"))
	order := c.Query("order")
	orderField := c.Query("order_field")
	option := dao.QueryOption{
		Page:       page,
		PageSize:   pageSize,
		OrderField: orderField,
		Order:      order,
	}

	list, _, err := dao.GetFullNodeInfoList(c.Request.Context(), info, option)
	if err != nil {
		log.Errorf("get full node info list: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.NotFound, c))
		return
	}

	reverse(list)

	var out []NodeDailyTrend
	for _, item := range list {
		out = append(out, NodeDailyTrend{
			X: item.Time.Format(formatter.TimeFormatDateOnly),
			Y: item,
		})
	}

	c.JSON(http.StatusOK, respJSON(out))
}

func reverse(s []*model.FullNodeInfo) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func GetKOLListHandler(c *gin.Context) {
	page, _ := strconv.ParseInt(c.Query("page"), 10, 64)
	size, _ := strconv.ParseInt(c.Query("size"), 10, 64)
	opt := dao.QueryOption{
		Page:     int(page),
		PageSize: int(size),
	}

	type kolReferral struct {
		*model.KOL
		*model.ReferralCounter
	}

	list, total, err := dao.GetKolList(c.Request.Context(), opt)
	if err != nil {
		log.Errorf("get kols: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	var out []*kolReferral
	for _, kol := range list {
		counter, err := dao.GetUserReferralCounter(c.Request.Context(), kol.UserId)
		if err != nil {
			log.Errorf("GetUserReferralCounter: %v", err)
		}

		if counter == nil {
			counter = &model.ReferralCounter{}
		}

		user, err := dao.GetUserByUsername(c.Request.Context(), kol.UserId)
		if err != nil {
			log.Errorf("GetUserByUsername: %v", err)
		}

		if user != nil {
			counter.ReferrerReward = user.RefereralReward
		}

		out = append(out, &kolReferral{
			KOL:             kol,
			ReferralCounter: counter,
		})
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  out,
		"total": total,
	}))
}

func GetKOLLevelConfigHandler(c *gin.Context) {
	page, _ := strconv.ParseInt(c.Query("page"), 10, 64)
	size, _ := strconv.ParseInt(c.Query("size"), 10, 64)
	opt := dao.QueryOption{
		Page:     int(page),
		PageSize: int(size),
	}

	list, total, err := dao.GetKolLevelConfig(c.Request.Context(), opt)
	if err != nil {
		log.Errorf("get kol level: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  list,
		"total": total,
	}))
}

func AddKOLHandler(c *gin.Context) {
	var params model.KOL
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	_, err := dao.GetKOLLevelByLevel(c.Request.Context(), params.Level)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusOK, respErrorCode(errors.KolLevelNotExist, c))
		return
	}

	kol, err := dao.GetKOLByUserId(c.Request.Context(), params.UserId)
	if err != nil && err != sql.ErrNoRows {
		log.Errorf("get kol: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	if kol != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.KolExist, c))
		return
	}

	if err := dao.AddKOL(c.Request.Context(), &params); err != nil {
		log.Errorf("add kol: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(nil))
}

func AddKOLLevelHandler(c *gin.Context) {
	var params model.KOLLevelConf
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	levels, err := dao.GetKOLLevelByLevel(c.Request.Context(), params.Level)
	if err != nil && err != sql.ErrNoRows {
		log.Errorf("get kol level: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	if levels != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.KolLevelExist, c))
		return
	}

	if err := dao.AddKOLLevelConfig(c.Request.Context(), &params); err != nil {
		log.Errorf("add kol level: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(nil))
}

func UpdateKOLLevelHandler(c *gin.Context) {
	var params model.KOLLevelConf
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	if params.Level <= 0 {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	if err := dao.UpdateKOLLevelConfig(c.Request.Context(), &params); err != nil {
		log.Errorf("update kol level: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(nil))
}

func DeleteKOLLevelHandler(c *gin.Context) {
	var params model.KOLLevelConf
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	if params.Level <= 0 {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	level, err := dao.GetKOLLevelByLevel(c.Request.Context(), params.Level)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusOK, respErrorCode(errors.KolLevelNotExist, c))
		return
	}

	if err != nil {
		log.Errorf("get kol level: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	if err := dao.DeleteKOLLevelConfig(c.Request.Context(), level.Level); err != nil {
		log.Errorf("delete kol level: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(nil))
}

func DeleteKOLHandler(c *gin.Context) {
	var params model.KOL
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	if params.UserId == "" {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	kol, err := dao.GetKOLByUserId(c.Request.Context(), params.UserId)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusOK, respErrorCode(errors.KolNotExist, c))
		return
	}

	if err != nil {
		log.Errorf("get kol: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	if err := dao.DeleteKOL(c.Request.Context(), kol.UserId); err != nil {
		log.Errorf("delete kol: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(nil))
}

func UpdateKOLHandler(c *gin.Context) {
	var params model.KOL
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	if params.UserId == "" {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	if err := dao.UpdateKOL(c.Request.Context(), &params); err != nil {
		log.Errorf("update kol: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(nil))
}

func GetReferralRewardDailyHandler(c *gin.Context) {
	page, _ := strconv.ParseInt(c.Query("page"), 10, 64)
	size, _ := strconv.ParseInt(c.Query("size"), 10, 64)
	start := c.Query("from")
	end := c.Query("to")

	if start == "" {
		start = carbon.Parse("2024-03-10").String()
	}

	userId := c.Query("user_id")
	opt := dao.QueryOption{
		Page:      int(page),
		PageSize:  int(size),
		UserID:    userId,
		StartTime: start,
		EndTime:   end,
	}

	list, total, err := dao.GetUserReferralReward(c.Request.Context(), opt)
	if err != nil {
		log.Errorf("LoadAllUserReferralReward: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  list,
		"total": total,
	}))
}

func ExportReferralRewardDailyHandler(c *gin.Context) {
	start := c.Query("from")
	end := c.Query("to")

	if start == "" {
		start = carbon.Parse("2024-03-10").String()
	}

	userId := c.Query("user_id")
	opt := dao.QueryOption{
		UserID:    userId,
		StartTime: start,
		EndTime:   end,
	}

	list, err := dao.LoadAllUserReferralReward(c.Request.Context(), opt)
	if err != nil {
		log.Errorf("LoadAllUserReferralReward: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	wb := xlsx.NewFile()
	sheet, err := wb.AddSheet("Sheet1")
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	headerRow := sheet.AddRow()
	headerRow.WriteSlice([]string{"父级邀请人", "被邀请人", "今日在线节点", "今日被邀请人奖励", "今日父级邀请人奖励", "创建时间"}, -1)

	for _, item := range list {
		row := sheet.AddRow()
		row.WriteSlice(
			[]interface{}{
				item.ReferrerUserId,
				item.UserId,
				item.DeviceOnlineCount,
				item.Reward,
				item.ReferrerReward,
				item.UpdatedAt.Format(time.DateOnly)},
			-1)
	}

	c.Writer.Header().Add("Content-Type", "application/octet-stream")
	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s", "RewardRecord.xlsx"))

	err = wb.Write(c.Writer)
	if err != nil {
		log.Errorf("write file: %v", err)
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	return
}
