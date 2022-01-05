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

	merrors "github.com/redhatinsights/insights-operator-conditional-gathering/internal/errors"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

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

		renderResponse(w, &GatheringRulesResponse{
			Version: "1.0",
			Rules:   rules.Items,
		}, http.StatusOK)
	}
}

func renderResponse(w http.ResponseWriter, resp interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")

	content, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(code)

	if _, err = w.Write(content); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
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
