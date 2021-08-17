package conditional

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-kit/kit/log"
	// "github.com/go-kit/kit/transport"
	// kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

func NewHandler(svc ServiceInterface, logger log.Logger) http.Handler {
	// opts := []kithttp.ServerOption{
	// 	kithttp.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
	// 	kithttp.ServerErrorEncoder(encodeError),
	// }

	// healthHandler := kithttp.NewServer(
	// 	makeHealthEndpoint(),
	// 	decodeHealthRequest,
	// 	encodeHealthResponse,
	// 	opts...,
	// )

	r := mux.NewRouter()

	// r.Handle("/health", healthHandler).Methods(http.MethodGet)

	return r
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch err {
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}

	// nolint: errcheck
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
