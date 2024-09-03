package main_test

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"testing"
	"time"

	main "github.com/RedHatInsights/insights-operator-gathering-conditions-service"
	"github.com/RedHatInsights/insights-operator-gathering-conditions-service/internal/config"
	"github.com/stretchr/testify/assert"
)

const totalRequests = 10000

var wantResponses = []int{200, 404, 400}

func BenchmarkMyHTTPServer(b *testing.B) {
	err := config.LoadConfiguration("tests/config")
	if err != nil {
		b.Fatal(err)
	}

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
		if err != nil {
			b.Fatalf("Server encountered an error: %v", err)
		}
	default:
		// No errors from the server
	}

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
			if err != nil {
				b.Fatalf("Failed to make request: %v", err)
			}
			assert.Contains(b, wantResponses, resp.StatusCode)
		}
	}
	b.StopTimer()

	// Programmatically send the interrupt signal to trigger server shutdown
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		b.Fatalf("Failed to find process: %v", err)
	}
	err = p.Signal(os.Interrupt)
	if err != nil {
		b.Fatalf("Failed to send interrupt signal: %v", err)
	}

	// Allow time for the server to gracefully shut down
	time.Sleep(1 * time.Second)

	requestsPerSecond := float64(totalRequests) / b.Elapsed().Seconds()
	fmt.Printf("Requests per second: %f\n", requestsPerSecond)
}

func getRandomVersion() string {
	const maxVersion = 99
	return fmt.Sprintf(
		"%d.%d.%d",
		rand.Intn(maxVersion),
		rand.Intn(maxVersion),
		rand.Intn(maxVersion))
}
