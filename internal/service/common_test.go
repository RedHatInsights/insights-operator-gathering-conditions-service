/*
Copyright © 2022, 2024 Red Hat, Inc.

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

import (
	"net/http"

	"github.com/RedHatInsights/insights-operator-gathering-conditions-service/internal/service"
)

var (
	validStableRules = service.Rules{
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
	validStableRulesJSON = `
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
	validCanaryRules = service.Rules{
		Version: "0.0.2",
		Items: []service.Rule{
			{
				Conditions: []interface{}{
					"condition 2",
					"condition 3",
				},
				GatheringFunctions: "the gathering functions",
			},
		},
	}
	validStableRemoteConfiguration = service.RemoteConfiguration{
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
				Previous:     true,
				PodNameRegex: "test regex",
				Messages: []string{
					"first message",
					"second message",
				},
			},
		},
	}
	validStableRemoteConfigurationJSON = `
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
				"previous": true,
				"pod_name_regex": "test regex",
				"messages": [
					"first message",
					"second message"
				]
			}
		]
	}
	`
	validCanaryRemoteConfiguration = service.RemoteConfiguration{
		Version: "0.0.2",
		ConditionalRules: []service.Rule{
			{
				Conditions: []interface{}{
					"condition 2",
					"condition 3",
				},
				GatheringFunctions: "the gathering functions",
			},
		},
		ContainerLogsRequests: []service.ContainerLogRequest{
			{
				Namespace:    "namespace-1",
				Previous:     true,
				PodNameRegex: "test regex",
				Messages: []string{
					"first message",
					"second message",
				},
			},
		},
	}

	configDefaultConfiguration = `{"conditional_gathering_rules":[{"conditions":[{"type":"alert_is_firing","alert":{"name":"AlertmanagerFailedReload"}}],"gathering_functions":{"containers_logs":{"alert_name":"AlertmanagerFailedReload","container":"alertmanager","tail_lines":50}}},{"conditions":[{"type":"alert_is_firing","alert":{"name":"AlertmanagerFailedToSendAlerts"}}],"gathering_functions":{"containers_logs":{"alert_name":"AlertmanagerFailedToSendAlerts","tail_lines":50,"container":"alertmanager"}}},{"conditions":[{"type":"alert_is_firing","alert":{"name":"APIRemovedInNextEUSReleaseInUse"}}],"gathering_functions":{"api_request_counts_of_resource_from_alert":{"alert_name":"APIRemovedInNextEUSReleaseInUse"}}},{"conditions":[{"type":"alert_is_firing","alert":{"name":"KubePodCrashLooping"}}],"gathering_functions":{"containers_logs":{"alert_name":"KubePodCrashLooping","tail_lines":20,"previous":true}}},{"conditions":[{"type":"alert_is_firing","alert":{"name":"KubePodNotReady"}}],"gathering_functions":{"containers_logs":{"alert_name":"KubePodNotReady","tail_lines":100},"pod_definition":{"alert_name":"KubePodNotReady"}}},{"conditions":[{"type":"alert_is_firing","alert":{"name":"PrometheusOperatorSyncFailed"}}],"gathering_functions":{"containers_logs":{"alert_name":"PrometheusOperatorSyncFailed","tail_lines":50,"container":"prometheus-operator"}}},{"conditions":[{"type":"alert_is_firing","alert":{"name":"PrometheusTargetSyncFailure"}}],"gathering_functions":{"containers_logs":{"alert_name":"PrometheusTargetSyncFailure","container":"prometheus","tail_lines":50}}},{"conditions":[{"type":"alert_is_firing","alert":{"name":"SamplesImagestreamImportFailing"}}],"gathering_functions":{"logs_of_namespace":{"namespace":"openshift-cluster-samples-operator","tail_lines":100},"image_streams_of_namespace":{"namespace":"openshift-cluster-samples-operator"}}},{"conditions":[{"type":"alert_is_firing","alert":{"name":"ThanosRuleQueueIsDroppingAlerts"}}],"gathering_functions":{"containers_logs":{"alert_name":"ThanosRuleQueueIsDroppingAlerts","container":"thanos-ruler","tail_lines":50}}}],"container_logs":[],"version":"1.1.0"}`
	emptyConfiguration         = `{"conditional_gathering_rules":[],"container_logs":[],"version":"1.1.0"}`
	experimental1Configuration = `{"conditional_gathering_rules":[{"conditions":[{"type":"alert_is_firing","alert":{"name":"Experimental1"}}],"gathering_functions":{"containers_logs":{"alert_name":"Experimental1","container":"experimental1","tail_lines":50}}}],"container_logs":[],"version":"1.1.0"}`
	bugWorkaroundConfiguration = `{"conditional_gathering_rules":[{"conditions":[{"type":"alert_is_firing","alert":{"name":"BugWorkaround"}}],"gathering_functions":{"containers_logs":{"alert_name":"BugWorkaround","container":"bug-workaround","tail_lines":50}}}],"container_logs":[],"version":"1.1.0"}`
	experimental2Configuration = `{"conditional_gathering_rules":[{"conditions":[{"type":"alert_is_firing","alert":{"name":"Experimental2"}}],"gathering_functions":{"containers_logs":{"alert_name":"Experimental2","container":"experimental2","tail_lines":50}}}],"container_logs":[],"version":"1.1.0"}`

	stableUserAgent = "insights-operator/4.14.27-$Format:%H$ cluster/9abc1e7a-d834-4c6d-99b1-826399958d1c"
	canaryUserAgent = "insights-operator/4.14.27-$Format:%H$ cluster/f9fbc65a-52e6-4781-979d-1d5c6b124f9b"
)

type mockStorage struct {
	conditionalRules                        []byte
	remoteConfig                            []byte
	remoteConfigFilepath                    string
	getRemoteConfigurationFilepathMockError error
}

func (m *mockStorage) ReadConditionalRules(*http.Request, string) []byte {
	return m.conditionalRules
}

func (m *mockStorage) ReadRemoteConfig(*http.Request, string) []byte {
	return m.remoteConfig
}

func (m *mockStorage) GetRemoteConfigurationFilepath(string) (string, error) {
	return m.remoteConfigFilepath, m.getRemoteConfigurationFilepathMockError
}
