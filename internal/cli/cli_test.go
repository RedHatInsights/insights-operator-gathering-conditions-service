/*
Copyright © 2021, 2022, 2023 Red Hat, Inc.

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

package cli_test

import (
	"testing"

	"github.com/RedHatInsights/insights-operator-gathering-conditions-service/internal/cli"
	"github.com/RedHatInsights/insights-operator-gathering-conditions-service/internal/config"
	"github.com/RedHatInsights/insights-operator-gathering-conditions-service/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/tisnik/go-capture"
)

func TestPrintConfiguration(t *testing.T) {
	testConfig := config.Configuration{
		ServerConfig: server.Config{
			Address: "test_address",
		},
	}

	output, err := capture.StandardOutput(func() {
		cli.PrintConfiguration(&testConfig)
	})
	assert.NoError(t, err)
	assert.Contains(t, output, "\"Address\": \"test_address\",")
}

func TestPrintAuthors(t *testing.T) {
	output, err := capture.StandardOutput(func() {
		cli.PrintAuthors()
	})
	assert.NoError(t, err)
	assert.Contains(t, output, "Red Hat Inc.")
}

func TestPrintVersion(t *testing.T) {
	output, err := capture.StandardOutput(func() {
		cli.PrintVersionInfo()
	})
	assert.NoError(t, err)

	assert.Contains(t, output, "Version:\t*not set")
	assert.Contains(t, output, "Build time:\t*not set")
	assert.Contains(t, output, "Branch:\t*not set")
	assert.Contains(t, output, "Commit:\t*not set")
}
