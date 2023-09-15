package errors

import (
	"github.com/Filecoin-Titan/titan/api/terrors"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"strings"
)

const (
	NotFound = iota + 1000
	InvalidParams
	UserNotFound
	InvalidPassword
	InternalServer
	DeviceExists
	DeviceNotExists
	AmountLimitExceeded
	UnbindingNotAllowed
	PassWordNotAllowed
	NameExists
	NameNotExists
	VerifyCodeErr
	VerifyCodeExpired
	UploadExceed
	FileExists
	KeyLimit
	InvalidAPPKey
	NoBearerToken

	Unknown     = -1
	GenericCode = 1
)

// ErrMap some errors from titan
var ErrMap = map[int]string{
	Unknown:                          "unknown error:未知错误",
	NotFound:                         "not found:信息未找到",
	InvalidParams:                    "invalid params:参数有误",
	UserNotFound:                     "user not found:用户不存在",
	InvalidPassword:                  "invalid password:密码错误",
	InternalServer:                   "internal server error:服务器错误",
	DeviceExists:                     "device already exists:设备已存在",
	DeviceNotExists:                  "device not exists:设备不存在",
	AmountLimitExceeded:              "request amount limit exceeded:请求数量限制",
	UnbindingNotAllowed:              "unbinding not allowed:暂不能解绑",
	PassWordNotAllowed:               "password not allowed:密码错误",
	NameExists:                       "the name Exists:该名称已存在",
	NameNotExists:                    "the name not exists:该名称不存在",
	VerifyCodeErr:                    "verify code err:验证码错误",
	VerifyCodeExpired:                "verify code expired:验证码过期",
	UploadExceed:                     "asset in the pulling exceeds the limit 10:上传数量限制10条",
	FileExists:                       "file already exists no need to upload:该文件已经存在",
	KeyLimit:                         "key limit only one:只允许一个key",
	terrors.NotFound:                 "not found:信息未找到",
	terrors.DatabaseErr:              "database error:数据库错误",
	terrors.ParametersAreWrong:       "The parameters are wrong:参数有误",
	terrors.CidToHashFiled:           "cid to hash failed:cid解析错误",
	terrors.UserStorageSizeNotEnough: "Insufficient user storage space:用户存储空间不足",
	terrors.UserNotFound:             "Unable to be found by the user:用户不存在",
	terrors.NoDuplicateUploads:       "No duplicate uploads:文件重复上传",
	terrors.BusyServer:               "Busy server:服务器繁忙",
	terrors.NotFoundNode:             "Can't find the node:节点未找到",
	terrors.RequestNodeErr:           "Request node error:请求节点错误",
	terrors.MarshalErr:               "Marshal error:数据解析错误",
	InvalidAPPKey:                    "invalid key:无效的key",
	NoBearerToken:                    "Could not find bearer token in Authorization header: 找不到 Bearer token",
}

type GenericError struct {
	Code int
	Err  error
}

func (e GenericError) Error() string {
	return e.Err.Error()
}

func NewErrorCode(Code int, c *gin.Context) GenericError {
	l := c.GetHeader("Lang")
	errSplit := strings.Split(ErrMap[Code], ":")
	var e string
	switch l {
	case "cn":
		e = errSplit[1]
	default:
		e = errSplit[0]
	}
	return GenericError{Code: Code, Err: errors.New(e)}

}
