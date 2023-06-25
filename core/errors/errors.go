package errors

import (
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
	KeyLimit

	Unknown     = -1
	GenericCode = 1
)

var (
	ErrUnknown             = newError(Unknown, "unknown error")
	ErrNotFound            = newError(NotFound, "not found")
	ErrInvalidParams       = newError(InvalidParams, "invalid params")
	ErrUserNotFound        = newError(UserNotFound, "user not found")
	ErrInvalidPassword     = newError(InvalidPassword, "invalid password")
	ErrInternalServer      = newError(InternalServer, "internal server error")
	ErrDeviceExists        = newError(DeviceExists, "device already exists")
	ErrDeviceNotExists     = newError(DeviceNotExists, "device not exists")
	ErrAmountLimitExceeded = newError(AmountLimitExceeded, "request amount limit exceeded")
	ErrUnbindingNotAllowed = newError(UnbindingNotAllowed, "unbinding not allowed")
	ErrPassWord            = newError(PassWordNotAllowed, "password not allowed")
	ErrNameExists          = newError(NameExists, "the name Exists")
	ErrNameNotExists       = newError(NameNotExists, "the name not exists")
	ErrVerifyCode          = newError(VerifyCodeErr, "verify code err ")
	ErrVerifyCodeExpired   = newError(VerifyCodeExpired, "verify code expired")
	ErrUploadExceed        = newError(UploadExceed, "asset in the pulling exceeds the limit 10")
	ErrFileExists          = newError(FileExists, "file already exists no need to upload")
	ErrKeyLimit            = newError(KeyLimit, "key limit only one")
)

type GenericError struct {
	Code int
	Err  error
}

func (e GenericError) Error() string {
	return e.Err.Error()
}

func newError(code int, message string) GenericError {
	return GenericError{Code: code, Err: errors.New(message)}
}

func NewError(msg string) GenericError {
	return newError(GenericCode, msg)
}
