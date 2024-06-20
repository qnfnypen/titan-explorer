package api

import (
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/errors"
)

// Test1NodeController test1节点管理
type Test1NodeController struct{}

type (
	// GetTest1NodeReq 获取test1节点请求
	GetTest1NodeReq struct {
		Kind int64  `form:"kind" binding:"required"`
		Page uint64 `form:"page" binding:"required"`
		Size uint64 `form:"size" binding:"required"`
	}

	// UpdateDeviceInfoReq 修改节点信息请求
	UpdateDeviceInfoReq struct {
		DeviceID   []string `json:"deviceId" binding:"required"` // 设备id
		DeviceName string   `json:"deviceName"`                  // 设备备注
	}
)

// GetNodes 获取节点信息 kind:1-在线 2-故障 3-离线 4-删除
func (tc *Test1NodeController) GetNodes(c *gin.Context) {
	var req GetTest1NodeReq

	err := c.BindQuery(&req)
	if err != nil || req.Kind <= 0 || req.Kind > 4 {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}
	if req.Page == 0 || req.Size == 0 {
		req.Page = 1
		req.Size = 10
	}

	total, infos, err := dao.GetTest1Nodes(c, req.Kind, req.Page, req.Size)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"total": total,
		"list":  infos,
	}))
}

// UpdateDeviceName 修改节点备注
func (tc *Test1NodeController) UpdateDeviceName(c *gin.Context) {
	var req UpdateDeviceInfoReq

	err := c.BindJSON(&req)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	id := req.DeviceID[0]

	err = dao.UpdateTest1DeviceName(c, id, req.DeviceName)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"msg": "ok",
	}))
}

// DeleteOffLineNode 删除离线节点
func (tc *Test1NodeController) DeleteOffLineNode(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	username := claims[identityKey].(string)
	var req UpdateDeviceInfoReq

	err := c.BindJSON(&req)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	err = dao.DeleteOfflineDevice(c, req.DeviceID, username)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"msg": "ok",
	}))
}

// MoveBackDeletedNode 移回删除的节点
func (tc *Test1NodeController) MoveBackDeletedNode(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	username := claims[identityKey].(string)
	var req UpdateDeviceInfoReq

	err := c.BindJSON(&req)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InvalidParams, c))
		return
	}

	err = dao.MoveBackDeletedDevice(c, req.DeviceID, username)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"msg": "ok",
	}))
}

// GetNodeNums 获取节点数量
func (tc *Test1NodeController) GetNodeNums(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	username := claims[identityKey].(string)

	online, abnormal, offline, deleted, err := dao.GetNodeNums(c, username)
	if err != nil {
		c.JSON(http.StatusOK, respErrorCode(errors.InternalServer, c))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"online":   online,
		"abnormal": abnormal,
		"offline":  offline,
		"deleted":  deleted,
	}))
}
