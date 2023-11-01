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

package server_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/RedHatInsights/insights-operator-gathering-conditions-service/internal/server"
	"github.com/gorilla/mux"
)

type authTestCase struct {
	name             string
	identity         string
	expectedError    string
	expectedIdentity server.Identity
	expectedUserID   server.UserID
}

var (
	validIdentity = server.Identity{
		AccountNumber: server.UserID("a user"),
		Internal:      server.Internal{OrgID: 1},
	}
)

// TestGetAuthToken function checks the Server.GetAuthToken method
func TestGetAuthToken(t *testing.T) {
	testCases := []authTestCase{
		{
			name:             "valid token",
			identity:         "valid",
			expectedError:    "",
			expectedIdentity: validIdentity,
		},
		{
			name:          "no token",
			identity:      "empty",
			expectedError: "token is not provided",
		},
		{
			name:          "invalid token",
			identity:      "bad",
			expectedError: "contextKeyUser has wrong type",
		},
	}

	testServer := server.Server{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := getRequest(t, tc.identity)

			identity, err := testServer.GetAuthToken(req)
			if tc.expectedError == "" {
				require.NoError(t, err)
				assert.Equal(t, &tc.expectedIdentity, identity)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			}
		})
	}
}

// TestGetCurrentUserID function check the method Server.GetCurrentUserID
func TestGetCurrentUserID(t *testing.T) {
	testCases := []authTestCase{
		{
			name:           "valid token",
			identity:       "valid",
			expectedError:  "",
			expectedUserID: validIdentity.AccountNumber,
		},
		{
			name:          "no token",
			identity:      "empty",
			expectedError: "user id is not provided",
		},
		{
			name:          "invalid token",
			identity:      "bad",
			expectedError: "contextKeyUser has wrong type",
		},
	}

	testServer := server.Server{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := getRequest(t, tc.identity)

			userID, err := testServer.GetCurrentUserID(req)
			if tc.expectedError == "" {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedUserID, userID)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			}
		})
	}
}

// Helper function to retrieve request
func getRequest(t *testing.T, identity string) *http.Request {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, "an url", http.NoBody)
	assert.NoError(t, err)

	if identity == "valid" {
		ctx := context.WithValue(req.Context(), server.ContextKeyUser, validIdentity)
		req = req.WithContext(ctx)
	}

	if identity == "bad" {
		ctx := context.WithValue(req.Context(), server.ContextKeyUser, "not an identity")
		req = req.WithContext(ctx)
	}

	return req
}

// TestStartServerWithAuth checks if server can be started in auth. mode
// enabled
func TestStartServerWithAuth(t *testing.T) {
	testServer := server.New(serverConfig, configAuth1, mux.NewRouter())
	go func() {
		err := testServer.Start()
		assert.NoError(t, err)
	}()
	time.Sleep(100 * time.Millisecond)
	err := testServer.Stop(context.TODO())
	assert.NoError(t, err)
}

// REST API based tests

const testedEndpoint = "/foobar"
const testedURL = "http://localhost:1234/foobar"

// server configuration used by tests
var serverConfig = server.Config{
	Address:    "localhost:1234",
	UseHTTPS:   false,
	EnableCORS: false,
	CertFolder: "testdata",
}

// auth. configuration used by tests for checking JWT token handling
var configAuth1 = server.AuthConfig{
	Enabled: true,
	Type:    "jwt",
}

// auth. configuration used by tests for checking x-rh token handling
var configAuth2 = server.AuthConfig{
	Enabled: true,
	Type:    "xrh",
}

// dummy HTTP request handler
func dummyHandler(_ http.ResponseWriter, _ *http.Request) {
}

// start new HTTP server, perform request, check response, and stop HTTP server
func testServerWithRequest(t *testing.T, configAuth server.AuthConfig, token string, expectedStatusCode int) {
	// construct HTTP server with one dummy handler
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc(testedEndpoint, dummyHandler)
	testServer := server.New(serverConfig, configAuth, router)

	// start HTTP server
	go func() {
		err := testServer.Start()
		assert.NoError(t, err)
	}()
	time.Sleep(100 * time.Millisecond)

	// perform request with auth. token
	request, err := http.NewRequest("GET", testedURL, http.NoBody)
	if err != nil {
		t.Fatal(err)
	}

	if token != "" {
		if configAuth.Type == "jwt" {
			request.Header.Set("Authorization", token)
		} else {
			request.Header.Set("x-rh-identity", token)
		}
	}
	client := &http.Client{}
	res, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}

	// check the response
	if res.StatusCode != expectedStatusCode {
		t.Errorf("Expected status code %v, got %v", expectedStatusCode, res.StatusCode)
	}

	// stop HTTP server
	err = testServer.Stop(context.TODO())
	assert.NoError(t, err)
}

type authServerTestCase struct {
	name           string
	authConfig     server.AuthConfig
	token          string
	expectedStatus int
}

// TestAuth checks how HTTP server handles auth. tokens
func TestAuth(t *testing.T) {
	testCases := []authServerTestCase{
		{
			name:           "Missing auth. token, JWT variant",
			authConfig:     configAuth1,
			token:          "",
			expectedStatus: 401,
		},
		{
			name:           "Proper JWT token",
			authConfig:     configAuth1,
			token:          "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2NvdW50X251bWJlciI6IjUyMTM0NzYiLCJvcmdfaWQiOiIxMjM0In0.Y9nNaZXbMEO6nz2EHNaCvHxPM0IaeT7GGR-T8u8h_nr_2b5dYsCQiZGzzkBupRJruHy9K6acgJ08JN2Q28eOAEVk_ZD2EqO43rSOS6oe8uZmVo-nCecdqovHa9PqW8RcZMMxVfGXednw82kKI8j1aT_nbJ1j9JZt3hnHM4wtqydelMij7zKyZLHTWFeZbDDCuEIkeWA6AdIBCMdywdFTSTsccVcxT2rgv4mKpxY1Fn6Vu_Xo27noZW88QhPTHbzM38l9lknGrvJVggrzMTABqWEXNVHbph0lXjPWsP7pe6v5DalYEBN2r3a16A6s3jPfI86cRC6_oeXotlW6je0iKQ",
			expectedStatus: 200,
		},
		{
			name:           "Malformed JWT token",
			authConfig:     configAuth1,
			token:          "Bearer ",
			expectedStatus: 401,
		},
		{
			name:           "Malformed JSON in JWT token",
			authConfig:     configAuth1,
			token:          "Bearer bm90LWpzb24K.bm90LWpzb24K.bm90LWpzb24K",
			expectedStatus: 401,
		},
		{
			name:           "Missing auth. token, JWT variant",
			authConfig:     configAuth2,
			token:          "",
			expectedStatus: 401,
		},
		{
			name:           "Proper XRH token",
			authConfig:     configAuth2,
			token:          "eyJpZGVudGl0eSI6IHsiaW50ZXJuYWwiOiB7Im9yZ19pZCI6ICIxMjM0In19fQo=",
			expectedStatus: 200,
		},
		{
			name:           "Malformed XRH token",
			authConfig:     configAuth2,
			token:          "!",
			expectedStatus: 401,
		},
		{
			name:           "Malformed XRH token",
			authConfig:     configAuth2,
			token:          "123456qwerty",
			expectedStatus: 401,
		},
		{
			name:           "Stripped XRH token",
			authConfig:     configAuth2,
			token:          "aW52YWxpZCBqc29uCg==",
			expectedStatus: 401,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testServerWithRequest(t, tc.authConfig, tc.token, tc.expectedStatus)
		})
	}
}
