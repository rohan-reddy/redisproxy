package main

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

/**
Test file for the cache. Names are fairly self-explanatory.
There is a test for each requirement of the proxy, except for sequential concurrent processing.
 */

var (
	redisServer = ":6379"
	localhostPort = 8080
	redisDirect redis.Conn
	client = &http.Client{}
	maxConnections = 3
	k1, k2, k3, k4 = "k1", "k2", "k3", "k4"
	v1, v2, v3, v4 = "v1", "v2", "v3", "v4"
)

// Helper function used for making requests to the HTTP service booted in docker-compose.
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

// Helper function to make requests to the Redis store booted in docker-compose.
func setKeyValPairsInRange(start, end int) {
	for i := start; i <= end; i++ {
		redisDirect.Do("SET", fmt.Sprintf("k%d", i), fmt.Sprintf("v%d", i))
	}
}

// Checks that the Redis image is up, and stores key value pairings in it for testing.
func TestRedisTestServerBootedSuccessfully(t *testing.T) {
	var redisErr error
	redisDirect, redisErr = redis.Dial("tcp", redisServer)
	setKeyValPairsInRange(1, 5)
	if redisErr != nil {
		t.Errorf("Failed to connect to test Redis server")
	}
}

// Checks that the HTTP service is up and can successfully connect to the Redis image.
func TestCacheAcceptsHTTPRequests(t *testing.T) {
	observedV1 := requestValue(k1)
	expectedV1 := "v1"
	if observedV1 != expectedV1 {
		t.Errorf("For key %s, expected %s but got %s", k1, expectedV1, observedV1)
	}
}

// Checks that the cache fetches values from Redis when it doesn't have them, but otherwise doesn't access Redis.
func TestCacheStoresValuesFetchedFromRedis(t *testing.T) {
	cache := NewCache(redisServer, 2, 60, maxConnections)
	defer cache.Close()

	_, fetchedFromRedis := cache.get(k1)
	if fetchedFromRedis != true {
		t.Errorf("For key %s, claimed val was not fetched from Redis but it must have been", k1)
	}

	_, fetchedFromRedis = cache.get(k2)
	if fetchedFromRedis != true {
		t.Errorf("For key %s, claimed val was not fetched from Redis but it must have been", k2)
	}

	_, fetchedFromRedis = cache.get(k2)
	if fetchedFromRedis != false {
		t.Errorf("For key %s, claimed val was fetched from Redis but it should have been fetched from cache", k2)
	}

	_, fetchedFromRedis = cache.get(k1)
	if fetchedFromRedis != false {
		t.Errorf("For key %s, claimed val was fetched from Redis but it should have been fetched from cache", k1)
	}
}

// Checks that cache items expire after the user-configured time limit.
func TestCacheItemsExpireAfterSpecifiedLimit(t *testing.T) {
	cache := NewCache(redisServer, 1, 3, maxConnections)
	defer cache.Close()

	cache.putInCache(k1, v1)
	time.Sleep(2 * time.Second)
	if cache.fetchFromCache(k1) != v1 {
		t.Errorf("Value expired or nil when it should have remained in cache")
	}

	time.Sleep(1 * time.Second)
	if cache.fetchFromCache(k1) != "E" {
		t.Errorf("Value expected to be expired, was not expired")
	}
}

// Checks that when the cache is at capacity, new entries evict the least recently used cache entry.
func TestLeastRecentlyUsedItemIsEvictedAtCapacity(t *testing.T) {
	cache := NewCache(redisServer, 3, 60, maxConnections)
	defer cache.Close()

	cache.putInCache(k1, v1)
	cache.putInCache(k2, v2)
	cache.putInCache(k3, v3)

	cache.fetchFromCache(k3)
	cache.fetchFromCache(k1)
	cache.putInCache(k4, v4)

	if !(cache.fetchFromCache(k3) == v3 && cache.fetchFromCache(k1) == v1 && cache.fetchFromCache(k4) == v4) {
		t.Errorf("Value expired or evicted when it should have remained in cache")
	}

	if cache.fetchFromCache(k2) != "" {
		t.Errorf("Value expired or present when it should have been evicted as the LRU item")
	}
}

// Checks that the cache size never exceeds the user-configured capacity.
func TestCacheSizeCompliesWithSpecifiedCapacity(t *testing.T) {
	cache := NewCache(redisServer, 3, 60, maxConnections)
	defer cache.Close()

	cache.putInCache(k1, v1)
	cache.putInCache(k2, v2)
	cache.putInCache(k3, v3)
	cache.putInCache(k4, v4)

	if cache.GetSize() > 3 {
		t.Errorf("Cache size exceeded specified capacity")
	}

	if !(cache.fetchFromCache(k2) == v2 && cache.fetchFromCache(k3) == v3 && cache.fetchFromCache(k4) == v4) {
		t.Errorf("Value expired or evicted when it should have remained in cache")
	}

	if cache.fetchFromCache(k1) != "" {
		t.Errorf("Value expired or present when it should have been evicted as the LRU item")
	}
}