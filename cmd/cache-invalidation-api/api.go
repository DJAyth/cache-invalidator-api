package main

import (
	"encoding/json"
	"net/http"
	"os"

	muxctx "github.com/gorilla/context"
)

// Simple function for adding in some basic authorization to the API. Set an env var named API_KEY and add the Authorization header with the value matching it.
func authorization(next http.HandlerFunc) http.HandlerFunc {
	api_key := os.Getenv("API_KEY")
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != api_key {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]bool{"Unauthorized": false})
			return
		}
		muxctx.Set(r, "isAuthorized", true)
		next.ServeHTTP(w, r)
	}
}
