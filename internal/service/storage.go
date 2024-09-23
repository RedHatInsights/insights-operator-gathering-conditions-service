/*
Copyright © 2021, 2022, 2024 Red Hat, Inc.

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

package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/Unleash/unleash-client-go/v4"
	"github.com/Unleash/unleash-client-go/v4/context"

	"github.com/blang/semver/v4"

	"github.com/rs/zerolog/log"

	merrors "github.com/RedHatInsights/insights-operator-gathering-conditions-service/internal/errors"
)

// StableVersion describes subdirectory with stable version of conditions and remote configurations
const StableVersion = "stable"

// CanaryVersion describes subdirectory with canary version of conditions and remote configurations
const CanaryVersion = "canary"

// UnleashClientInterface describes interface for using Unleash in canary rollouts
type UnleashClientInterface interface {
	IsCanary(clusterID string) bool
}

// StorageInterface describe interface to be implemented by resource storage
// implementations.
type StorageInterface interface {
	ReadConditionalRules(res string, clusterID string) []byte
	ReadRemoteConfig(p string, clusterID string) []byte
	GetRemoteConfigurationFilepath(ocpVersion string) (string, error)
}

// StorageConfig structure contains configuration for resource storage.
type StorageConfig struct {
	RulesPath               string `mapstructure:"rules_path" toml:"rules_path"`
	RemoteConfigurationPath string `mapstructure:"remote_configuration" toml:"remote_configuration"`
	ClusterMappingPath      string `mapstructure:"cluster_mapping" toml:"cluster_mapping"`
	UnleashURL              string `mapstructure:"unleash_url" toml:"unleash_url"`
	UnleashToken            string `mapstructure:"unleash_token" toml:"unleash_token"`
	UnleashApp              string `mapstructure:"unleash_app" toml:"unleash_app"`
	UnleashToggle           string `mapstructure:"unleash_toggle" toml:"unleash_toggle"`
	UnleashEnabled          bool   `mapstructure:"unleash_enabled" toml:"unleash_enabled"`
}

// Cache type represents thread safe map for storing loaded configurations
type Cache struct {
	cache sync.Map
}

// Get retrieves value from the cache
func (c *Cache) Get(key string) []byte {
	data, _ := c.cache.Load(key)
	if data != nil {
		return data.([]byte)
	}
	return nil
}

// Set stores value under given key to the cache
func (c *Cache) Set(key string, value []byte) {
	c.cache.Store(key, value)
}

// UnleashClient initializes Unleash on its creation and provides interface to query it
type UnleashClient struct {
	unleashToggle string
}

// NewUnleashClient constructs new Unleash client along with Unleash initialization
func NewUnleashClient(cfg StorageConfig) (*UnleashClient, error) {
	c := UnleashClient{unleashToggle: cfg.UnleashToggle}
	log.Info().Msg("Initializing Unleash")
	err := unleash.Initialize(
		unleash.WithAppName(cfg.UnleashApp),
		unleash.WithUrl(cfg.UnleashURL),
		unleash.WithCustomHeaders(http.Header{"Authorization": {cfg.UnleashToken}}),
	)
	if err != nil {
		return nil, err
	}
	unleash.WaitForReady()
	log.Info().Msg("Unleash initialized")
	return &c, nil
}

// IsCanary queries Unleash to determine whether to serve stable or canary version of data
func (c *UnleashClient) IsCanary(clusterID string) bool {
	return unleash.IsEnabled(c.unleashToggle, unleash.WithContext(context.Context{UserId: clusterID}))
}

// Storage type represents container for resources.
type Storage struct {
	conditionalRulesPath    string
	remoteConfigurationPath string
	cache                   Cache
	clusterMappingPath      string
	clusterMapping          ClusterMapping
	unleashClient           UnleashClientInterface
	unleashEnabled          bool
}

// NewStorage constructs new storage object.
func NewStorage(cfg StorageConfig, unleashClient UnleashClientInterface) (*Storage, error) {
	log.Debug().Interface("config", cfg).Msg("Constructing storage object")
	s := Storage{
		conditionalRulesPath:    cfg.RulesPath,
		remoteConfigurationPath: cfg.RemoteConfigurationPath,
		clusterMappingPath:      cfg.ClusterMappingPath,
		unleashEnabled:          cfg.UnleashEnabled,
		unleashClient:           unleashClient,
	}

	if s.clusterMappingPath == "" {
		errStr := "cluster mapping filepath is not defined"
		log.Error().Msg(errStr)
		return &s, errors.New(errStr)
	}
	// Parse the cluster map
	cm := ClusterMapping{}
	rawData := s.readDataFromPath(s.clusterMappingPath)
	if rawData == nil {
		return &s, errors.New("cannot find cluster map")
	}
	err := json.Unmarshal(rawData, &cm)
	if err != nil {
		log.Error().Err(err).Msg("Cannot load cluster map")
		return &s, err
	}

	log.Debug().Interface("cluster-map", cm).Msg("Cluster map loaded")

	if cm.IsValid(s.remoteConfigurationPath, StableVersion) {
		log.Info().Msg("The stable version of cluster map JSON is valid")
		s.clusterMapping = cm
	} else {
		log.Error().Msg("Stable version of cluster map is invalid")
		return nil, errors.New("cannot parse cluster map")
	}

	if cm.IsValid(s.remoteConfigurationPath, CanaryVersion) {
		log.Info().Msg("The canary version of cluster map JSON is valid")
	} else {
		log.Error().Msg("Canary version of cluster map is invalid")
		return nil, errors.New("cannot parse cluster map")
	}

	return &s, nil
}

// ReadConditionalRules tries to find conditional rule with given name in the storage.
func (s *Storage) ReadConditionalRules(path string, clusterID string) []byte {
	log.Debug().Str("path to resource", path).Msg("Finding resource")
	version := StableVersion
	if s.unleashEnabled {
		if s.unleashClient.IsCanary(clusterID) {
			log.Debug().Str("cluster", clusterID).Msg("Served canary version of rules")
			version = CanaryVersion
		} else {
			log.Debug().Str("cluster", clusterID).Msg("Served stable version of rules")
		}
	}
	conditionalRulesPath := fmt.Sprintf("%s/%s/%s", s.conditionalRulesPath, version, path)
	return s.readDataFromPath(conditionalRulesPath)
}

// ReadRemoteConfig tries to find remote configuration with given name in the storage
func (s *Storage) ReadRemoteConfig(path string, clusterID string) []byte {
	log.Debug().Str("path to resource", path).Msg("Finding resource")
	version := StableVersion
	if s.unleashEnabled {
		if s.unleashClient.IsCanary(clusterID) {
			log.Debug().Str("cluster", clusterID).Msg("Served canary version of remote configurations")
			version = CanaryVersion
		} else {
			log.Debug().Str("cluster", clusterID).Msg("Served stable version of remote configurations")
		}
	}
	remoteConfigPath := fmt.Sprintf("%s/%s/%s", s.remoteConfigurationPath, version, path)
	return s.readDataFromPath(remoteConfigPath)
}

// GetRemoteConfigurationFilepath returns the filepath to the remote configuration
// that should be returned for the given OCP version based on the cluster map
func (s *Storage) GetRemoteConfigurationFilepath(ocpVersion string) (string, error) {
	ocpVersionParsed, err := semver.Make(ocpVersion)
	if err != nil {
		log.Error().Str("ocpVersion", ocpVersion).Err(err).Msg("Invalid semver")
		return "", &merrors.RouterParsingError{
			ParamName:  "ocpVersion",
			ParamValue: ocpVersion,
			ErrString:  err.Error()}
	}

	return s.clusterMapping.GetFilepathForVersion(ocpVersionParsed)
}

func (s *Storage) readDataFromPath(path string) []byte {
	// use the in-memory data
	data := s.cache.Get(path)
	if data != nil {
		return data
	}

	// or try to load it from the file
	data, err := s.readFile(path)
	if err != nil {
		log.Warn().Msgf("Resource not found: '%s'", path)
		return nil
	}

	log.Debug().Int("bytes", len(data)).Msg("Resource file has been read")

	return data
}

func (s *Storage) readFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := f.Close()
		if err != nil {
			log.Error().Err(err).Msgf("Close file %s", path)
		}
	}()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	// add the bytes to cache
	s.cache.Set(path, data)

	return data, nil
}
