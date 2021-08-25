package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/redhatinsights/insights-operator-conditional-gathering/pkg/service"
	"github.com/redhatinsights/insights-operator-conditional-gathering/pkg/service/transport"
	"golang.org/x/sync/errgroup"
)

func main() {
	var (
		// Flags
		httpAddr  = flag.String("httpAddr", ":8080", "http listen address")
		rulesPath = flag.String("rulesPath", "./rules", "the path where to find the rules files")

		httpServer *http.Server
	)

	flag.Parse()

	// Logger
	var logger log.Logger
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Repository
	if _, err := os.Stat(*rulesPath); os.IsNotExist(err) {
		logger.Log("msg", "repository data path not exists", err) // nolint: errcheck
		os.Exit(1)
	}
	repo := service.NewRepository(*rulesPath)

	// Services
	svc := service.New(repo)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupt)

	g, ctx := errgroup.WithContext(ctx)

	// HTTP
	g.Go(func() error {
		httplogger := log.With(logger, "component", "http")

		mux := http.NewServeMux()
		// Access Control and CORS
		http.Handle("/", accessControl(mux))

		// Handlers
		mux.Handle("/", transport.NewHTTPHandler(svc, httplogger))

		logger.Log("transport", "http", "address", *httpAddr, "msg", "listening") // nolint: errcheck

		httpServer = &http.Server{
			Addr:         *httpAddr,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		}

		err := httpServer.ListenAndServe()
		if err != http.ErrServerClosed {
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
		httpServer.Shutdown(shutdownCtx) // nolint: errcheck
	}

	err := g.Wait()
	if err != nil {
		logger.Log("msg", "server returning an error", "error", err)
		os.Exit(2)
	}

	logger.Log("msg", "server successfully closed")
}

func accessControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Response type
		w.Header().Set("Content-Type", "application/json")
		// CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")

		if r.Method == "OPTIONS" {
			return
		}

		h.ServeHTTP(w, r)
	})
}
