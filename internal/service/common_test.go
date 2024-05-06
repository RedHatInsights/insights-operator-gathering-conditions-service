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

package service_test

import "github.com/RedHatInsights/insights-operator-gathering-conditions-service/internal/service"

var (
	validRules = service.Rules{
		Version: "0.0.1",
		Items: []service.Rule{
			{
				Conditions: []interface{}{
					"condition 1",
					"condition 2",
				},
				GatheringFunctions: "the gathering functions",
			},
		},
	}
	validRulesJSON = `
	{
		"version": "0.0.1",
		"rules": [
			{
				"conditions": ["condition 1", "condition 2"],
				"gathering_functions": "the gathering functions"
			}
		]
	}
	`
	validRemoteConfiguration = service.RemoteConfiguration{
		Version: "0.0.1",
		ConditionalRules: []service.Rule{
			{
				Conditions: []interface{}{
					"condition 1",
					"condition 2",
				},
				GatheringFunctions: "the gathering functions",
			},
		},
		ContainerLogsRequests: []service.ContainerLogRequest{
			{
				Namespace:    "namespace-1",
				Previous:     "true",
				PodNameRegex: "test regex",
				Messages: []string{
					"first message",
					"second message",
				},
			},
		},
	}
	validRemoteConfigurationJSON = `
	{
		"version": "0.0.1",
		"conditional_gathering_rules": [
			{
				"conditions": ["condition 1", "condition 2"],
				"gathering_functions": "the gathering functions"
			}
		],
		"container_logs": [
			{
				"namespace": "namespace-1",
				"previous": "true",
				"pod_name_regex": "test regex",
				"messages": [
					"first message",
					"second message"
				]
			}
		]
	}
	`
)

type mockStorage struct {
	conditionalRules []byte
	remoteConfig     []byte
}

func (m *mockStorage) ReadConditionalRules(string) []byte {
	return m.conditionalRules
}

func (m *mockStorage) ReadRemoteConfig(string) []byte {
	return m.remoteConfig
}
