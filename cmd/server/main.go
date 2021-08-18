package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/kit/log"
	"github.com/openshift/insights-operator-conditional-gathering/pkg/service"
	"github.com/openshift/insights-operator-conditional-gathering/pkg/service/transport"
)

const (
	httpAddr = ":8080"
)

func main() {
	// Logger
	var logger log.Logger
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)

	// Services
	svc := service.New()

	httplogger := log.With(logger, "component", "http")

	mux := http.NewServeMux()
	// Control CORs and other stuff
	http.Handle("/", accessControl(mux))

	// Handlers
	mux.Handle("/", transport.NewHTTPHandler(svc, httplogger))

	// HTTP
	errs := make(chan error, 2)

	go func() {
		logger.Log("transport", "http", "address", httpAddr, "msg", "listening") // nolint: errcheck
		errs <- http.ListenAndServe(httpAddr, nil)
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	logger.Log("terminated", <-errs) // nolint: errcheck
}

func accessControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")

		if r.Method == "OPTIONS" {
			return
		}

		h.ServeHTTP(w, r)
	})
}
