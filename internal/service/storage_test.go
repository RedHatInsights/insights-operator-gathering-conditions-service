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
	"testing"

	"github.com/RedHatInsights/insights-operator-gathering-conditions-service/internal/service"
	"github.com/stretchr/testify/assert"
)

const (
	validRulesFile         = "rules.json"
	rulesFolder            = "testdata/v1"
	v2Folder               = "testdata/v2"
	clusterMappingFilepath = "testdata/cluster-mapping.json"
	stableClusterID        = "27badff1-9f61-4168-a309-6878a66075b3"
	canaryClusterID        = "f9fbc65a-52e6-4781-979d-1d5c6b124f9b"
)

type MockUnleashClient struct{}

func (c *MockUnleashClient) IsCanary(clusterID string) bool {
	return clusterID == canaryClusterID
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
				ClusterMappingPath:      "testdata/cluster-mapping.json",
				UnleashEnabled:          false,
			},
			expectError: false,
		},
		{
			name: "config is invalid: stable version of data does not exit",
			config: service.StorageConfig{
				RulesPath:               "testdata/v1",
				RemoteConfigurationPath: "testdata/v2_missing_stable",
				ClusterMappingPath:      "testdata/cluster-mapping.json",
				UnleashEnabled:          false,
			},
			expectError: true,
		},
		{
			name: "config is invalid: canary version of data does not exit",
			config: service.StorageConfig{
				RulesPath:               "testdata/v1",
				RemoteConfigurationPath: "testdata/v2_missing_canary",
				ClusterMappingPath:      "testdata/cluster-mapping.json",
				UnleashEnabled:          false,
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := service.NewStorage(tc.config, nil)
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
					ClusterMappingPath:      clusterMappingFilepath,
					RemoteConfigurationPath: v2Folder,
				}, nil)
			assert.NoError(t, err)
			checkConditionalRules(t, storage, tc.rulesFile, tc.expectedRules, stableClusterID)
		})
	}

	t.Run("use cache with a previous read file", func(t *testing.T) {
		// run Find anothertime to test the cache function
		storage, err := service.NewStorage(
			service.StorageConfig{
				RulesPath:               rulesFolder,
				ClusterMappingPath:      clusterMappingFilepath,
				RemoteConfigurationPath: v2Folder,
			}, nil)
		assert.NoError(t, err)
		for i := 0; i < 2; i++ {
			checkConditionalRules(t, storage, validRulesFile, validStableRules, stableClusterID)
		}
	})
}

func TestReadRulesCanaryRollout(t *testing.T) {
	tests := []struct {
		name             string
		clusterID        string
		remoteConfigFile string
		expectedRules    service.Rules
	}{
		{
			name:             "cluster is served with stable version of rules",
			clusterID:        stableClusterID,
			remoteConfigFile: validRulesFile,
			expectedRules:    validStableRules,
		},
		{
			name:             "cluster is served with canary version of rules",
			clusterID:        canaryClusterID,
			remoteConfigFile: validRulesFile,
			expectedRules:    validCanaryRules,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, err := service.NewStorage(
				service.StorageConfig{
					RulesPath:               rulesFolder,
					ClusterMappingPath:      clusterMappingFilepath,
					RemoteConfigurationPath: v2Folder,
					UnleashEnabled:          true,
				}, &MockUnleashClient{})
			assert.NoError(t, err)
			checkConditionalRules(t, storage, tt.remoteConfigFile, tt.expectedRules, tt.clusterID)
		})
	}
}

func checkConditionalRules(t *testing.T, storage *service.Storage, rulesFile string, expectedRules service.Rules, clusterID string) {
	var rules service.Rules
	data := storage.ReadConditionalRules(rulesFile, clusterID)
	if len(data) == 0 {
		rules = service.Rules{}
	} else {
		err := json.Unmarshal(data, &rules)
		assert.NoError(t, err)
	}
	assert.Equal(t, expectedRules, rules)
}

func checkRemoteConfig(t *testing.T, storage *service.Storage, remoteConfigFile string, expectedRemoteConfig service.RemoteConfiguration, clusterID string) {
	var remoteConfig service.RemoteConfiguration
	data := storage.ReadRemoteConfig(remoteConfigFile, clusterID)
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
					ClusterMappingPath:      clusterMappingFilepath,
				}, nil)
			assert.NoError(t, err)
			checkRemoteConfig(t, storage, tt.remoteConfigFile, tt.expectedRemoteConfig, stableClusterID)
		})
	}
}

func TestReadRemoteConfigurationCanaryRollout(t *testing.T) {
	tests := []struct {
		name                 string
		clusterID            string
		remoteConfigFile     string
		expectedRemoteConfig service.RemoteConfiguration
	}{
		{
			name:                 "cluster is served with stable version of remote config",
			clusterID:            stableClusterID,
			remoteConfigFile:     validRulesFile,
			expectedRemoteConfig: validStableRemoteConfiguration,
		},
		{
			name:                 "cluster is served with canary version of remote config",
			clusterID:            canaryClusterID,
			remoteConfigFile:     validRulesFile,
			expectedRemoteConfig: validCanaryRemoteConfiguration,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, err := service.NewStorage(
				service.StorageConfig{
					RemoteConfigurationPath: v2Folder,
					ClusterMappingPath:      clusterMappingFilepath,
					UnleashEnabled:          true,
				}, &MockUnleashClient{})
			assert.NoError(t, err)
			checkRemoteConfig(t, storage, tt.remoteConfigFile, tt.expectedRemoteConfig, tt.clusterID)
		})
	}
}
