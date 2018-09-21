package main

import (
	"encoding/json"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"io/ioutil"
	"net/http"
	"testing"
)

var (
	redisServer=":6379"
	capacity=2
	expiryTime=60
	localhostPort=8080
	redisDirect redis.Conn
	client = &http.Client{}
)

func requestValue(key string) string {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/", localhostPort), nil)
	if err != nil {
		panic(err)
	}

	req.Header.Add("key", key)
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	response := new(response)
	err = json.Unmarshal([]byte(body), &response)
	if err != nil {
		panic(err)
	}

	return response.Value
}

func TestRedisTestServerBootedSuccessfully(t *testing.T) {
	redisDirect, redisErr := redis.Dial("tcp", redisServer)
	if redisErr != nil {
		t.Errorf("Failed to connect to test Redis server")
	}

	for i := 1; i <= 5; i++ {
		redisDirect.Do("SET", fmt.Sprintf("k%d", i), fmt.Sprintf("v%d", i))
	}
}
//
//func TestCacheAcceptsHTTPRequests(t *testing.T) {
//	k1 := "k1"
//	observedV1 := requestValue(k1)
//	expectedV1 := "v1"
//	if observedV1 != expectedV1 {
//		t.Errorf("For key %s, expected %s but got %s", k1, expectedV1, observedV1)
//	}
//}

//func TestCacheItemsExpireAfterSpecifiedLimit(t *testing.T) {
//	cache := NewCache(redisServer, capacity, expiryTime)
//	k1, k2, k3 := "k1", "k2", ""
//}
//
//func TestLeastRecentlyUsedItemIsEvictedAtCapacity(t *testing.T) {
//
//}
//
//func TestCacheSizeNeverExceedsSpecifiedCapacity(t *testing.T) {
//
//}
