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
	"fmt"
	"io"
	"os"
	"reflect"
	"sync"

	"golang.org/x/mod/semver"

	"github.com/rs/zerolog/log"
)

// StorageInterface describe interface to be implemented by resource storage
// implementations.
type StorageInterface interface {
	ReadConditionalRules(res string) []byte
	ReadRemoteConfig(p string) []byte
	GetRemoteConfigurationFilepath(ocpVersion string) string
}

// StorageConfig structure contains configuration for resource storage.
type StorageConfig struct {
	RulesPath               string `mapstructure:"rules_path" toml:"rules_path"`
	RemoteConfigurationPath string `mapstructure:"remote_configuration" toml:"remote_configuration"`
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
	cache                   Cache
	clusterMappingPath      string
	clusterMapping          [][]string // TODO: Make custom type with validation and so on
}

// NewStorage constructs new storage object.
func NewStorage(cfg StorageConfig) *Storage {
	log.Debug().Interface("config", cfg).Msg("Constructing storage object")

	s := Storage{
		conditionalRulesPath:    cfg.RulesPath,
		remoteConfigurationPath: cfg.RemoteConfigurationPath,
		clusterMappingPath:      cfg.ClusterMappingPath, // TODO: Test this
	}

	// Read the cluster map
	cm := [][]string{}
	err := json.Unmarshal(
		s.readDataFromPath(s.clusterMappingPath),
		&cm)
	if err != nil {
		// TODO: Break here or log an error
		log.Error().Err(err).Msg("Cannot load cluster map")
		return &s
	}

	log.Debug().Interface("cluster-mapping", cm).Msg("Cluster mapping loaded")

	versions := []string{}
	for _, slice := range cm {
		if len(slice) != 2 {
			log.Error().Int("len", len(slice)).Strs("slice", slice).Msg("Unexpected slice length")
		}
		version := slice[0]
		if !semver.IsValid(version) {
			log.Error().Str("version", version).Msg("Invalid semver")
		} else {
			log.Debug().Str("version", version).Msg("Valid semver")
		}
		versions = append(versions, version)
	}

	// TODO: Add support for non v* versions

	// Check if the cluster mapping is sorted
	sortedVersions := make([]string, len(versions))
	copy(sortedVersions, versions)
	semver.Sort(sortedVersions)
	if !reflect.DeepEqual(sortedVersions, versions) {
		log.Error().Strs("sortedVersions", sortedVersions).Strs("versions", versions).Msg("Cluster mapping is not sorted")
	}

	s.clusterMapping = cm
	return &s
}

// ReadConditionalRules tries to find conditional rule with given name in the storage.
func (s *Storage) ReadConditionalRules(path string) []byte {
	log.Debug().Str("path to resource", path).Msg("Finding resource")
	conditionalRulesPath := fmt.Sprintf("%s/%s", s.conditionalRulesPath, path)
	return s.readDataFromPath(conditionalRulesPath)
}

// ReadRemoteConfig tries to find remote configuration with given name in the storage
func (s *Storage) ReadRemoteConfig(path string) []byte {
	log.Debug().Str("path to resource", path).Msg("Finding resource")
	remoteConfigPath := fmt.Sprintf("%s/%s", s.remoteConfigurationPath, path)
	return s.readDataFromPath(remoteConfigPath)
}

// GetRemoteConfigurationFilepath returns the filepath to the remote configuration
// that should be returned for the given OCP version based on the cluster map
func (s *Storage) GetRemoteConfigurationFilepath(ocpVersion string) string {
	if !semver.IsValid(ocpVersion) {
		log.Error().Str("ocpVersion", ocpVersion).Msg("Invalid semver")
		// TODO: return 404 or 400
		return "config_default.json"
	}

	for _, slice := range s.clusterMapping {
		version := slice[0]
		filepath := slice[1]

		if !semver.IsValid(version) {
			log.Error().Str("version", version).Msg("Invalid semver")
			break
		}

		log.Debug().Str("ocpVersion", ocpVersion).Str("version", version).Int("comparison", semver.Compare(ocpVersion, version)).Msg("comparing semver")
		if semver.Compare(ocpVersion, version) <= 0 {
			return filepath
		}

	}
	return "config_default.json"
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
