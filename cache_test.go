package main

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

var (
	redisServer = ":6379"
	localhostPort = 8080
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

	return string(body)
}

func setKeyValPairsInRange(start, end int) {
	for i := start; i <= end; i++ {
		redisDirect.Do("SET", fmt.Sprintf("k%d", i), fmt.Sprintf("v%d", i))
	}
}

func TestRedisTestServerBootedSuccessfully(t *testing.T) {
	var redisErr error
	redisDirect, redisErr = redis.Dial("tcp", redisServer)
	setKeyValPairsInRange(1, 5)
	if redisErr != nil {
		t.Errorf("Failed to connect to test Redis server")
	}
}

func TestCacheAcceptsHTTPRequests(t *testing.T) {
	k1 := "k1"
	observedV1 := requestValue(k1)
	expectedV1 := "v1"
	if observedV1 != expectedV1 {
		t.Errorf("For key %s, expected %s but got %s", k1, expectedV1, observedV1)
	}
}

func TestCacheItemsExpireAfterSpecifiedLimit(t *testing.T) {
	cache := NewCache(redisServer, 1, 3)
	defer cache.Close()

	cache.put("k1", "v1")
	time.Sleep(2 * time.Second)
	if cache.get("k1") != "v1" {
		t.Errorf("Value expired or nil when it should have remained in cache")
	}

	time.Sleep(1 * time.Second)
	if cache.get("k1") != "E" {
		t.Errorf("Value expected to be expired, was not expired")
	}
}

func TestLeastRecentlyUsedItemIsEvictedAtCapacity(t *testing.T) {
	cache := NewCache(redisServer, 3, 60)
	defer cache.Close()

	k1, k2, k3, k4 := "k1", "k2", "k3", "k4"
	v1, v2, v3, v4 := "v1", "v2", "v3", "v4"

	cache.put(k1, v1)
	cache.put(k2, v2)
	cache.put(k3, v3)

	cache.get(k3)
	cache.get(k1)
	cache.put(k4, v4)

	if !(cache.get(k3) == v3 && cache.get(k1) == v1 && cache.get(k4) == v4) {
		t.Errorf("Value expired or evicted when it should have remained in cache")
	}

	if cache.get(k2) != "" {
		t.Errorf("Value expired or present when it should have been evicted as the LRU item")
	}
}


func TestCacheSizeCompliesWithSpecifiedCapacity(t *testing.T) {
	cache := NewCache(redisServer, 3, 60)
	defer cache.Close()

	k1, k2, k3, k4 := "k1", "k2", "k3", "k4"
	v1, v2, v3, v4 := "v1", "v2", "v3", "v4"

	cache.put(k1, v1)
	cache.put(k2, v2)
	cache.put(k3, v3)
	cache.put(k4, v4)

	if cache.GetSize() > 3 {
		t.Errorf("Cache size exceeded specified capacity")
	}

	if !(cache.get(k2) == v2 && cache.get(k3) == v3 && cache.get(k4) == v4) {
		t.Errorf("Value expired or evicted when it should have remained in cache")
	}

	if cache.get(k1) != "" {
		t.Errorf("Value expired or present when it should have been evicted as the LRU item")
	}
}
