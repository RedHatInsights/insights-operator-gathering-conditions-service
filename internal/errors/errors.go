package errors

import (
	"fmt"
)

type Error struct {
	orig error
	msg  string
	code ErrorCode
}

type ErrorCode uint

const (
	ErrorCodeUnknown ErrorCode = iota
	ErrorCodeNotFound
	ErrorCodeInvalidArgument
)

func (e *Error) Error() string {
	if e.orig != nil {
		return fmt.Sprintf("%s: %v", e.msg, e.orig)
	}

	return e.msg
}

func (e *Error) Unwrap() error {
	return e.orig
}

func (e *Error) Code() ErrorCode {
	return e.code
}

func WrapErrorf(orig error, code ErrorCode, format string, a ...interface{}) error {
	return &Error{
		orig: orig,
		code: code,
		msg:  fmt.Sprintf(format, a...),
	}
}

func NewErrorf(code ErrorCode, format string, a ...interface{}) error {
	return WrapErrorf(nil, code, format, a...)
}
