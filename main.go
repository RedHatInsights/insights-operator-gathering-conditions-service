<<<<<<< HEAD:main.go
// Copyright 2022 Red Hat, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
=======
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
>>>>>>> main:cmd/server/main.go

package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RedHatInsights/insights-operator-utils/logger"
	"github.com/gorilla/mux"
	"github.com/redhatinsights/insights-operator-conditional-gathering/internal/cli"
	"github.com/redhatinsights/insights-operator-conditional-gathering/internal/config"
	"github.com/redhatinsights/insights-operator-conditional-gathering/internal/server"
	"github.com/redhatinsights/insights-operator-conditional-gathering/internal/service"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

const (
	defaultConfigFile = "config/config"
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

func parseFlags() (cliFlags cli.CliFlags) {
	flag.BoolVar(&cliFlags.ShowConfiguration, "show-configuration", false, "show configuration")
	flag.BoolVar(&cliFlags.ShowAuthors, "show-authors", false, "show authors")
	flag.BoolVar(&cliFlags.ShowVersion, "show-version", false, "show version")

	flag.Parse()
	return
}

func doSelectedOperation(cliFlags cli.CliFlags) {
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

func runServer() {
	var httpServer *server.Server

	serverConfig := config.ServerConfig()
	storageConfig := config.StorageConfig()

	// Logger
	err := logger.InitZerolog(
		config.LoggingConfig(),
		config.CloudWatchConfig(),
		config.SentryLoggingConfig(),
		config.KafkaZerologConfig(),
	)
	if err != nil {
		log.Error().Err(err).Msg("Logger could not be initialized")
		os.Exit(1)
	}

	// Storage
	if _, err = os.Stat(storageConfig.RulesPath); err != nil {
		log.Error().Err(err).Msg("Storage data path not found")
		os.Exit(1)
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
