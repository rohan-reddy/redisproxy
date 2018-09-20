package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"redisproxy/types"
)

func main() {
	addressPtr := flag.String("address", ":6379", "Address of backing Redis instance")
	capacityPtr := flag.Int("capacity", 3, "Cache capacity")
	expiryPtr := flag.Int("expiry", 60, "Cache expiry time in seconds")
	portPtr := flag.Int("port", 8080, "TCP/IP port number the proxy will listen on")
	flag.Parse()

	cache := types.NewCache(*addressPtr, *capacityPtr, *expiryPtr)
	defer cache.Close()

	router := mux.NewRouter()
	router.HandleFunc("/get", cache.GetValue).Methods("GET")

	hostAddress := fmt.Sprintf(":%d", *portPtr)
	log.Fatal(http.ListenAndServe(hostAddress, router))
}




