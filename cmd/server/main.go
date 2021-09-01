package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/redhatinsights/insights-operator-conditional-gathering/internal/config"
	"github.com/redhatinsights/insights-operator-conditional-gathering/internal/server"
	"github.com/redhatinsights/insights-operator-conditional-gathering/internal/service"
	"golang.org/x/sync/errgroup"
)

const (
	defaultConfigFile = "config"
)

func main() {
	var httpServer *server.Server

	// Load config
	err := config.LoadConfiguration(defaultConfigFile)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v", config.Config)

	serverConfig := config.ServerConfig()
	serviceConfig := config.ServiceConfig()

	// Logger
	var logger log.Logger
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)

	// Repository
	if _, err = os.Stat(serviceConfig.RulesPath); err != nil {
		logger.Log("msg", "repository data path error", err) // nolint: errcheck
		os.Exit(1)
	}
	repo := service.NewRepository(serviceConfig.RulesPath)
	svc := service.New(repo)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupt)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	// HTTP
	g.Go(func() error {
		logger.Log("transport", "http", "address", serverConfig.Address, "msg", "listening") // nolint: errcheck

		router := mux.NewRouter().StrictSlash(true)

		// Register the service
		service.NewHandler(svc).Register(router)

		// Create the HTTP Server
		httpServer = server.New(serverConfig, router)

		err = httpServer.Start()
		if err != nil {
			return err
		}

		return nil
	})

	select {
	case <-interrupt:
		break
	case <-ctx.Done():
		break
	}

	logger.Log("msg", "received shutdown signal") // nolint: errcheck

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if httpServer != nil {
		httpServer.Stop(shutdownCtx) // nolint: errcheck
	}

	err = g.Wait()
	if err != nil {
		logger.Log("msg", "server returning an error", "error", err) // nolint: errcheck
		defer os.Exit(2)
	}

	logger.Log("msg", "server closed") // nolint: errcheck
}
