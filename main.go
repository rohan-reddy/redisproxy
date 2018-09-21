package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
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

	cache := NewCache(redisServer, capacity, expiryTime)
	defer cache.Close()

	router := mux.NewRouter()
	router.HandleFunc("/", cache.GetValue).Methods("GET")


	hostAddress := fmt.Sprintf(":%d", localhostPort)
	log.Fatal(http.ListenAndServe(hostAddress, router))
}




