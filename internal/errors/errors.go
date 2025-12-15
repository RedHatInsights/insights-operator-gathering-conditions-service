/*
Copyright © 2021, 2022 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package errors provide custom error related structs and utilities
//
//nolint:revive // Package name intentionally matches stdlib for consistency with error handling patterns
package errors

import (
	"fmt"
)

// Error data structure contains the original error object + assigned error code
type Error struct {
	orig error
	msg  string
	code ErrorCode
}

// ErrorCode is enumeration type to specify numeric error code
type ErrorCode uint

const (
	// ErrorCodeUnknown represents numeric error code for unknown error
	ErrorCodeUnknown ErrorCode = iota

	// ErrorCodeNotFound represents numeric error code for error that
	// occurs when the rule data are not found
	ErrorCodeNotFound

	// ErrorCodeInvalidArgument represents numeric error code for error
	// that occurs when invalid argument is provided in request
	ErrorCodeInvalidArgument
)

func (e *Error) Error() string {
	if e.orig != nil {
		return fmt.Sprintf("%s: %v", e.msg, e.orig)
	}

	return e.msg
}

// Unwrap will return the original error
func (e *Error) Unwrap() error {
	return e.orig
}

// Code method returns numeric error code
func (e *Error) Code() ErrorCode {
	return e.code
}

// WrapErrorf function constructs Error data structure with original error
// object, numeric error code, and message.
func WrapErrorf(orig error, code ErrorCode, format string, a ...interface{}) error {
	return &Error{
		orig: orig,
		code: code,
		msg:  fmt.Sprintf(format, a...),
	}
}

// NewErrorf function constructs Error data structure with numeric error code,
// and message.
func NewErrorf(code ErrorCode, format string, a ...interface{}) error {
	return WrapErrorf(nil, code, format, a...)
}

// ResponseDataError is used as the error message when the responses functions return an error
const ResponseDataError = "Unexpected error during response data encoding"

// RouterMissingParamError missing parameter in request
type RouterMissingParamError struct {
	ParamName string
}

// Error method transforms error structure to a string representation
func (e *RouterMissingParamError) Error() string {
	return fmt.Sprintf("Missing required param from request: %v", e.ParamName)
}

// RouterParsingError parsing error, for example string when we expected integer
type RouterParsingError struct {
	ParamName  string
	ParamValue interface{}
	ErrString  string
}

// Error method transforms error structure to a string representation
func (e *RouterParsingError) Error() string {
	return fmt.Sprintf(
		"Error during parsing param '%v' with value '%v'. Error: '%v'",
		e.ParamName, e.ParamValue, e.ErrString,
	)
}

// AuthenticationError happens during auth problems, for example malformed token
type AuthenticationError struct {
	ErrString string
}

// Error method transforms error structure to a string representation
func (e *AuthenticationError) Error() string {
	return e.ErrString
}

// UnauthorizedError means server can't authorize you, for example the token is missing or malformed
type UnauthorizedError struct {
	ErrString string
}

// Error method transforms error structure to a string representation
func (e *UnauthorizedError) Error() string {
	return e.ErrString
}

// ForbiddenError means user does not have permission to do a particular
// action, for example the account belongs to a different organization
type ForbiddenError struct {
	ErrString string
}

// Error method transforms error structure to a string representation
func (e *ForbiddenError) Error() string {
	return e.ErrString
}

// NoBodyError error meaning that client didn't provide body when it's required
type NoBodyError struct{}

func (*NoBodyError) Error() string {
	return "client didn't provide request body"
}

// NotFoundError meaning that the requested resource wasn't found
type NotFoundError struct {
	ErrString string
}

func (e *NotFoundError) Error() string {
	return e.ErrString
}

// ValidationError validation error, for example when string is longer then expected
type ValidationError struct {
	ParamName  string
	ParamValue interface{}
	ErrString  string
}

// Error method transforms error structure to a string representation
func (e *ValidationError) Error() string {
	return fmt.Sprintf(
		"Error during validating param '%v' with value '%v'. Error: '%v'",
		e.ParamName, e.ParamValue, e.ErrString,
	)
}
