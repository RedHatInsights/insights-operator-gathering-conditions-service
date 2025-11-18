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
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Unleash/unleash-go-sdk/v5"
	"github.com/Unleash/unleash-go-sdk/v5/context"

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
	IsCanary(canaryArgument string) bool
}

// StorageInterface describe interface to be implemented by resource storage
// implementations.
type StorageInterface interface {
	IsCanary(request *http.Request) bool
	ReadConditionalRules(isCanary bool, res string) []byte
	ReadRemoteConfig(p string) []byte
	GetRemoteConfigurationFilepath(isCanary bool, ocpVersion string) (string, error)
}

// StorageConfig structure contains configuration for resource storage.
type StorageConfig struct {
	RulesPath                string `mapstructure:"rules_path" toml:"rules_path"`
	RemoteConfigurationsPath string `mapstructure:"remote_configurations" toml:"remote_configurations"`
}

// CanaryConfig structure contains configuration for canary rollout
type CanaryConfig struct {
	UnleashURL     string `mapstructure:"unleash_url" toml:"unleash_url"`
	UnleashToken   string `mapstructure:"unleash_token" toml:"unleash_token"`
	UnleashApp     string `mapstructure:"unleash_app" toml:"unleash_app"`
	UnleashToggle  string `mapstructure:"unleash_toggle" toml:"unleash_toggle"`
	UnleashEnabled bool   `mapstructure:"unleash_enabled" toml:"unleash_enabled"`
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
func NewUnleashClient(cfg CanaryConfig) (*UnleashClient, error) {
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
func (c *UnleashClient) IsCanary(canaryArgument string) bool {
	return unleash.IsEnabled(c.unleashToggle, unleash.WithContext(context.Context{UserId: canaryArgument}))
}

// Storage type represents container for resources.
type Storage struct {
	conditionalRulesPath     string
	remoteConfigurationsPath string
	cache                    Cache
	stableClusterMapping     *ClusterMapping
	canaryClusterMapping     *ClusterMapping
	unleashClient            UnleashClientInterface
	unleashEnabled           bool
}

// NewStorage constructs new storage object.
func NewStorage(storageConfig StorageConfig, unleashEnabled bool, unleashClient UnleashClientInterface) (*Storage, error) {
	log.Debug().Interface("config", storageConfig).Msg("Constructing storage object")
	s := Storage{
		conditionalRulesPath:     storageConfig.RulesPath,
		remoteConfigurationsPath: storageConfig.RemoteConfigurationsPath,
		unleashEnabled:           unleashEnabled,
		unleashClient:            unleashClient,
	}

	cm, err := s.loadClusterMapping(StableVersion)
	if err != nil {
		log.Error().Err(err).Msg("Could not load stable version of cluster mapping")
		return &s, err
	}
	s.stableClusterMapping = cm

	cm, err = s.loadClusterMapping(CanaryVersion)
	if err != nil {
		log.Error().Err(err).Msg("Could not load canary version of cluster mapping")
		return &s, err
	}
	s.canaryClusterMapping = cm

	return &s, nil
}

func (s *Storage) loadClusterMapping(version string) (*ClusterMapping, error) {
	if s.remoteConfigurationsPath == "" {
		errStr := "remote configurations directory path is not defined"
		log.Error().Msg(errStr)
		return nil, errors.New(errStr)
	}

	configsRootDir := filepath.Join(s.remoteConfigurationsPath, version)

	// Parse the cluster map
	cm := ClusterMapping{
		rootDir: configsRootDir,
		mapping: [][]string{},
	}

	fullFilepath := filepath.Join(configsRootDir, "cluster_version_mapping.json")
	log.Info().Msg(fullFilepath)
	rawData := s.readDataFromPath(fullFilepath)
	if rawData == nil {
		return nil, errors.New("cannot find cluster map")
	}
	err := json.Unmarshal(rawData, &cm.mapping)
	if err != nil {
		log.Error().Str("version", version).Err(err).Msg("Cannot load cluster map")
		return nil, err
	}

	log.Debug().Interface("cluster-map", cm.mapping).Msg("Cluster map loaded")

	if cm.IsValid() {
		log.Info().Str("version", version).Msg("The cluster map JSON is valid")
	} else {
		log.Error().Str("version", version).Msg("The cluster map is invalid")
		return nil, errors.New("cannot parse cluster map")
	}

	return &cm, nil
}

// IsCanary queries UnleashClient to determine which version of configurations to serve
func (s *Storage) IsCanary(r *http.Request) bool {
	if !s.unleashEnabled {
		return false
	}
	// We use User-Agent header to decide between stable and canary version (header contains cluster ID)
	clusterID := GetClusterID(r)
	isCanary := s.unleashClient.IsCanary(clusterID)
	if isCanary {
		log.Debug().Str("canary argument", clusterID).Msg("Serving canary version of configurations")
	} else {
		log.Debug().Str("canary argument", clusterID).Msg("Serving stable version of configurations")
	}
	return isCanary
}

// ReadConditionalRules tries to find conditional rule with given name in the storage.
func (s *Storage) ReadConditionalRules(isCanary bool, path string) []byte {
	log.Debug().Str("path to resource", path).Msg("Finding resource")
	version := StableVersion
	if isCanary {
		version = CanaryVersion
	}
	conditionalRulesPath := filepath.Join(s.conditionalRulesPath, version, path)
	return s.readDataFromPath(conditionalRulesPath)
}

// ReadRemoteConfig tries to find remote configuration with given path in the storage
func (s *Storage) ReadRemoteConfig(path string) []byte {
	log.Debug().Str("path to resource", path).Msg("Finding resource")
	return s.readDataFromPath(path)
}

// GetRemoteConfigurationFilepath returns the filepath to the remote configuration
// that should be returned for the given OCP version based on the cluster map
func (s *Storage) GetRemoteConfigurationFilepath(isCanary bool, ocpVersion string) (string, error) {
	ocpVersionParsed, err := semver.Make(ocpVersion)
	if err != nil {
		log.Info().Str("ocpVersion", ocpVersion).Err(err).Msg("Invalid semver")
		return "", &merrors.RouterParsingError{
			ParamName:  "ocpVersion",
			ParamValue: ocpVersion,
			ErrString:  err.Error()}
	}

	if isCanary {
		return s.canaryClusterMapping.GetFilepathForVersion(ocpVersionParsed)
	}
	return s.stableClusterMapping.GetFilepathForVersion(ocpVersionParsed)
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
	f, err := os.Open(path) // #nosec G304
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

// GetClusterID obtain the cluster ID from user agent
func GetClusterID(r *http.Request) string {
	userAgent := r.UserAgent()
	if !strings.Contains(userAgent, "cluster/") {
		err := errors.New("UserAgent does not contain cluster ID")
		log.Warn().Str("UserAgent", userAgent).Err(err).Msg("Failed to retrieve cluster ID")
		return ""
	}
	_, clusterID, _ := strings.Cut(userAgent, "cluster/")

	// Get rid of any text that would follow after cluster ID
	clusterID = strings.Split(clusterID, " ")[0]
	clusterID = strings.Split(clusterID, ",")[0]
	return clusterID
}
