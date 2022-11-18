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
	var apiError err.ApiError
	if !errors.As(e, &apiError) {
		apiError = err.ErrUnknown
	}

	return gin.H{
		"code": -1,
		"err":  apiError.Code(),
		"msg":  apiError.Error(),
	}
}
