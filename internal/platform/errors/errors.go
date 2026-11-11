// Package errors 定义不依赖 Gin、GORM 和具体业务的错误分类。
package errors

import "errors"

type Class string

const (
	ClassInvalid      Class = "invalid"
	ClassUnauthorized Class = "unauthorized"
	ClassForbidden    Class = "forbidden"
	ClassNotFound     Class = "not_found"
	ClassConflict     Class = "conflict"
	ClassTooLarge     Class = "too_large"
	ClassRateLimit    Class = "rate_limit"
	ClassInternal     Class = "internal"
	ClassUnavailable  Class = "unavailable"
)

// Error 是可公开消息与原始原因分离的应用错误。
type Error struct {
	Class  Class
	Code   int
	Public string
	Cause  error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause != nil {
		return e.Public + ": " + e.Cause.Error()
	}
	return e.Public
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func (e *Error) AppCode() int {
	if e == nil {
		return 500
	}
	if e.Code != 0 {
		return e.Code
	}
	switch e.Class {
	case ClassInvalid:
		return 400
	case ClassUnauthorized:
		return 401
	case ClassForbidden:
		return 403
	case ClassNotFound:
		return 404
	case ClassConflict:
		return 409
	case ClassTooLarge:
		return 413
	case ClassRateLimit:
		return 429
	case ClassUnavailable:
		return 503
	default:
		return 500
	}
}

func newError(class Class, code int, public string, cause error) *Error {
	return &Error{Class: class, Code: code, Public: public, Cause: cause}
}

func Invalid(public string) *Error {
	return newError(ClassInvalid, 400, public, nil)
}

func Unauthorized(public string) *Error {
	return newError(ClassUnauthorized, 401, public, nil)
}

func Forbidden(public string) *Error {
	return newError(ClassForbidden, 403, public, nil)
}

func NotFound(public string) *Error {
	return newError(ClassNotFound, 404, public, nil)
}

func Conflict(code int, public string) *Error {
	if code == 0 {
		code = 409
	}
	return newError(ClassConflict, code, public, nil)
}

func TooLarge(public string) *Error {
	return newError(ClassTooLarge, 413, public, nil)
}

func RateLimit(public string) *Error {
	return newError(ClassRateLimit, 429, public, nil)
}

func Unavailable(public string) *Error {
	return newError(ClassUnavailable, 503, public, nil)
}

func Internal(cause error) *Error {
	return newError(ClassInternal, 500, "服务内部错误", cause)
}

func MethodNotAllowed() *Error {
	return newError(ClassInvalid, 405, "方法不允许", nil)
}

func WithCode(class Class, code int, public string) *Error {
	return newError(class, code, public, nil)
}

func As(err error) (*Error, bool) {
	var target *Error
	if errors.As(err, &target) {
		return target, true
	}
	return nil, false
}

func PublicMessage(err error) string {
	if err == nil {
		return ""
	}
	if appErr, ok := As(err); ok {
		return appErr.Public
	}
	return "服务内部错误"
}
