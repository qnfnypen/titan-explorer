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

	err = dao.CreateDeviceInfo(c.Request.Context(), deviceInfo)
	if err != nil {
		c.JSON(http.StatusBadRequest, respError(errors.ErrInternalServer))
		return
	}
	//
	go GDevice.GetDeviceIDs()

	c.JSON(http.StatusOK, respJSON(nil))
}

func DeviceBidingHandler(c *gin.Context) {
	deviceInfo := &model.DeviceInfo{}
	deviceInfo.DeviceID = c.Query("device_id")
	deviceInfo.UserID = c.Query("userId")

	err := dao.UpsertUserDevice(c.Request.Context(), deviceInfo)
	if err != nil {
		c.JSON(http.StatusBadRequest, respError(errors.ErrInternalServer))
		return
	}

	GDevice.GetDeviceIDs()
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
