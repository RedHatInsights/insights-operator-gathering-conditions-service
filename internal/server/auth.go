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

package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/rs/zerolog/log"

	"github.com/RedHatInsights/insights-operator-gathering-conditions-service/internal/collections"
	"github.com/RedHatInsights/insights-operator-gathering-conditions-service/internal/errors"
)

const (
	// #nosec G101
	malformedTokenMessage = "Malformed authentication token"
	invalidTokenMessage   = "Invalid/Malformed auth token"
)

// ContextKey is a type for user authentication token in request
type ContextKey string

// ContextKeyUser is a constant for user authentication token in request
const ContextKeyUser = ContextKey("user")

// OrgID represents organization ID
type OrgID uint32

// UserID represents type for user id
type UserID string

// Internal contains information about organization ID
type Internal struct {
	OrgID OrgID `json:"org_id,string"`
}

// Identity contains internal user info
type Identity struct {
	AccountNumber UserID   `json:"account_number"`
	Internal      Internal `json:"internal"`
}

// Token is x-rh-identity struct
type Token struct {
	Identity Identity `json:"identity"`
}

// JWTPayload is structure that contain data from parsed JWT token
type JWTPayload struct {
	AccountNumber UserID `json:"account_number"`
	OrgID         OrgID  `json:"org_id,string"`
}

// Authentication middleware for checking auth rights
func (server *Server) Authentication(next http.Handler, noAuthURLs []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// for specific URLs it is ok to not use auth. mechanisms at all
		// this is specific to OpenAPI JSON response and for all OPTION HTTP methods
		if collections.StringInSlice(r.RequestURI, noAuthURLs) || r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		// try to read auth. header from HTTP request (if provided by client)
		token, err := server.getAuthTokenHeader(w, r)
		if err != nil {
			log.Error().Err(err).Msg(err.Error())
			HandleServerError(w, err)
			return
		}

		log.Info().Msgf("Authentication token: %s", token)

		var decoded []byte

		// decode auth. token to JSON string
		if server.AuthConfig.Type == "jwt" {
			decoded, err = jwt.DecodeSegment(token)
		} else {
			decoded, err = base64.StdEncoding.DecodeString(token)
		}

		// if token is malformed return HTTP code 403 to client
		if err != nil {
			// malformed token, returns with http code 403 as usual
			log.Error().Err(err).Msg(malformedTokenMessage)
			HandleServerError(w, &errors.UnauthorizedError{ErrString: malformedTokenMessage})
			return
		}

		tk := &Token{}

		// if we took JWT token, it has different structure then x-rh-identity
		if server.AuthConfig.Type == "jwt" {
			jwtPayload := &JWTPayload{}
			err = json.Unmarshal([]byte(decoded), jwtPayload)
			if err != nil {
				// Malformed token, returns with http code 403 as usual
				log.Error().Err(err).Msg(malformedTokenMessage)
				HandleServerError(w, &errors.UnauthorizedError{ErrString: malformedTokenMessage})
				return
			}
			// Map JWT token to inner token
			tk.Identity = Identity{
				AccountNumber: jwtPayload.AccountNumber,
				Internal: Internal{
					OrgID: jwtPayload.OrgID,
				},
			}
		} else {
			err = json.Unmarshal([]byte(decoded), tk)

			if err != nil {
				// malformed token, returns with HTTP code 403 as usual
				log.Error().Err(err).Msg(malformedTokenMessage)
				HandleServerError(w, &errors.UnauthorizedError{ErrString: malformedTokenMessage})
				return
			}
		}

		// Everything went well, proceed with the request and set the
		// caller to the user retrieved from the parsed token
		ctx := context.WithValue(r.Context(), ContextKeyUser, tk.Identity)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// GetCurrentUserID retrieves current user's id from request
func (server *Server) GetCurrentUserID(request *http.Request) (UserID, error) {
	i := request.Context().Value(ContextKeyUser)

	if i == nil {
		log.Error().Msgf("user id was not found in request's context")
		return "", &errors.UnauthorizedError{ErrString: "user id is not provided"}
	}

	identity, ok := i.(Identity)
	if !ok {
		return "", &errors.AuthenticationError{ErrString: "contextKeyUser has wrong type"}
	}

	return identity.AccountNumber, nil
}

// GetAuthToken returns current authentication token
func (server *Server) GetAuthToken(request *http.Request) (*Identity, error) {
	i := request.Context().Value(ContextKeyUser)

	if i == nil {
		return nil, &errors.AuthenticationError{ErrString: "token is not provided"}
	}

	identity, ok := i.(Identity)
	if !ok {
		return nil, &errors.AuthenticationError{ErrString: "contextKeyUser has wrong type"}
	}

	return &identity, nil
}

func (server *Server) getAuthTokenHeader(w http.ResponseWriter, r *http.Request) (string, error) {
	var tokenHeader string
	// In case of testing on local machine we don't take x-rh-identity
	// header, but instead Authorization with JWT token in it
	if server.AuthConfig.Type == "jwt" {
		log.Info().Msg("Retrieving jwt token")

		// Grab the token from the header
		tokenHeader = r.Header.Get("Authorization")

		// The token normally comes in format `Bearer {token-body}`, we
		// check if the retrieved token matched this requirement
		splitted := strings.Split(tokenHeader, " ")
		if len(splitted) != 2 {
			return "", &errors.UnauthorizedError{ErrString: invalidTokenMessage}
		}

		// Here we take JWT token which include 3 parts, we need only
		// second one
		splitted = strings.Split(splitted[1], ".")
		if len(splitted) < 1 {
			return "", &errors.UnauthorizedError{ErrString: invalidTokenMessage}
		}
		tokenHeader = splitted[1]
	} else {
		log.Info().Msg("Retrieving x-rh-identity token")
		// Grab the token from the header
		tokenHeader = r.Header.Get("x-rh-identity")
	}

	log.Info().Int("Length", len(tokenHeader)).Msg("Token retrieved")

	if tokenHeader == "" {
		const message = "Missing auth token"
		return "", &errors.UnauthorizedError{ErrString: message}
	}

	return tokenHeader, nil
}
