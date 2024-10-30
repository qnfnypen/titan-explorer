package api

import (
	"strings"

	"github.com/gin-gonic/gin"
	err "github.com/gnasnik/titan-explorer/core/errors"
	"github.com/gnasnik/titan-explorer/core/generated/model"
)

type JsonObject map[string]interface{}

func respJSON(v interface{}) gin.H {
	return gin.H{
		"code": 0,
		"data": v,
	}
}

func respErrorCode(code int, c *gin.Context, extra ...string) gin.H {
	lang := c.GetHeader("Lang")

	var msg string

	messages := strings.Split(err.ErrMap[code], ":")
	if len(messages) == 0 {
		msg = err.ErrMap[code]
	} else {
		if lang == model.LanguageCN && len(messages) > 1 {
			msg = messages[1]
		} else {
			msg = messages[0]
		}
	}

	return gin.H{
		"code":  -1,
		"err":   code,
		"msg":   msg,
		"extra": extra,
	}
}

func respError(code int, err error) gin.H {
	return gin.H{
		"code": -1,
		"err":  code,
		"msg":  err.Error(),
	}
}
