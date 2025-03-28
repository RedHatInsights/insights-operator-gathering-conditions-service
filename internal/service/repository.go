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
	"fmt"
	"net/http"
)

// RepositoryInterface defines methods to be implemented by any rules providers
type RepositoryInterface interface {
	Rules(r *http.Request) (*Rules, error)
	RemoteConfiguration(r *http.Request, ocpVersion string) (*RemoteConfiguration, error)
}

// Rule data type definition based on original JSON schema
type Rule struct {
	Conditions         []interface{} `json:"conditions,omitempty"`
	GatheringFunctions interface{}   `json:"gathering_functions,omitempty"`
}

// Rules data type definition based on original JSON schema
type Rules struct {
	Items   []Rule `json:"rules,omitempty"`
	Version string `json:"version,omitempty"`
}

// ContainerLogRequest defines a type for requesting container
// log data
type ContainerLogRequest struct {
	Namespace    string   `json:"namespace"`
	PodNameRegex string   `json:"pod_name_regex"`
	Previous     bool     `json:"previous,omitempty"`
	Messages     []string `json:"messages"`
}

// RemoteConfiguration represents the new data structure
// served by the v2 API
type RemoteConfiguration struct {
	ConditionalRules      []Rule                `json:"conditional_gathering_rules"`
	ContainerLogsRequests []ContainerLogRequest `json:"container_logs"`
	Version               string                `json:"version"`
}

// Repository is definition of objects that implement the RepositoryInterface
type Repository struct {
	store StorageInterface
}

// NewRepository constructs new instance of Repository
func NewRepository(s StorageInterface) *Repository {
	return &Repository{store: s}
}

// Rules method reads all and unmarshals all rules stored under given path
func (r *Repository) Rules(request *http.Request) (*Rules, error) {
	filepath := "rules.json" // TODO: Make this configurable
	data := r.store.ReadConditionalRules(r.store.IsCanary(request), filepath)
	if data == nil {
		return nil, fmt.Errorf("store data not found for '%s'", filepath)
	}

	var rules Rules
	err := json.Unmarshal(data, &rules)
	if err != nil {
		return nil, err
	}

	return &rules, nil
}

// RemoteConfiguration returns a remote configuration for v2 endpoint based on
// the cluster map defined in the settings and loaded on startup
func (r *Repository) RemoteConfiguration(request *http.Request, ocpVersion string) (*RemoteConfiguration, error) {
	isCanary := r.store.IsCanary(request)
	filepath, err := r.store.GetRemoteConfigurationFilepath(isCanary, ocpVersion)
	if err != nil {
		return nil, err
	}
	data := r.store.ReadRemoteConfig(filepath)
	if data == nil {
		return nil, fmt.Errorf("store data not found for '%s'", filepath)
	}
	var remoteConfig RemoteConfiguration
	err = json.Unmarshal(data, &remoteConfig)
	if err != nil {
		return nil, err
	}

	// Count the number of times a given remote configuration is returned
	remoteConfigurationsMetric.WithLabelValues(filepath, remoteConfig.Version).Inc()

	return &remoteConfig, nil
}
