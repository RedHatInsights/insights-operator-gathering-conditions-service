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
	"net/http"
	"testing"

	"github.com/RedHatInsights/insights-operator-gathering-conditions-service/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestRepositoryRules(t *testing.T) {
	type testCase struct {
		name                 string
		mockConditionalRules []byte
		expectedAnError      bool
		expectedRules        service.Rules
	}

	testCases := []testCase{
		{
			name:                 "unparsable rule",
			mockConditionalRules: []byte("not a JSON"),
			expectedAnError:      true,
		},
		{
			name:                 "no data returned by storage",
			mockConditionalRules: nil,
			expectedAnError:      true,
		},
		{
			name:                 "valid rules returned by storage",
			mockConditionalRules: []byte(validStableRulesJSON),
			expectedAnError:      false,
			expectedRules:        validStableRules,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := mockStorage{
				conditionalRules: tc.mockConditionalRules,
			}
			r := service.NewRepository(&m)
			rules, err := r.Rules(&http.Request{})
			if tc.expectedAnError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, &tc.expectedRules, rules)
			}
		})
	}
}

func TestRepositoryRemoteConfiguration(t *testing.T) {
	const anyVer = "1.2.3"

	tests := []struct {
		name                 string
		mockRemoteConfig     []byte
		expectedAnError      bool
		expectedRemoteConfig service.RemoteConfiguration
	}{
		{
			name:             "unparsable rule",
			mockRemoteConfig: []byte("not a JSON"),
			expectedAnError:  true,
		},
		{
			name:             "no data returned by storage",
			mockRemoteConfig: nil,
			expectedAnError:  true,
		},
		{
			name:                 "valid remote configuration returned by storage",
			mockRemoteConfig:     []byte(validStableRemoteConfigurationJSON),
			expectedAnError:      false,
			expectedRemoteConfig: validStableRemoteConfiguration,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := mockStorage{
				remoteConfig: tt.mockRemoteConfig,
			}
			r := service.NewRepository(&m)
			remoteConfig, err := r.RemoteConfiguration(&http.Request{}, anyVer)
			if tt.expectedAnError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, &tt.expectedRemoteConfig, remoteConfig)
			}
		})
	}
}
