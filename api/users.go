package api

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/errors"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/google/uuid"
	"golang.org/x/xerrors"
	"net/http"
	"strings"
)

const (
	NodeUnknown = iota
	NodeEdge
	NodeCandidate
)

func DeviceCreateHandler(c *gin.Context) {
	nodeType := c.Query("node_type")
	nodeTypeInt := Str2Int(nodeType)
	deviceID, err := newDeviceID(nodeTypeInt)
	if err != nil {
		return
	}
	res := make(map[string]interface{})
	secret := newSecret(deviceID)
	deviceInfo := &model.DeviceInfo{}
	deviceInfo.DeviceID = deviceID
	deviceInfo.Secret = secret
	deviceInfo.NodeType = int32(nodeTypeInt)
	res["device_id"] = deviceID
	res["secret"] = secret

	err = dao.AddDeviceInfo(c.Request.Context(), deviceInfo)
	if err != nil {
		c.JSON(http.StatusBadRequest, respError(errors.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, respJSON(nil))
}

func DeviceBindingHandler(c *gin.Context) {
	deviceInfo := &model.DeviceInfo{}
	deviceInfo.DeviceID = c.Query("device_id")
	deviceInfo.UserID = c.Query("userId")

	old, err := dao.GetDeviceInfoByID(c.Request.Context(), deviceInfo.DeviceID)
	if err != nil {
		log.Errorf("get user device: %v", err)
		c.JSON(http.StatusBadRequest, respError(errors.ErrInternalServer))
		return
	}

	if old != nil && old.UserID != "" {
		c.JSON(http.StatusBadRequest, respError(errors.ErrDeviceExists))
		return
	}

	err = dao.UpdateUserDeviceInfo(c.Request.Context(), deviceInfo)
	if err != nil {
		log.Errorf("update user device: %v", err)
		c.JSON(http.StatusBadRequest, respError(errors.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, respJSON(nil))
}

func newDeviceID(nodeType int) (string, error) {
	u2, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}

	s := strings.Replace(u2.String(), "-", "", -1)
	switch nodeType {
	case NodeEdge:
		s = fmt.Sprintf("e_%s", s)
		return s, nil
	case NodeCandidate:
		s = fmt.Sprintf("c_%s", s)
		return s, nil
	}

	return "", xerrors.Errorf("nodetype err:%v", nodeType)
}

func newSecret(input string) string {
	c := sha1.New()
	c.Write([]byte(input))
	bytes := c.Sum(nil)
	return hex.EncodeToString(bytes)
}
