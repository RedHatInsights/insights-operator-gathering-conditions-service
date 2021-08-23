package transport

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/redhatinsights/insights-operator-conditional-gathering/pkg/service"
	"github.com/redhatinsights/insights-operator-conditional-gathering/pkg/service/endpoints"
)

func NewHTTPHandler(svc service.Interface, logger log.Logger) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		kithttp.ServerErrorEncoder(encodeError),
	}

	gatheringRulesHandler := kithttp.NewServer(
		endpoints.MakeGatheringRulesEndpoint(svc),
		endpoints.DecodeGatheringRulesRequest,
		endpoints.EncodeGatheringRulesResponse,
		opts...,
	)

	r := mux.NewRouter()

	r.Handle("/gathering_rules", gatheringRulesHandler).Methods(http.MethodGet)

	return r
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch err {
	case endpoints.ErrNotFound:
		w.WriteHeader(http.StatusNotFound)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}

	// nolint: errcheck
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
