/*
Copyright © 2022, 2024 Red Hat, Inc.

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

func TestWrongMethods(t *testing.T) {
	endpoints := []string{
		"/openapi.json",
		"/gathering_rules",
		"/v1/gathering_rules",
		"/v1/openapi.json",
		"/v2/openapi.json",
		"/v2/1.2.3/gathering_rules",
	}

	methods := []string{"POST", "PUT", "DELETE"}

	store := mockStorage{
		conditionalRules: []byte(validStableRemoteConfigurationJSON),
	}
	repo := service.NewRepository(&store)
	svc := service.New(repo)
	handler := service.NewHandler(svc)
	router := mux.NewRouter()
	handler.Register(router)

	for _, endpoint := range endpoints {
		for _, method := range methods {
			t.Run(fmt.Sprintf("%s:%s", endpoint, method), func(t *testing.T) {
				// Create the request:
				req, err := http.NewRequest(
					method,
					fmt.Sprintf("%s%s", service.APIPrefix, endpoint),
					http.NoBody)
				if err != nil {
					t.Fatal(err)
				}

				rr := httptest.NewRecorder() // Used to record the response.
				router.ServeHTTP(rr, req)
				assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
			})
		}
	}
}
