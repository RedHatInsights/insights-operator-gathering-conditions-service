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
