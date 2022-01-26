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

package config_test

import (
	"os"
	"testing"

	"github.com/RedHatInsights/insights-operator-gathering-conditions-service/internal/config"
	"github.com/RedHatInsights/insights-operator-gathering-conditions-service/internal/server"
	"github.com/RedHatInsights/insights-operator-gathering-conditions-service/internal/service"
	"github.com/RedHatInsights/insights-operator-utils/logger"
	"github.com/stretchr/testify/assert"
)

const (
	validConfPath       = "testdata/valid-config"
	invalidConfPath     = "testdata/invalid-config"
	nonExistentConfPath = "testdata/notfound-config"
)

var (
	validConf = config.Configuration{
		ServerConfig: server.Config{
			Address:    "address",
			UseHTTPS:   true,
			EnableCORS: true,
		},
		AuthConfig: server.AuthConfig{
			Enabled: false,
			Type:    "",
		},
		StorageConfig: service.StorageConfig{
			RulesPath: "rules_path",
		},
		SentryLoggingConfig: logger.SentryLoggingConfiguration{
			SentryDSN: "dsn",
		},
		KafkaZerologConfig: logger.KafkaZerologConfiguration{
			Broker:   "broker",
			Topic:    "topic",
			CertPath: "cert_path",
			Level:    "level",
		},
	}
	emptyConfig = config.Configuration{}

	customAddress = "custom address"
	customConfig  = config.Configuration{
		ServerConfig: server.Config{
			Address: customAddress,
		},
	}
)

type testCase struct {
	name                            string
	configPath                      string
	shouldExist                     bool
	expectedConfiguration           config.Configuration
	expectErrorLoadingConfiguration bool
	envVariables                    map[string]string
}

func TestLoadConfiguration(t *testing.T) {
	testCases := []testCase{
		{
			"file exists and configuration is valid",
			validConfPath,
			true,
			validConf,
			false,
			nil,
		},
		{
			"file exists and configuration is invalid",
			invalidConfPath,
			true,
			emptyConfig,
			true,
			nil,
		},
		{
			"file doesn't exist",
			nonExistentConfPath,
			false,
			emptyConfig,
			false,
			nil,
		},
		{
			"set the configuration file as environment variable",
			nonExistentConfPath,
			false,
			validConf,
			false,
			map[string]string{
				"INSIGHTS_OPERATOR_GATHERING_CONDITIONS_SERVICE_CONFIG_FILE": validConfPath,
			},
		},
		{
			"change a configuration field using an environment variable",
			nonExistentConfPath,
			false,
			customConfig,
			false,
			map[string]string{
				"INSIGHTS_OPERATOR_GATHERING_CONDITIONS_SERVICE__SERVER__ADDRESS": customAddress,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Clearenv()
			config.Config = config.Configuration{} // reset the configuration so that is not loaded from previous test case

			for k, v := range tc.envVariables {
				err := os.Setenv(k, v)
				assert.NoError(t, err, "didn't expect any error while setting the environment variables")
			}

			if tc.shouldExist {
				assert.FileExists(t, tc.configPath+".toml", "this file should exist")
			} else {
				assert.NoFileExists(t, tc.configPath+".toml", "this file shouldn't exist")
			}
			err := config.LoadConfiguration(tc.configPath)
			if tc.expectErrorLoadingConfiguration {
				assert.Error(t, err, "expected error loading configuration")
			} else {
				assert.NoError(t, err, "error loading configuration")
				assert.Equal(t, tc.expectedConfiguration, config.Config)
			}
		})
	}

}

func TestGetConfigFunctions(t *testing.T) {
	os.Clearenv()
	assert.NoError(t, config.LoadConfiguration(validConfPath))

	t.Run("ServerConfig", func(t *testing.T) {
		assert.Equal(t, config.Config.ServerConfig, config.ServerConfig())
	})
	t.Run("AuthConfig", func(t *testing.T) {
		assert.Equal(t, config.Config.AuthConfig, config.AuthConfig())
	})
	t.Run("StorageConfig", func(t *testing.T) {
		assert.Equal(t, config.Config.StorageConfig, config.StorageConfig())
	})
	t.Run("LoggingConfig", func(t *testing.T) {
		assert.Equal(t, config.Config.LoggingConfig, config.LoggingConfig())
	})
	t.Run("CloudWatchConfig", func(t *testing.T) {
		assert.Equal(t, config.Config.CloudWatchConfig, config.CloudWatchConfig())
	})
	t.Run("SentryLoggingConfig", func(t *testing.T) {
		assert.Equal(t, config.Config.SentryLoggingConfig, config.SentryLoggingConfig())
	})
	t.Run("KafkaZerologConfig", func(t *testing.T) {
		assert.Equal(t, config.Config.KafkaZerologConfig, config.KafkaZerologConfig())
	})
}
