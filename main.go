/*
Copyright © 2021, 2022, 2023, 2024 Red Hat, Inc.

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
	cliFlags := ParseFlags()
	exitCode := doSelectedOperation(cliFlags)
	os.Exit(exitCode)
}

// ParseFlags returns a variable cli.Flags with the input parameters
func ParseFlags() (cliFlags cli.Flags) {
	flag.BoolVar(&cliFlags.ShowConfiguration, "show-configuration", false, "show configuration")
	flag.BoolVar(&cliFlags.ShowAuthors, "show-authors", false, "show authors")
	flag.BoolVar(&cliFlags.ShowVersion, "show-version", false, "show version")
	flag.BoolVar(&cliFlags.CheckConfig, "check-config", false, "initialize the service in order to check all the configuration is right")

	flag.Parse()
	return
}

func doSelectedOperation(cliFlags cli.Flags) int {
	switch {
	case cliFlags.ShowConfiguration:
		cli.PrintConfiguration(&config.Config)
	case cliFlags.ShowAuthors:
		cli.PrintAuthors()
	case cliFlags.ShowVersion:
		cli.PrintVersionInfo()
	case cliFlags.CheckConfig:
		_, err := InitService()
		if err != nil {
			return 1
		}
	default:
		err := RunServer()
		if err != nil {
			return 1
		}
	}
	return 0
}

// InitService creates a new *service.Service after parsing all the storage
// configuration. It's a good way of checking all the inputs are right
func InitService() (*service.Service, error) {
	storageConfig := config.StorageConfig()
	// Logger
	err := initLogger()
	if err != nil {
		log.Error().Err(err).Msg("Logger could not be initialized")
		return nil, err
	}
	defer logger.CloseZerolog()

	// Storage
	if _, err = os.Stat(storageConfig.RulesPath); err != nil {
		logStorageError(err, storageConfig.RulesPath)
		return nil, err
	}
	var unleashClient *service.UnleashClient
	if storageConfig.UnleashEnabled {
		unleashClient, err = service.NewUnleashClient(storageConfig)
		if err != nil {
			log.Error().Err(err).Msg("Unleash could not be initialized")
			return nil, err
		}
	}
	store, err := service.NewStorage(storageConfig, unleashClient)
	if err != nil {
		log.Error().Err(err).Msg("Error initializing the storage")
		return nil, err
	}

	// Repository & Service
	repo := service.NewRepository(store)
	return service.New(repo), nil
}

// RunServer starts the server with the loaded configuration. Make sure to call
// config.LoadConfiguration in order not to start with an empty configuration.
func RunServer() error {
	var httpServer *server.Server

	serverConfig := config.ServerConfig()
	authConfig := config.AuthConfig()

	svc, err := InitService()
	if err != nil {
		log.Error().Err(err).Msg("Error occurred during service initialization")
		return err
	}

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

		// Register the service
		service.NewHandler(svc).Register(router)

		// Create the HTTP Server
		httpServer = server.New(serverConfig, authConfig, router)

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

	stopHTTPServer(shutdownCtx, g, httpServer)

	log.Info().Msg("Server closed")
	return nil
}

// stopHTTPServer function initialize the HTTP server shutdown operations
func stopHTTPServer(shutdownCtx context.Context, g *errgroup.Group, httpServer *server.Server) {
	if httpServer != nil {
		err := httpServer.Stop(shutdownCtx)
		if err != nil {
			log.Error().Err(err).Msg("HTTP(s) server Shutdown operation failure")
		}
	}

	err := g.Wait()
	if err != nil {
		log.Error().Err(err).Msg("Server returning an error")
		defer os.Exit(2)
	}
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

func logStorageError(err error, path string) {
	log.Error().
		Err(err).
		Str("rulesPath", path).
		Msg("Storage data path not found")
}
