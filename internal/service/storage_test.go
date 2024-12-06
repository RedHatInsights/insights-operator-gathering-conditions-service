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
	"encoding/json"
	"net/http"
	"testing"

	"github.com/RedHatInsights/insights-operator-gathering-conditions-service/internal/service"
	"github.com/stretchr/testify/assert"
)

const (
	validRulesFile     = "rules.json"
	rulesFolder        = "testdata/v1"
	v2Folder           = "testdata/v2"
	clusterMappingPath = "testdata/mapping"
	clusterMappingFile = "cluster-mapping.json"
)

type MockUnleashClient struct{}

func (c *MockUnleashClient) IsCanary(canaryArgument string) bool {
	return canaryArgument == canaryClusterID
}

func TestNewStorage(t *testing.T) {
	type testCase struct {
		name        string
		config      service.StorageConfig
		expectError bool
	}

	testCases := []testCase{
		{
			name: "config is valid",
			config: service.StorageConfig{
				RulesPath:               "testdata/v1",
				RemoteConfigurationPath: "testdata/v2",
				ClusterMappingPath:      "testdata/mapping",
				ClusterMappingFile:      "cluster-mapping.json",
			},
			expectError: false,
		},
		{
			name: "config is invalid: stable version of data does not exit",
			config: service.StorageConfig{
				RulesPath:               "testdata/v1",
				RemoteConfigurationPath: "testdata/v2_missing_stable",
				ClusterMappingPath:      "testdata/mapping",
				ClusterMappingFile:      "cluster-mapping.json",
			},
			expectError: true,
		},
		{
			name: "config is invalid: canary version of data does not exit",
			config: service.StorageConfig{
				RulesPath:               "testdata/v1",
				RemoteConfigurationPath: "testdata/v2_missing_canary",
				ClusterMappingPath:      "testdata/mapping",
				ClusterMappingFile:      "cluster-mapping.json",
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := service.NewStorage(tc.config, false, nil)
			if tc.expectError {
				assert.NotNil(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

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
			expectedRules: validStableRules,
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
					RulesPath:               rulesFolder,
					ClusterMappingPath:      clusterMappingPath,
					ClusterMappingFile:      clusterMappingFile,
					RemoteConfigurationPath: v2Folder,
				}, false, nil)
			assert.NoError(t, err)
			checkConditionalRules(t, storage, tc.rulesFile, tc.expectedRules, &http.Request{})
		})
	}

	t.Run("use cache with a previous read file", func(t *testing.T) {
		// run Find anothertime to test the cache function
		storage, err := service.NewStorage(
			service.StorageConfig{
				RulesPath:               rulesFolder,
				ClusterMappingPath:      clusterMappingPath,
				ClusterMappingFile:      clusterMappingFile,
				RemoteConfigurationPath: v2Folder,
			}, false, nil)
		assert.NoError(t, err)
		for i := 0; i < 2; i++ {
			checkConditionalRules(t, storage, validRulesFile, validStableRules, &http.Request{})
		}
	})
}

func TestReadRulesCanaryRollout(t *testing.T) {
	tests := []struct {
		name             string
		canaryArgument   string
		remoteConfigFile string
		expectedRules    service.Rules
	}{
		{
			name:             "cluster is served with stable version of rules",
			canaryArgument:   stableUserAgent,
			remoteConfigFile: validRulesFile,
			expectedRules:    validStableRules,
		},
		{
			name:             "cluster is served with canary version of rules",
			canaryArgument:   canaryUserAgent,
			remoteConfigFile: validRulesFile,
			expectedRules:    validCanaryRules,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, err := service.NewStorage(
				service.StorageConfig{
					RulesPath:               rulesFolder,
					ClusterMappingPath:      clusterMappingPath,
					ClusterMappingFile:      clusterMappingFile,
					RemoteConfigurationPath: v2Folder,
				}, true, &MockUnleashClient{})
			assert.NoError(t, err)
			req, err := http.NewRequest("GET", "http://example.com", nil)
			assert.NoError(t, err)
			req.Header.Add("User-Agent", tt.canaryArgument)
			checkConditionalRules(t, storage, tt.remoteConfigFile, tt.expectedRules, req)
		})
	}
}

func checkConditionalRules(t *testing.T, storage *service.Storage, rulesFile string, expectedRules service.Rules, r *http.Request) {
	var rules service.Rules
	data := storage.ReadConditionalRules(storage.IsCanary(r), rulesFile)
	if len(data) == 0 {
		rules = service.Rules{}
	} else {
		err := json.Unmarshal(data, &rules)
		assert.NoError(t, err)
	}
	assert.Equal(t, expectedRules, rules)
}

func checkRemoteConfig(t *testing.T, storage *service.Storage, remoteConfigFile string, expectedRemoteConfig service.RemoteConfiguration, r *http.Request) {
	var remoteConfig service.RemoteConfiguration
	data := storage.ReadRemoteConfig(storage.IsCanary(r), remoteConfigFile)
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
			expectedRemoteConfig: validStableRemoteConfiguration,
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
					ClusterMappingPath:      clusterMappingPath,
					ClusterMappingFile:      clusterMappingFile,
				}, false, nil)
			assert.NoError(t, err)
			checkRemoteConfig(t, storage, tt.remoteConfigFile, tt.expectedRemoteConfig, &http.Request{})
		})
	}
}

func TestReadRemoteConfigurationCanaryRollout(t *testing.T) {
	tests := []struct {
		name                 string
		canaryArgument       string
		remoteConfigFile     string
		expectedRemoteConfig service.RemoteConfiguration
	}{
		{
			name:                 "cluster is served with stable version of remote config",
			canaryArgument:       stableUserAgent,
			remoteConfigFile:     validRulesFile,
			expectedRemoteConfig: validStableRemoteConfiguration,
		},
		{
			name:                 "cluster is served with canary version of remote config",
			canaryArgument:       canaryUserAgent,
			remoteConfigFile:     validRulesFile,
			expectedRemoteConfig: validCanaryRemoteConfiguration,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, err := service.NewStorage(
				service.StorageConfig{
					RemoteConfigurationPath: v2Folder,
					ClusterMappingPath:      clusterMappingPath,
					ClusterMappingFile:      clusterMappingFile,
				}, true, &MockUnleashClient{})
			assert.NoError(t, err)
			req, err := http.NewRequest("GET", "http://example.com", nil)
			assert.NoError(t, err)
			req.Header.Add("User-Agent", tt.canaryArgument)
			checkRemoteConfig(t, storage, tt.remoteConfigFile, tt.expectedRemoteConfig, req)
		})
	}
}

func TestGetClusterID(t *testing.T) {
	tests := []struct {
		name           string
		userAgent      string
		expectedResult string
	}{
		{
			name:           "header with cluster ID",
			userAgent:      canaryUserAgent,
			expectedResult: canaryClusterID,
		},
		{
			name:           "header without cluster ID",
			userAgent:      "Go-http-client/1.1",
			expectedResult: "",
		},
		{
			name:           "header with cluster ID and extra suffix separated by comma",
			userAgent:      canaryUserAgent + ", extra data",
			expectedResult: canaryClusterID,
		},
		{
			name:           "header with cluster ID and extra suffix separated by space",
			userAgent:      canaryUserAgent + " extra data",
			expectedResult: canaryClusterID,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "http://example.com", nil)
			assert.NoError(t, err)
			req.Header.Add("User-Agent", tt.userAgent)

			clusterID := service.GetClusterID(req)
			assert.Equal(t, clusterID, tt.expectedResult)
		})
	}
}
