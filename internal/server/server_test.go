/*
Copyright © 2022 Red Hat, Inc.

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
	"testing"
	"time"

	"github.com/RedHatInsights/insights-operator-gathering-conditions-service/internal/server"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

type testCase struct {
	name          string
	config        server.Config
	authConfig    server.AuthConfig
	expectAnError bool
}

func TestServer(t *testing.T) {
	testCases := []testCase{
		{
			name: "without CORS",
			config: server.Config{
				Address: "localhost:1234",
			},
			authConfig: server.AuthConfig{
				Enabled: false,
				Type:    "",
			},
		},
		{
			name: "with CORS",
			config: server.Config{
				Address:    "localhost:1234",
				EnableCORS: true,
			},
			authConfig: server.AuthConfig{
				Enabled: false,
				Type:    "",
			},
		},
		{
			name: "with TLS",
			config: server.Config{
				Address:    "localhost:1234",
				UseHTTPS:   true,
				CertFolder: "testdata/",
			},
			authConfig: server.AuthConfig{
				Enabled: false,
				Type:    "",
			},
		},
		{
			name: "with TLS but returning an error",
			config: server.Config{
				Address:    "localhost:1234",
				UseHTTPS:   true,
				CertFolder: "not-a-folder/",
			},
			authConfig: server.AuthConfig{
				Enabled: false,
				Type:    "",
			},
			expectAnError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testServer := server.New(tc.config, tc.authConfig, mux.NewRouter())
			go func() {
				err := testServer.Start()
				if tc.expectAnError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			}()
			time.Sleep(100 * time.Millisecond)
			err := testServer.Stop(context.TODO())
			assert.NoError(t, err)
		})
	}
}
