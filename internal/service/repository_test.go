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
	"testing"

	"github.com/redhatinsights/insights-operator-gathering-conditions-service/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestRepository(t *testing.T) {
	type testCase struct {
		name            string
		mockData        []byte
		expectedAnError bool
		expectedRules   service.Rules
	}

	testCases := []testCase{
		{
			name:            "unparsable rule",
			mockData:        []byte("not a JSON"),
			expectedAnError: true,
		},
		{
			name:            "no data returned by storage",
			mockData:        nil,
			expectedAnError: true,
		},
		{
			name:            "valid rules returned by storage",
			mockData:        []byte(validRulesJSON),
			expectedAnError: false,
			expectedRules:   validRules,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := mockStorage{
				mockData: tc.mockData,
			}
			r := service.NewRepository(&m)
			rules, err := r.Rules()
			if tc.expectedAnError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, &tc.expectedRules, rules)
			}
		})
	}
}
