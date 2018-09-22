package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"strconv"
)

/**
This file boots up the HTTP service and is the starting point for running the application.
 */
func main() {
	// Extract environment variables, set via Dockerfile.
	redisServer := os.Getenv("redisServer")
	capacity, capErr := strconv.Atoi(os.Getenv("capacity"))
	if capErr != nil {
		log.Fatal("Cache capacity must be an integer value")
	}

	expiryTime, expErr := strconv.Atoi(os.Getenv("expiryTime"))
	if expErr != nil {
		log.Fatal("Expiry time must be an integer value")
	}

	localhostPort, portErr := strconv.Atoi(os.Getenv("localhostPort"))
	if portErr != nil {
		log.Fatal("Localhost port must be an integer value")
	}

	maxConnections, maxConnErr := strconv.Atoi(os.Getenv("maxConnections"))
	if maxConnErr != nil {
		log.Fatal("Max connections must be an integer value")
	}

	// Initialize the cache, and defer closing its Redis connection when the service is stopped.
	cache := NewCache(redisServer, capacity, expiryTime, maxConnections)
	defer cache.Close()

	// Set the handler for GET requests to the GetValue function in cache.
	router := mux.NewRouter()
	router.HandleFunc("/", cache.GetValue).Methods("GET")

	// Set up the HTTP service to listen at localhost at the user-configured port.
	hostAddress := fmt.Sprintf(":%d", localhostPort)
	log.Fatal(http.ListenAndServe(hostAddress, router))
}




