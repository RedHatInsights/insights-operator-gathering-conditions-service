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
	"fmt"
	"io/ioutil"
	"os"

	"github.com/rs/zerolog/log"
)

type StorageInterface interface {
	Find(res string) []byte
}

type StorageConfig struct {
	RulesPath string `mapstructure:"rules_path" toml:"rules_path"`
}

type Storage struct {
	path  string
	cache map[string][]byte
}

func NewStorage(cfg StorageConfig) *Storage {
	return &Storage{
		path:  cfg.RulesPath,
		cache: make(map[string][]byte),
	}
}

func (s *Storage) Find(path string) []byte {
	// use the in-memory data
	data, ok := s.cache[path]
	if ok {
		return data
	}

	// or try to load it from the file
	data, err := s.readFile(path)
	if err != nil {
		log.Warn().Msgf("Resource not found: '%s'", path)
		return nil
	}

	return data
}

func (s *Storage) readFile(path string) ([]byte, error) {
	f, err := os.Open(fmt.Sprintf("%s/%s", s.path, path))
	if err != nil {
		return nil, err
	}
	defer func() {
		err := f.Close()
		if err != nil {
			log.Error().Err(err).Msgf("Close file %s", path)
		}
	}()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	// add the bytes to cache
	s.cache[path] = data

	return data, nil
}
