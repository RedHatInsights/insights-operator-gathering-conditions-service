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

// Package server exposes the service as a REST API
package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

	"github.com/RedHatInsights/insights-operator-gathering-conditions-service/internal/errors"
)

const (
	openAPIURL = "/openapi.json"
)

// Config data structure represents HTTP/HTTPS server configuration.
type Config struct {
	Address    string `mapstructure:"address" toml:"address"`
	UseHTTPS   bool   `mapstructure:"use_https" toml:"use_https"`
	EnableCORS bool   `mapstructure:"enable_cors" toml:"enable_cors"`
	CertFolder string // added for testing purposes
}

// AuthConfig structure represents auth. settings for the server
type AuthConfig struct {
	Enabled bool   `mapstructure:"enabled" toml:"enabled"`
	Type    string `mapstructure:"type" toml:"type"`
}

// Server data structure represents instances of HTTP/HTTPS server.
type Server struct {
	Config     Config
	AuthConfig AuthConfig
	Router     *mux.Router
	HTTPServer *http.Server
}

// New function constructs new server instance.
func New(cfg Config, authCfg AuthConfig, router *mux.Router) *Server {
	return &Server{
		Config:     cfg,
		AuthConfig: authCfg,
		Router:     router,
	}
}

// Start method starts HTTP or HTTPS server.
func (server *Server) Start() error {
	var err error

	addr := server.Config.Address
	log.Info().Msgf("Starting HTTP server at '%s'", addr)

	if server.Config.EnableCORS {
		server.Router.Use(CORSMiddleware())
	}

	if server.AuthConfig.Enabled {
		log.Info().Str("type", server.AuthConfig.Type).Msg("Enabling auth")
		// we have to enable authentication for all endpoints, including endpoints
		// for Prometheus metrics and OpenAPI specification, because there is not
		// single prefix of other REST API calls. The special endpoints needs to
		// be handled in middleware which is not optimal
		noAuthURLs := []string{
			openAPIURL,
			openAPIURL + "?", // to be able to test using Frisby
		}
		server.Router.Use(func(next http.Handler) http.Handler {
			return server.Authentication(next, noAuthURLs)
		})
	} else {
		log.Info().Msg("Auth disabled")
	}

	server.HTTPServer = &http.Server{
		Addr:              addr,
		Handler:           server.Router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	if server.Config.UseHTTPS {
		log.Info().
			Str("cert.folder", server.Config.CertFolder).
			Msg("Using TLS")
		err = server.HTTPServer.ListenAndServeTLS(
			server.Config.CertFolder+"server.crt",
			server.Config.CertFolder+"server.key")
	} else {
		err = server.HTTPServer.ListenAndServe()
	}

	if err != nil && err != http.ErrServerClosed {
		log.Error().Err(err).Msg("Unable to start HTTP/S server")
		return err
	}

	return nil
}

// Stop method stops server's execution.
func (server *Server) Stop(ctx context.Context) error {
	return server.HTTPServer.Shutdown(ctx)
}

// HandleServerError handles separate server errors and sends appropriate responses
func HandleServerError(writer http.ResponseWriter, err error) {
	logFunc := log.Error

	var respErr error

	switch err := err.(type) {
	case *errors.RouterMissingParamError, *errors.RouterParsingError, *errors.NoBodyError, *errors.ValidationError:
		respErr = SendBadRequest(writer, err.Error())
		logFunc = log.Warn
	case *errors.NotFoundError:
		respErr = SendNotFound(writer, err.Error())
	case *json.UnmarshalTypeError:
		respErr = SendBadRequest(writer, "bad type in json data")
	case *errors.UnauthorizedError:
		respErr = SendUnauthorized(writer, err.Error())
	case *errors.ForbiddenError:
		respErr = SendForbidden(writer, err.Error())
	default:
		respErr = SendInternalServerError(writer, "Internal Server Error")
	}

	logFunc().Type("errType", err).Err(err).Msg("handleServerError()")

	if respErr != nil {
		log.Error().Err(respErr).Msg(errors.ResponseDataError)
	}
}
