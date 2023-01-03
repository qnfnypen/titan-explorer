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
	AmountLimitExceeded

	Unknown = -1
)

var (
	ErrUnknown             = newError(Unknown, "unknown error")
	ErrNotFound            = newError(NotFound, "record not found")
	ErrInvalidParams       = newError(InvalidParams, "invalid params")
	ErrUserNotFound        = newError(UserNotFound, "user not found")
	ErrInvalidPassword     = newError(InvalidPassword, "invalid password")
	ErrInternalServer      = newError(InternalServer, "internal server error")
	ErrDeviceExists        = newError(DeviceExists, "device already exists")
	ErrAmountLimitExceeded = newError(AmountLimitExceeded, "request amount limit exceeded")
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
