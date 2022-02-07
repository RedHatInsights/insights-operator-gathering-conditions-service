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

package tests

import (
	"encoding/json"

	"github.com/verdverm/frisby"
)

// Rule data type definition based on original JSON schema
type Rule struct {
	Conditions         []interface{} `json:"conditions,omitempty"`
	GatheringFunctions interface{}   `json:"gathering_functions,omitempty"`
}

// Payload to be returned by server
type Payload struct {
	Version string `json:"version"`
	Rules   []Rule `json:"rules"`
}

// readGatheringRules read and decode JSON payload read from server
func readGatheringRules(f *frisby.Frisby) Payload {
	response := Payload{}
	text, err := f.Resp.Content()
	if err != nil {
		f.AddError(err.Error())
	} else {
		err := json.Unmarshal(text, &response)
		if err != nil {
			f.AddError(err.Error())
		}
	}
	return response
}

// checkResponse checks if the payload returned by server has proper format
func checkResponse(f *frisby.Frisby, response Payload) {
	if response.Version != "1.0" {
		f.AddError("Improper 'version' attribute returned: " + response.Version)
	}

	rules := response.Rules
	if len(rules) != 1 {
		f.AddError("Expected just one rule to be returned")
		return
	}

	// just one rule is returned
	rule := rules[0]

	if len(rule.Conditions) != 1 {
		f.AddError("Expected just one condition to be returned")
	}
}

// checkRulesEndpoint tests if rules endpoint is reachable and that it returns expected data
func checkRulesEndpoint() {
	f := frisby.Create("Check the /gathering-rules endpoint").Get(apiURL + "gathering_rules")
	f.Send()
	f.ExpectStatus(200)
	f.ExpectHeader(contentTypeHeader, "application/json")

	payload := readGatheringRules(f)
	checkResponse(f, payload)

	f.PrintReport()
}
