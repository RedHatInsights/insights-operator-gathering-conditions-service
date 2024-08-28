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
	"fmt"

	"github.com/gorilla/mux"
)

// APIPrefix is the prefix used in the console-dot environment
const APIPrefix = "/api/gathering"

const (
	// V1Prefix is the version prefix used in the console-dot environment
	V1Prefix = "/v1"
	// V2Prefix is the version prefix used for the newer API v2
	V2Prefix = "/v2"
)

// Handler structure represents HTTP request handler.
type Handler struct {
	svc RulesProvider
}

// NewHandler function constructs new HTTP request handler.
func NewHandler(svc RulesProvider) *Handler {
	return &Handler{
		svc: svc,
	}
}

// Register function registers new handler for given endpoint URL.
func (s *Handler) Register(r *mux.Router) {
	r.HandleFunc(APIPrefix+"/openapi.json", serveOpenAPI).Methods("GET")
	r.Handle(APIPrefix+"/gathering_rules", gatheringRulesEndpoint(s.svc)).Methods("GET")
	r.HandleFunc(APIPrefix+V1Prefix+"/openapi.json", serveOpenAPI).Methods("GET")
	r.Handle(APIPrefix+V1Prefix+"/gathering_rules", gatheringRulesEndpoint(s.svc)).Methods("GET")
	r.HandleFunc(APIPrefix+V2Prefix+"/openapi.json", serveOpenAPI).Methods("GET")

	v2Path := fmt.Sprintf("%s%s/{ocpVersion}/gathering_rules", APIPrefix, V2Prefix)
	r.Handle(v2Path, remoteConfigurationEndpoint(s.svc)).Methods("GET")
}
