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

// RulesProvider defines methods to be implemented by any rules provider
type RulesProvider interface {
	Rules() (*Rules, error)
	RemoteConfiguration() (*RemoteConfiguration, error)
}

// Service data type represents the whole service for repository interface.
type Service struct {
	repo RepositoryInterface
}

// New function constructs new service for given repository interface.
func New(repo RepositoryInterface) *Service {
	return &Service{
		repo,
	}
}

// Rules method returns all rules provided by the service.
func (s *Service) Rules() (*Rules, error) {
	rules, err := s.repo.Rules()
	if err != nil {
		return nil, err
	}

	return rules, nil
}

// RemoteConfiguration method returns the remote configuration provided by the service.
func (s *Service) RemoteConfiguration() (*RemoteConfiguration, error) {
	remoteConfiguration, err := s.repo.RemoteConfiguration()
	if err != nil {
		return nil, err
	}

	return remoteConfiguration, nil
}
