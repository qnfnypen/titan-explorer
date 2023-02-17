package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/errors"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/utils"
	"net/http"
	"strconv"
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
		c.JSON(http.StatusBadRequest, respError(errors.ErrNotFound))
		return
	}

	reverse(list)

	var out []NodeDailyTrend
	for _, item := range list {
		out = append(out, NodeDailyTrend{
			X: item.Time.Format(utils.TimeFormatYMD),
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
