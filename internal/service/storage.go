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

package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/blang/semver/v4"

	"github.com/rs/zerolog/log"

	merrors "github.com/RedHatInsights/insights-operator-gathering-conditions-service/internal/errors"
)

// StorageInterface describe interface to be implemented by resource storage
// implementations.
type StorageInterface interface {
	ReadConditionalRules(res string) []byte
	ReadRemoteConfig(p string) []byte
	GetRemoteConfigurationFilepath(ocpVersion string) (string, error)
}

// StorageConfig structure contains configuration for resource storage.
type StorageConfig struct {
	RulesPath               string `mapstructure:"rules_path" toml:"rules_path"`
	RemoteConfigurationPath string `mapstructure:"remote_configuration" toml:"remote_configuration"`
	StableVersion           string `mapstructure:"stable_version" toml:"stable_version"`
	ClusterMappingPath      string `mapstructure:"cluster_mapping" toml:"cluster_mapping"`
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

// Storage type represents container for resources.
type Storage struct {
	conditionalRulesPath    string
	remoteConfigurationPath string
	stableVersion           string
	cache                   Cache
	clusterMappingPath      string
	clusterMapping          ClusterMapping
}

// NewStorage constructs new storage object.
func NewStorage(cfg StorageConfig) (*Storage, error) {
	log.Debug().Interface("config", cfg).Msg("Constructing storage object")

	s := Storage{
		conditionalRulesPath:    cfg.RulesPath,
		remoteConfigurationPath: cfg.RemoteConfigurationPath,
		stableVersion:           cfg.StableVersion,
		clusterMappingPath:      cfg.ClusterMappingPath,
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

	if cm.IsValid(s.remoteConfigurationPath, s.stableVersion) {
		log.Info().Msg("The cluster map JSON is valid")
		s.clusterMapping = cm
	} else {
		log.Error().Msg("Cluster map is invalid")
		return nil, errors.New("cannot parse cluster map")
	}

	return &s, nil
}

// ReadConditionalRules tries to find conditional rule with given name in the storage.
func (s *Storage) ReadConditionalRules(path string) []byte {
	log.Debug().Str("path to resource", path).Msg("Finding resource")
	conditionalRulesPath := fmt.Sprintf("%s/%s/%s", s.conditionalRulesPath, s.stableVersion, path)
	return s.readDataFromPath(conditionalRulesPath)
}

// ReadRemoteConfig tries to find remote configuration with given name in the storage
func (s *Storage) ReadRemoteConfig(path string) []byte {
	log.Debug().Str("path to resource", path).Msg("Finding resource")
	remoteConfigPath := fmt.Sprintf("%s/%s/%s", s.remoteConfigurationPath, s.stableVersion, path)
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
