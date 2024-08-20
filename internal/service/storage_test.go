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
	"encoding/json"
	"testing"

	"github.com/RedHatInsights/insights-operator-gathering-conditions-service/internal/service"
	"github.com/stretchr/testify/assert"
)

const (
	validRulesFile         = "rules.json"
	rulesFolder            = "testdata"
	v2Folder               = "testdata/v2"
	clusterMappingFilepath = "testdata/empty.json"
)

func TestReadConditionalRules(t *testing.T) {
	type testCase struct {
		name          string
		rulesFile     string
		expectedRules service.Rules
	}

	testCases := []testCase{
		{
			name:          "file exists and is empty",
			rulesFile:     "empty.json",
			expectedRules: service.Rules{},
		},
		{
			name:          "file doesn't exit",
			rulesFile:     "not-found.json",
			expectedRules: service.Rules{},
		},
		{
			name:          "file exists and is valid",
			rulesFile:     validRulesFile,
			expectedRules: validRules,
		},
		{
			name:          "reading from 'directory' instead of file",
			rulesFile:     "",
			expectedRules: service.Rules{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storage, err := service.NewStorage(
				service.StorageConfig{
					RulesPath: rulesFolder,
				})
			assert.NoError(t, err)
			checkConditionalRules(t, storage, tc.rulesFile, tc.expectedRules)
		})
	}

	t.Run("use cache with a previous read file", func(t *testing.T) {
		// run Find anothertime to test the cache function
		storage, err := service.NewStorage(
			service.StorageConfig{
				RulesPath: rulesFolder,
			})
		assert.NoError(t, err)
		for i := 0; i < 2; i++ {
			checkConditionalRules(t, storage, validRulesFile, validRules)
		}
	})
}

func checkConditionalRules(t *testing.T, storage *service.Storage, rulesFile string, expectedRules service.Rules) {
	var rules service.Rules
	data := storage.ReadConditionalRules(rulesFile)
	if len(data) == 0 {
		rules = service.Rules{}
	} else {
		err := json.Unmarshal(data, &rules)
		assert.NoError(t, err)
	}
	assert.Equal(t, expectedRules, rules)
}

func checkRemoteConfig(t *testing.T, storage *service.Storage, remoteConfigFile string, expectedRemoteConfig service.RemoteConfiguration) {
	var remoteConfig service.RemoteConfiguration
	data := storage.ReadRemoteConfig(remoteConfigFile)
	if len(data) == 0 {
		remoteConfig = service.RemoteConfiguration{}
	} else {
		err := json.Unmarshal(data, &remoteConfig)
		assert.NoError(t, err)
	}
	assert.Equal(t, expectedRemoteConfig, remoteConfig)
}

func TestReadRemoteConfiguration(t *testing.T) {
	tests := []struct {
		name                 string
		remoteConfigFile     string
		expectedRemoteConfig service.RemoteConfiguration
	}{
		{
			name:                 "file exists and is empty",
			remoteConfigFile:     "empty.json",
			expectedRemoteConfig: service.RemoteConfiguration{},
		},
		{
			name:                 "file doesn't exit",
			remoteConfigFile:     "not-found.json",
			expectedRemoteConfig: service.RemoteConfiguration{},
		},
		{
			name:                 "file exists and is valid",
			remoteConfigFile:     validRulesFile,
			expectedRemoteConfig: validRemoteConfiguration,
		},
		{
			name:                 "reading from 'directory' instead of file",
			remoteConfigFile:     "",
			expectedRemoteConfig: service.RemoteConfiguration{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, err := service.NewStorage(
				service.StorageConfig{
					RemoteConfigurationPath: v2Folder,
				})
			assert.NoError(t, err)
			checkRemoteConfig(t, storage, tt.remoteConfigFile, tt.expectedRemoteConfig)
		})
	}
}
