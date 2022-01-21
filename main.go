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

package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RedHatInsights/insights-operator-gathering-conditions-service/internal/cli"
	"github.com/RedHatInsights/insights-operator-gathering-conditions-service/internal/config"
	"github.com/RedHatInsights/insights-operator-gathering-conditions-service/internal/server"
	"github.com/RedHatInsights/insights-operator-gathering-conditions-service/internal/service"
	"github.com/RedHatInsights/insights-operator-utils/logger"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

const (
	defaultConfigFile = "config"
)

// main function perform operation based on the flags defined on command line
func main() {
	// Load config
	err := config.LoadConfiguration(defaultConfigFile)
	if err != nil {
		log.Error().Err(err).Msg("Configuration could not be loaded")
		os.Exit(1)
	}
	cliFlags := parseFlags()
	doSelectedOperation(cliFlags)
}

func parseFlags() (cliFlags cli.Flags) {
	flag.BoolVar(&cliFlags.ShowConfiguration, "show-configuration", false, "show configuration")
	flag.BoolVar(&cliFlags.ShowAuthors, "show-authors", false, "show authors")
	flag.BoolVar(&cliFlags.ShowVersion, "show-version", false, "show version")

	flag.Parse()
	return
}

func doSelectedOperation(cliFlags cli.Flags) {
	switch {
	case cliFlags.ShowConfiguration:
		cli.PrintConfiguration(config.Config)
	case cliFlags.ShowAuthors:
		cli.PrintAuthors()
	case cliFlags.ShowVersion:
		cli.PrintVersionInfo()
	default:
		runServer()
	}
}

// serveOpenAPI function handles requests to get OpenAPI specification file
func serveOpenAPI(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Serving openapi.json file")
	p := "./openapi.json"
	w.Header().Set("Content-type", "application/json")
	http.ServeFile(w, r, p)
}

func runServer() {
	var httpServer *server.Server

	serverConfig := config.ServerConfig()
	storageConfig := config.StorageConfig()

	// Logger
	err := initLogger()
	if err != nil {
		log.Error().Err(err).Msg("Logger could not be initialized")
		return
	}
	defer logger.CloseZerolog()

	// Storage
	if _, err = os.Stat(storageConfig.RulesPath); err != nil {
		log.Error().
			Err(err).
			Str("rulesPath", storageConfig.RulesPath).
			Msg("Storage data path not found")
		return
	}
	store := service.NewStorage(storageConfig)

	// Repository & Service
	repo := service.NewRepository(store)
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
		router := mux.NewRouter().StrictSlash(true)

		// Handler for static contennt with OpenAPI specification
		router.HandleFunc("/openapi.json", serveOpenAPI)

		// Register the service
		service.NewHandler(svc).Register(router)

		// Create the HTTP Server
		httpServer = server.New(serverConfig, router)

		return httpServer.Start()
	})

	select {
	case <-interrupt:
		break
	case <-ctx.Done():
		break
	}

	log.Info().Msg("Received shutdown signal")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if httpServer != nil {
		err := httpServer.Stop(shutdownCtx)
		if err != nil {
			log.Error().Err(err).Msg("HTTP(s) server Shutdown operation failure")
		}
	}

	err = g.Wait()
	if err != nil {
		log.Error().Err(err).Msg("Server returning an error")
		defer os.Exit(2)
	}

	log.Info().Msg("Server closed")
}

// initLogger function initializes logger instance
func initLogger() error {
	return logger.InitZerolog(
		config.LoggingConfig(),
		config.CloudWatchConfig(),
		config.SentryLoggingConfig(),
		config.KafkaZerologConfig(),
	)
}
