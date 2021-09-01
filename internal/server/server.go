package server

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
)

type Config struct {
	Address    string `mapstructure:"address" toml:"address"`
	UseHTTPS   bool   `mapstructure:"use_https" toml:"use_https"`
	EnableCORS bool   `mapstructure:"enable_cors" toml:"enable_cors"`
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
	// log.Info().Msgf("Starting HTTP server at '%s'", addr)

	if server.Config.EnableCORS {
		server.Router.Use(CORSMiddleware())
	}

	server.HTTPServer = &http.Server{
		Addr:    addr,
		Handler: server.Router,
	}

	if server.Config.UseHTTPS {
		err = server.HTTPServer.ListenAndServeTLS("server.crt", "server.key")
	} else {
		err = server.HTTPServer.ListenAndServe()
	}

	if err != nil && err != http.ErrServerClosed {
		// log.Error().Err(err).Msg("Unable to start HTTP/S server")
		return err
	}

	return nil
}

// Stop method stops server's execution.
func (server *Server) Stop(ctx context.Context) error {
	return server.HTTPServer.Shutdown(ctx)
}
