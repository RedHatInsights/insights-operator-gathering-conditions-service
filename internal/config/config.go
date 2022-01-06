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

package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/RedHatInsights/insights-operator-utils/logger"
	"github.com/redhatinsights/insights-operator-conditional-gathering/internal/server"
	"github.com/redhatinsights/insights-operator-conditional-gathering/internal/service"
	"github.com/spf13/viper"
)

const (
	// configFileEnvVariableName is name of environment variable that
	// contains name of configuration file
	configFileEnvVariableName = "INSIGHTS_OPERATOR_CONDITIONAL_SERVICE_CONFIG_FILE"

	// envPrefix is prefix for all environment variables that contains
	// various configuration options
	envPrefix = "INSIGHTS_OPERATOR_CONDITIONAL_SERVICE_"
)

// Configuration is a structure holding the whole service configuration
type Configuration struct {
	ServerConfig        server.Config                     `mapstructure:"server" toml:"server"`
	StorageConfig       service.StorageConfig             `mapstructure:"storage" toml:"storage"`
	LoggingConfig       logger.LoggingConfiguration       `mapstructure:"logging" toml:"logging"`
	CloudWatchConfig    logger.CloudWatchConfiguration    `mapstructure:"cloudwatch" toml:"cloudwatch"`
	SentryLoggingConfig logger.SentryLoggingConfiguration `mapstructure:"sentry" toml:"sentry"`
	KafkaZerologConfig  logger.KafkaZerologConfiguration  `mapstructure:"kafka_zerolog" toml:"kafka_zerolog"`
}

var Config Configuration

// LoadConfiguration loads configuration from defaultConfigFile, file set in
// configFileEnvVariableName or from env
func LoadConfiguration(defaultConfigFile string) error {
	configFile, specified := os.LookupEnv(configFileEnvVariableName)
	if specified {
		// we need to separate the directory name and filename without
		// extension
		directory, basename := filepath.Split(configFile)
		file := strings.TrimSuffix(basename, filepath.Ext(basename))
		// parse the configuration
		viper.SetConfigName(file)
		viper.AddConfigPath(directory)
	} else {
		// parse the configuration
		viper.SetConfigName(defaultConfigFile)
		viper.AddConfigPath(".")
	}

	err := viper.ReadInConfig()
	if _, isNotFoundError := err.(viper.ConfigFileNotFoundError); !specified && isNotFoundError {
		// viper is not smart enough to understand the structure of
		// config by itself
		fakeTomlConfigWriter := new(bytes.Buffer)

		err = toml.NewEncoder(fakeTomlConfigWriter).Encode(Config)
		if err != nil {
			return err
		}

		fakeTomlConfig := fakeTomlConfigWriter.String()

		viper.SetConfigType("toml")

		err = viper.ReadConfig(strings.NewReader(fakeTomlConfig))
		if err != nil {
			return err
		}
	} else if err != nil {
		return fmt.Errorf("fatal error config file: %s", err)
	}

	// override config from env if there's variable in env
	viper.AutomaticEnv()
	viper.SetEnvPrefix(envPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "__"))

	err = viper.Unmarshal(&Config)
	if err != nil {
		return fmt.Errorf("fatal error config file: %s", err)
	}

	// everything's should be ok
	return nil
}

// ServerConfig function returns actual server configuration.
func ServerConfig() server.Config {
	return Config.ServerConfig
}

// StorageConfig function returns actual storage configuration.
func StorageConfig() service.StorageConfig {
	return Config.StorageConfig
}

// LoggingConfig function returns actual logger configuration.
func LoggingConfig() logger.LoggingConfiguration {
	return Config.LoggingConfig
}

// CloudWatchConfig function returns actual CloudWatch configuration.
func CloudWatchConfig() logger.CloudWatchConfiguration {
	return Config.CloudWatchConfig
}

// SentryLoggingConfig function returns the sentry log configuration.
func SentryLoggingConfig() logger.SentryLoggingConfiguration {
	return Config.SentryLoggingConfig
}

// KafkaZerologConfig function returns the configuration of ZeroLog for Kafka.
func KafkaZerologConfig() logger.KafkaZerologConfiguration {
	return Config.KafkaZerologConfig
}
