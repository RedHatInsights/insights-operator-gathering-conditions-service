package server

import (
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func CORSMiddleware() mux.MiddlewareFunc {
	headersOK := handlers.AllowedHeaders([]string{
		"Content-Type",
		"Content-Length",
		"Accept-Encoding",
		"X-CSRF-Token",
		"Authorization",
	})
	originsOK := handlers.AllowedOrigins([]string{"*"})
	methodsOK := handlers.AllowedMethods([]string{
		http.MethodPost,
		http.MethodGet,
		http.MethodOptions,
		http.MethodPut,
		http.MethodDelete,
	})
	credsOK := handlers.AllowCredentials()

	return handlers.CORS(originsOK, headersOK, methodsOK, credsOK)
}
