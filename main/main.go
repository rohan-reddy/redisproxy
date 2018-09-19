package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"redisproxy/types"
)

func main() {
	router := mux.NewRouter()
	cache := types.NewCache(":6379", 3, 10)
	router.HandleFunc("/get", cache.GetValue).Methods("GET")
	log.Fatal(http.ListenAndServe(":8080", router))
}




