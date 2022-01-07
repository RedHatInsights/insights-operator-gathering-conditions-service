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
)

// RepositoryInterface defines methods to be implemented by any rules providers
type RepositoryInterface interface {
	Rules() (*Rules, error)
}

// Rule data type definition based on original JSON schema
type Rule struct {
	Conditions         []interface{} `json:"conditions,omitempty"`
	GatheringFunctions interface{}   `json:"gathering_functions,omitempty"`
}

// Rules data type definition based on original JSON schema
type Rules struct {
	Items []Rule `json:"rules,omitempty"`
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
func (r *Repository) Rules() (*Rules, error) {
	filepath := "rules.json" // TODO: Make this configurable
	data := r.store.Find(filepath)
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
