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

package service_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RedHatInsights/insights-operator-gathering-conditions-service/internal/service"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestService(t *testing.T) {
	type testCase struct {
		name            string
		mockData        []byte
		expectedAnError bool
	}

	testCases := []testCase{
		{
			name:            "valid rule stored",
			mockData:        []byte(validRulesJSON),
			expectedAnError: false,
		},
		{
			name:            "invalid rule stored",
			mockData:        []byte("not a rule"),
			expectedAnError: true,
		},
	}

	// CCXDEV-7919: for checking with v1 version and without version
	versions := []string{
		service.V1Prefix,
		"",
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, version := range versions {
				store := mockStorage{
					mockData: tc.mockData,
				}
				repo := service.NewRepository(&store)
				svc := service.New(repo)

				// Create the request:
				req, err := http.NewRequest("GET", fmt.Sprintf("%s%s/gathering_rules", service.APIPrefix, version), http.NoBody)
				if err != nil {
					t.Fatal(err)
				}

				rr := httptest.NewRecorder() // Used to record the response.
				handler := service.NewHandler(svc)

				router := mux.NewRouter()

				handler.Register(router)

				router.ServeHTTP(rr, req)

				if tc.expectedAnError {
					assert.Equal(t, http.StatusInternalServerError, rr.Code)
					assert.Contains(t, rr.Body.String(), "error")
				} else {
					assert.Equal(t, http.StatusOK, rr.Code)
					assert.Contains(
						t,
						rr.Body.String(),
						`"version":"0.0.1","rules":[{"conditions":["condition 1","condition 2"],"gathering_functions":"the gathering functions"}]`)
				}
			}
		})
	}
}
