package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-kit/kit/endpoint"
)

var ErrEncoding = errors.New("encoding error")
var ErrInvalidArgument = errors.New("invalid argument")
var ErrNotFound = errors.New("not found")

type HealthRequest struct{}
type HealthResponse struct {
	Status string
}

func makeHealthEndpoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		return nil, nil
	}
}

func decodeHealthRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return nil, nil
}

func encodeHealthResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	resp, ok := response.(*HealthResponse)
	if !ok {
		return ErrEncoding
	}

	return json.NewEncoder(w).Encode(resp)
}
