package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

const (
	InvalidationThreshold     = time.Second * 600
	InvalidationCheckInterval = time.Second * 30
)

// Struct for importing and using the json body of the http call to invalidate
type Post struct {
	//URL: URL of the site that will be cleared.
	URL string `json:"URL"`
	//A json array of paths that will be in the invalidation.
	PATHS []string `json:"PATHS"`
}

type Status struct {
	//DIST_ID: Cloudfront Distribution ID of the site that will have it's cache cleared.
	URL string `json:"URL"`
	//ID: The invalidation ID that's returned from the invalidate web route.
	ID *string `json:"ID"`
}

var ctx = context.Background()

var redis_host string = os.Getenv("REDIS_HOST")
var redis_port string = os.Getenv("REDIS_PORT")

func main() {
	//Set the port the API will listen on, default to 8000 if not provided.
	port := os.Getenv("API_PORT")
	if len(port) == 0 {
		port = ":8000"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     redis_host + ":" + redis_port,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	cfData := getCFDistributions()

	loadCloudfrontData(cfData, rdb)

	//Initialize our router
	router := mux.NewRouter()

	//	  router.HandleFunc("/books", GetBooks).Methods("GET")
	//    router.HandleFunc("/books/{id}", GetBook).Methods("GET")
	//    router.HandleFunc("/books", CreateBook).Methods("POST")
	//    router.HandleFunc("/books/{id}", UpdateBook).Methods("PUT")
	//    router.HandleFunc("/books/{id}", DeleteBook).Methods("DELETE")

	//Small route/function to give us a healthcheck endpoint.
	router.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})

	//Our main routes/endpoints/functions
	router.HandleFunc("/invalidate", authorization(invalidate)).Methods("POST")
	router.HandleFunc("/status", invalidation_status).Methods("POST")

	//Serve the API
	log.Fatal(http.ListenAndServe(port, router))
}
