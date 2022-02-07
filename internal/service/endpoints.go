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

package service

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rs/zerolog/log"

	merrors "github.com/RedHatInsights/insights-operator-gathering-conditions-service/internal/errors"
)

// ErrorResponse structure represents HTTP response with error message.
type ErrorResponse struct {
	Error string `json:"error"`
}

// GatheringRulesResponse structure represents HTTP response with rules-related
// content.
type GatheringRulesResponse struct {
	Version string      `json:"version"`
	Rules   interface{} `json:"rules"`
}

func gatheringRulesEndpoint(svc Interface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rules, err := svc.Rules()
		if err != nil {
			renderErrorResponse(w, "internal error", err)
			return
		}

		log.Info().Int("rules count", len(rules.Items)).Msg("Serving gathering rules")
		renderResponse(w, &GatheringRulesResponse{
			Version: rules.Version,
			Rules:   rules.Items,
		}, http.StatusOK)
	}
}

func renderResponse(w http.ResponseWriter, resp interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")

	content, err := json.Marshal(resp)
	if err != nil {
		log.Error().Err(err).Msg("Unable to marshal response data to JSON")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(code)

	if _, err = w.Write(content); err != nil {
		log.Error().Err(err).Msg("Unable to write response data")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Info().Msg("Response has been sent")
}

func renderErrorResponse(w http.ResponseWriter, msg string, err error) {
	resp := ErrorResponse{Error: msg}
	code := http.StatusInternalServerError

	log.Error().Msgf("%v", err)

	var ierr *merrors.Error
	if !errors.As(err, &ierr) {
		resp.Error = "internal error"
	} else {
		// TODO: These codes are never used, shall we remove them?
		switch ierr.Code() {
		case merrors.ErrorCodeNotFound:
			code = http.StatusNotFound
		case merrors.ErrorCodeInvalidArgument:
			code = http.StatusBadRequest
		case merrors.ErrorCodeUnknown:
			code = http.StatusInternalServerError
		}
	}

	renderResponse(w, resp, code)
}
