package errors

import (
	"strings"

	"github.com/Filecoin-Titan/titan/api/terrors"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
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
	InvalidAPPKey
	NoBearerToken
	GetVCFrequently
	InvalidSignature
	WalletBound
	InvalidReferralCode

	Unknown     = -1
	GenericCode = 1
)

// ErrMap some errors from titan
var ErrMap = map[int]string{
	Unknown:                                  "unknown error:未知错误",
	NotFound:                                 "not found:信息未找到",
	InvalidParams:                            "invalid params:参数有误",
	UserNotFound:                             "user not found:用户不存在",
	InvalidPassword:                          "invalid password:密码错误",
	InternalServer:                           "internal server error:服务器错误",
	DeviceExists:                             "device already exists:设备已存在",
	DeviceNotExists:                          "device not exists:设备不存在",
	AmountLimitExceeded:                      "request amount limit exceeded:请求数量限制",
	UnbindingNotAllowed:                      "unbinding not allowed:暂不能解绑",
	PassWordNotAllowed:                       "password not allowed:密码错误",
	NameExists:                               "the name Exists:该名称已存在",
	NameNotExists:                            "the name not exists:该名称不存在",
	VerifyCodeErr:                            "verify code err:验证码错误",
	VerifyCodeExpired:                        "verify code expired:验证码过期",
	UploadExceed:                             "asset in the pulling exceeds the limit 10:上传数量限制10条",
	FileExists:                               "file already exists no need to upload:该文件已经存在",
	int(terrors.NotFound):                    "not found:信息未找到",
	int(terrors.DatabaseErr):                 "database error:数据库错误",
	int(terrors.ParametersAreWrong):          "The parameters are wrong:参数有误",
	int(terrors.CidToHashFiled):              "cid to hash failed:cid解析错误",
	int(terrors.UserStorageSizeNotEnough):    "Insufficient user storage space:用户存储空间不足",
	int(terrors.UserNotFound):                "Unable to be found by the user:用户不存在",
	int(terrors.NoDuplicateUploads):          "No duplicate uploads:文件重复上传",
	int(terrors.BusyServer):                  "Busy server:服务器繁忙",
	int(terrors.NotFoundNode):                "Can't find the node:节点未找到",
	int(terrors.RequestNodeErr):              "Request node error:请求节点错误",
	int(terrors.MarshalErr):                  "Marshal error:数据解析错误",
	int(terrors.OutOfMaxAPIKeyLimit):         "api key exceeds quantity limit:APIKey超过数量限制",
	int(terrors.APPKeyAlreadyExist):          "same name APIKey already exist:相同名字的APIKey已经存在",
	int(terrors.APPKeyNotFound):              "same name APIKey not found:找不到相同名字的APIKey",
	int(terrors.APIKeyACLError):              "API Key permission list error:APIKey 权限列表错误",
	int(terrors.GroupNotEmptyCannotBeDelete): "The group is not empty and cannot be deleted:不能删除非空分组",
	int(terrors.GroupNotExist):               "The group is not exist:分组不存在",
	int(terrors.GroupLimit):                  "The group is limit:分组限制",
	InvalidAPPKey:                            "invalid key:无效的key",
	NoBearerToken:                            "Could not find bearer token in Authorization header: 找不到 Bearer token",
	GetVCFrequently:                          "Please do not obtain verification codes frequently. Please wait for some time and try again.:请勿频繁获取验证码。请等待一段时间后再试。",
	InvalidSignature:                         "invalid signature: 无效的签名",
	WalletBound:                              "wallet bound: 已绑定钱包",
	InvalidReferralCode:                      "invalid referral code: 无效的邀请码",
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
