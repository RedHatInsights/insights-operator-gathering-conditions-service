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

package service_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/RedHatInsights/insights-operator-gathering-conditions-service/internal/service"
)

// mockedWriter is simple httpWriter implementation that fails on Write method.
type mockedWriter struct {
	Code int
}

func (mw mockedWriter) Header() http.Header {
	return make(http.Header)
}

func (mw mockedWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write error")
}

func (mw mockedWriter) WriteString(_ string) (int, error) {
	return 0, nil
}

func (mw *mockedWriter) WriteHeader(code int) {
	mw.Code = code
}

func TestRenderResponseWriteError(t *testing.T) {
	writer := mockedWriter{}

	var resp interface{} = "foobar"

	// this should fail
	service.RenderResponse(&writer, resp, 200)

	// we expect that error should be returned
	assert.Equal(t, http.StatusInternalServerError, writer.Code)
}

func TestRenderResponseJSONMarshalError(t *testing.T) {
	writer := httptest.NewRecorder()

	// this is ugly trick - complex numbers are not serializable to JSON
	var resp interface{} = 3 + 2i
	service.RenderResponse(writer, resp, 0)

	// we expect that error should be returned
	assert.Equal(t, http.StatusInternalServerError, writer.Code)
}

func TestLogHeaders(t *testing.T) {
	req, err := http.NewRequest("GET", "http://example.com", nil)
	assert.NoError(t, err)

	req.Header.Add("Authorization", "Bearer token")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", "Go-http-client/1.1")

	t.Run("no filter", func(t *testing.T) {
		logs := &logSink{}
		logger := zerolog.New(logs)
		logEvent := logger.Debug()

		service.LogHeaders(req, []string{}, logEvent)
		logEvent.Msg("test")

		assert.Len(t, logs.logs, 1, "received more than 1 log")
		got := logs.logs[0]
		assert.Contains(t, got, `"Authorization":["Bearer token"]`)
		assert.Contains(t, got, `"User-Agent":["Go-http-client/1.1"]`)
		assert.Contains(t, got, `"Content-Type":["application/json"]`)
	})

	t.Run("with filter", func(t *testing.T) {
		logs := &logSink{}
		logger := zerolog.New(logs)
		logEvent := logger.Debug()

		service.LogHeaders(req, []string{"Authorization"}, logEvent)
		logEvent.Msg("test")

		assert.Len(t, logs.logs, 1, "received more than 1 log")
		got := logs.logs[0]
		// Note that "Authorization":["Bearer token"] is no longer expected
		assert.Contains(t, got, `"User-Agent":["Go-http-client/1.1"]`)
		assert.Contains(t, got, `"Content-Type":["application/json"]`)
	})
}

type logSink struct {
	logs []string
}

func (l *logSink) Write(p []byte) (n int, err error) {
	l.logs = append(l.logs, string(p))
	return len(p), nil
}

func (l *logSink) Index(i int) string {
	return l.logs[i]
}
