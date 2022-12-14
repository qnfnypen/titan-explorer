package api

import (
	"github.com/gin-gonic/gin"
	err "github.com/gnasnik/titan-explorer/core/errors"
	"github.com/pkg/errors"
)

type JsonObject map[string]interface{}

func respJSON(v interface{}) gin.H {
	return gin.H{
		"code": 0,
		"data": v,
	}
}

func respError(e error) gin.H {
	var genericError err.GenericError
	if !errors.As(e, &genericError) {
		genericError = err.ErrUnknown
	}

	return gin.H{
		"code": -1,
		"err":  genericError.Code,
		"msg":  genericError.Error(),
	}
}
