package main

import (
	"encoding/json"
	"net/http"

	"github.com/braswelljr/socki/ws"
	"github.com/gorilla/mux"
)

// NewRouter creates a new router.
func NewHTTPRouter(r *mux.Router) {

	// prefix all routes with /api
	api := r.PathPrefix("/api").Subrouter()

	// register websocket handler
	api.HandleFunc("/chatroom", ws.NewWebSocketServer)

	// set general headers
	api.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// set headers
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Content-Type-Options", "nosniff")

			// call the next handler in the chain
			next.ServeHTTP(w, r)
		})
	})

	// Add routes.
	api.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)

		// return a json response
		if err := json.NewEncoder(w).Encode(map[string]string{"message": "Welcome to the socki API"}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
		}
	}).Methods(http.MethodGet)
}
