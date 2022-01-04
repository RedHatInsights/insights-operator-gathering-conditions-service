package server_test

import (
	"context"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/redhatinsights/insights-operator-conditional-gathering/internal/server"
	"github.com/stretchr/testify/assert"
)

type testCase struct {
	name          string
	config        server.Config
	expectAnError bool
}

func TestServer(t *testing.T) {
	testCases := []testCase{
		{
			name: "without CORS",
			config: server.Config{
				Address: "localhost:1234",
			},
		},
		{
			name: "with CORS",
			config: server.Config{
				Address:    "localhost:1234",
				EnableCORS: true,
			},
		},
		{
			name: "with TLS",
			config: server.Config{
				Address:    "localhost:1234",
				UseHTTPS:   true,
				CertFolder: "testdata/",
			},
		},
		{
			name: "with TLS but returning an error",
			config: server.Config{
				Address:    "localhost:1234",
				UseHTTPS:   true,
				CertFolder: "not-a-folder/",
			},
			expectAnError: true,
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {
			testServer := server.New(tc.config, mux.NewRouter())
			go func() {
				err := testServer.Start()
				if tc.expectAnError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			}()
			time.Sleep(100 * time.Millisecond)
			err := testServer.Stop(context.TODO())
			assert.NoError(t, err)
		})
	}
}
