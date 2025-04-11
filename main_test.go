package main_test

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"testing"
	"time"

	main "github.com/RedHatInsights/insights-operator-gathering-conditions-service"
	"github.com/RedHatInsights/insights-operator-gathering-conditions-service/internal/cli"
	"github.com/RedHatInsights/insights-operator-gathering-conditions-service/internal/config"
	"github.com/stretchr/testify/assert"
)

const totalRequests = 10000

var wantResponses = []int{200, 404, 400}

func TestMain(m *testing.M) {
	err := config.LoadConfiguration("tests/config")
	if err != nil {
		os.Exit(1)
	}
	m.Run()
}

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedFlags cli.Flags
	}{
		{
			name: "No flags",
			args: []string{},
			expectedFlags: cli.Flags{
				ShowConfiguration: false,
				ShowAuthors:       false,
				ShowVersion:       false,
				CheckConfig:       false,
			},
		},
		{
			name: "Show configuration flag",
			args: []string{"-show-configuration"},
			expectedFlags: cli.Flags{
				ShowConfiguration: true,
				ShowAuthors:       false,
				ShowVersion:       false,
				CheckConfig:       false,
			},
		},
		{
			name: "Show authors flag",
			args: []string{"-show-authors"},
			expectedFlags: cli.Flags{
				ShowConfiguration: false,
				ShowAuthors:       true,
				ShowVersion:       false,
				CheckConfig:       false,
			},
		},
		{
			name: "Multiple flags",
			args: []string{"-show-configuration", "-check-config"},
			expectedFlags: cli.Flags{
				ShowConfiguration: true,
				ShowAuthors:       false,
				ShowVersion:       false,
				CheckConfig:       true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset the command-line flags for each test case
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			// Set os.Args to simulate command-line arguments
			os.Args = append([]string{os.Args[0]}, tt.args...)

			// Parse the flags and get the result
			parsedFlags := main.ParseFlags()

			// Use testify to assert the expected and actual flags
			assert.Equal(t, tt.expectedFlags, parsedFlags)
		})
	}
}

func TestInitService(t *testing.T) {
	_, err := main.InitService()
	assert.NoError(t, err)
}

func TestRunService(t *testing.T) {
	startServer(t)

	// Create an HTTP client
	client := &http.Client{}

	version := getRandomVersion()

	resp, err := client.Get(
		fmt.Sprintf(
			"http://localhost%s/v2/%s/gathering_rules",
			config.ServerConfig().Address, version))

	assert.NoError(t, err)
	assert.Contains(t, wantResponses, resp.StatusCode)

	stopServer(t)
}

func BenchmarkMyHTTPServer(b *testing.B) {
	startServer(b)

	// Create an HTTP client
	client := &http.Client{}

	b.ResetTimer()
	b.StopTimer()
	// Benchmark loop
	for i := 0; i < b.N; i++ {
		for j := 0; j < totalRequests; j++ {
			version := getRandomVersion()

			b.StartTimer()
			resp, err := client.Get(
				fmt.Sprintf(
					"http://localhost%s/v2/%s/gathering_rules",
					config.ServerConfig().Address, version))
			b.StopTimer()

			assert.NoError(b, err)
			assert.Contains(b, wantResponses, resp.StatusCode)
		}
	}
	b.StopTimer()

	stopServer(b)

	requestsPerSecond := float64(totalRequests) / b.Elapsed().Seconds()
	fmt.Printf("Requests per second: %f\n", requestsPerSecond)
}

// #nosec G404
func getRandomVersion() string {
	const maxVersion = 99
	return fmt.Sprintf(
		"%d.%d.%d",
		rand.Intn(maxVersion),
		rand.Intn(maxVersion),
		rand.Intn(maxVersion))
}

func startServer(t assert.TestingT) {
	// Channel to report server errors
	serverErrChan := make(chan error, 1)

	// Start the server in a background goroutine
	go func() {
		// Run the server and send any errors to the channel
		err := main.RunServer()
		serverErrChan <- err
	}()

	// Give the server a moment to finish initialization
	time.Sleep(1 * time.Second)

	// Check if the server returned any error
	select {
	case err := <-serverErrChan:
		assert.NoError(t, err)
	default:
		// No errors from the server
	}
}

func stopServer(t assert.TestingT) {
	// Programmatically send the interrupt signal to trigger server shutdown
	p, err := os.FindProcess(os.Getpid())
	assert.NoError(t, err)

	err = p.Signal(os.Interrupt)
	assert.NoError(t, err)

	// Allow time for the server to gracefully shut down
	time.Sleep(1 * time.Second)
}
