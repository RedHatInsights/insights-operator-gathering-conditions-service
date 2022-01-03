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

package cli_test

import (
	"testing"

	"github.com/redhatinsights/insights-operator-conditional-gathering/internal/cli"
	"github.com/redhatinsights/insights-operator-conditional-gathering/internal/config"
	"github.com/redhatinsights/insights-operator-conditional-gathering/internal/server"
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
		err := cli.PrintConfiguration(testConfig)
		assert.NoError(t, err)
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
