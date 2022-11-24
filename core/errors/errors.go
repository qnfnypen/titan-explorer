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

	Unknown = -1
)

var (
	ErrUnknown         = newError(Unknown, "unknown error")
	ErrNotFound        = newError(NotFound, "record not found")
	ErrInvalidParams   = newError(InvalidParams, "invalid params")
	ErrUserNotFound    = newError(UserNotFound, "user not found")
	ErrInvalidPassword = newError(InvalidPassword, "invalid password")
	ErrInternalServer  = newError(InternalServer, "internal server error")
	ErrDeviceExists    = newError(DeviceExists, "device already exists")
)

type ApiError struct {
	code int
	err  error
}

func (e ApiError) Code() int {
	return e.code
}

func (e ApiError) Error() string {
	return e.err.Error()
}

func (e ApiError) APIError() (int, string) {
	return e.code, e.err.Error()
}

func newError(code int, message string) ApiError {
	return ApiError{code, errors.New(message)}
}

func New(message string) error {
	return errors.New(message)
}
