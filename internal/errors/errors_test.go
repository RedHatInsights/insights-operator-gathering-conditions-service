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

package errors_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/RedHatInsights/insights-operator-gathering-conditions-service/internal/errors"
)

// TestRouterMissingParamError checks the method Error() for data structure
// RouterMissingParamError
func TestRouterMissingParamError(t *testing.T) {
	err := errors.RouterMissingParamError{
		ParamName: "paramName",
	}

	const expected = "Missing required param from request: paramName"
	assert.Equal(t, err.Error(), expected)
}

// TestRouterParsingError checks the method Error() for data structure
// RouterParsingError
func TestRouterParsingError(t *testing.T) {
	err := errors.RouterParsingError{
		ParamName:  "paramName",
		ParamValue: "paramValue",
		ErrString:  "errorMessage"}

	const expected = "Error during parsing param 'paramName' with value 'paramValue'. Error: 'errorMessage'"
	assert.Equal(t, err.Error(), expected)
}

// TestAuthenticationError checks the method Error() for data structure
// AuthenticationError
func TestAuthenticationError(t *testing.T) {
	err := errors.AuthenticationError{
		ErrString: "errorMessage"}

	const expected = "errorMessage"
	assert.Equal(t, err.Error(), expected)
}

// TestUnauthorizedError checks the method Error() for data structure
// UnauthorizedError
func TestUnauthorizedError(t *testing.T) {
	err := errors.UnauthorizedError{
		ErrString: "errorMessage"}

	const expected = "errorMessage"
	assert.Equal(t, err.Error(), expected)
}

// TestForbiddenError checks the method Error() for data structure
// ForbiddenError
func TestForbiddenError(t *testing.T) {
	err := errors.ForbiddenError{
		ErrString: "errorMessage"}

	const expected = "errorMessage"
	assert.Equal(t, err.Error(), expected)
}

// TestNoBodyError checks the method Error() for data structure
// NoBodyError
func TestNoBodyError(t *testing.T) {
	err := errors.NoBodyError{}

	const expected = "client didn't provide request body"
	assert.Equal(t, err.Error(), expected)
}

// TestRouterParsingError checks the method Error() for data structure
// ValidationError
func TestValidationError(t *testing.T) {
	err := errors.ValidationError{
		ParamName:  "paramName",
		ParamValue: "paramValue",
		ErrString:  "errorMessage"}

	const expected = "Error during validating param 'paramName' with value 'paramValue'. Error: 'errorMessage'"
	assert.Equal(t, err.Error(), expected)
}

// TestError checks the method Error() for data structure Error
func TestError(t *testing.T) {
	err := errors.Error{}
	assert.Equal(t, err.Error(), "")
}

// TestErrorCode checks the method Code() for data structure Error
func TestErrorCode(t *testing.T) {
	err := errors.Error{}
	assert.Equal(t, err.Code(), errors.ErrorCode(0))
}

// TestErrorUnwrap checks the method Unwrap() for data structure Error
func TestErrorUnwrap(t *testing.T) {
	err := errors.Error{}
	assert.Nil(t, err.Unwrap())
}

// TestNewErrorf checks the constructor NewErrorf
func TestNewErrorf(t *testing.T) {
	err := errors.NewErrorf(10, "%s = %d", "1", 1)
	assert.Equal(t, err.Error(), "1 = 1")
}
