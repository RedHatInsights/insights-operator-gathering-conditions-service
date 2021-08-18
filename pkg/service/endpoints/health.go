package endpoints

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-kit/kit/endpoint"
)

type HealthRequest struct{}
type HealthResponse struct {
	Status string
}

func MakeHealthEndpoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		return nil, nil
	}
}

func DecodeHealthRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return nil, nil
}

func EncodeHealthResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	resp, ok := response.(*HealthResponse)
	if !ok {
		return ErrEncoding
	}

	return json.NewEncoder(w).Encode(resp)
}
