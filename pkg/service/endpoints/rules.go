package endpoints

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	"github.com/openshift/insights-operator-conditional-gathering/pkg/service"
)

type RulesResponse struct {
	Version string      `json:"version"`
	Rules   interface{} `json:"rules"`
}

func MakeRulesEndpoint(svc service.Interface) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		rules, err := svc.Rules()
		if err != nil {
			return nil, err
		}

		return &RulesResponse{
			Version: "1.0",
			Rules:   rules.Items,
		}, nil
	}
}

func DecodeRulesRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return nil, nil
}

func EncodeRulesResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	resp, ok := response.(*RulesResponse)
	if !ok {
		return ErrEncoding
	}

	return json.NewEncoder(w).Encode(resp)
}
