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

package server

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Address    string `mapstructure:"address" toml:"address"`
	UseHTTPS   bool   `mapstructure:"use_https" toml:"use_https"`
	EnableCORS bool   `mapstructure:"enable_cors" toml:"enable_cors"`
	CertFolder string // added for testing purposes
}

type Server struct {
	Config     Config
	Router     *mux.Router
	HTTPServer *http.Server
}

func New(cfg Config, router *mux.Router) *Server {
	return &Server{
		Config: cfg,
		Router: router,
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

	server.HTTPServer = &http.Server{
		Addr:    addr,
		Handler: server.Router,
	}

	if server.Config.UseHTTPS {
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
