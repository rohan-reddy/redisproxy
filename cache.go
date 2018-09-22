package main

import (
	"bytes"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"log"
	"net/http"
	"time"
)

// Linked list nodes.
type node struct {
	prev, next *node
	key, value string
	creationTime time.Time
}

func newNode(key, value string) *node {
	n := new(node)
	n.key = key
	n.value = value
	n.creationTime = time.Now()
	return n
}

// Creates a connection pool for Redis, capped at the max number of connections specified in Dockerfile.
func newPool(redisServer string, maxConnections int) *redis.Pool {
	return &redis.Pool{
		MaxIdle: maxConnections,
		MaxActive: maxConnections,
		IdleTimeout: 60 * time.Second,
		Dial: func () (redis.Conn, error) { return redis.Dial("tcp", redisServer) },
	}
}

// Contains pointers to head and tail of its linked list, a (string -> node) map keyed by entry key,
// and a Redis connection, as well as capacity and expirationTime settings.
type cache struct {
	conn redis.Conn
	head, tail *node
	key2ElementMap map[string]*node
	capacity int
	expirationTime time.Duration
}

func NewCache(redisServer string, capacity int, expirationTime int, maxConnections int) *cache {
	c := new(cache)
	conn, err := newPool(redisServer, maxConnections).Dial()
	if err != nil {
		log.Fatal(err)
	}

	c.conn = conn
	c.key2ElementMap = make(map[string]*node)
	c.capacity = capacity
	c.expirationTime = time.Duration(expirationTime) * time.Second
	return c
}

func (cache *cache) Close() {
	cache.conn.Close()
}

func (cache *cache) GetSize() int {
	return len(cache.key2ElementMap)
}

// This is the function that is attached to our HTTP service. It just parses the request header to get the
// requested key, and sends this off to our get() method. The resulting value is written as an HTTP response.
// Uncomment the logContents() call to see the cache contents after each call to GetValue(). Note, these log statements
// may not show up in terminal if the application is run with Docker.
func (cache *cache) GetValue(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("key")
	value, _ := cache.get(key)
	fmt.Fprint(w, string(value))
	//cache.logContents()
}


// Tries to fetch the value from the cache, otherwise fetches it from Redis.
func (cache *cache) get(key string) (value string, fetchedFromRedis bool) {
	value = cache.fetchFromCache(key)
	if value == "" || value == "E" {
		return cache.fetchFromRedis(key), true
	} else {
		return value, false
	}
}

// Fetches a value from Redis. If the key is not present, returns an empty string.
func (cache *cache) fetchFromRedis(key string) string {
	var data []byte
	data, err := redis.Bytes(cache.conn.Do("GET", key))
	if err != nil {
		return ""
	} else {
		value := string(data)
		cache.putInCache(key, value)
		return value
	}

}

// Returns the value if found in the cache, "E" if found in the cache but expired, and an empty string if not found.
func (cache *cache) fetchFromCache(key string) string {
	if foundNode, ok := cache.key2ElementMap[key]; ok {
		elapsed := time.Now().Sub(foundNode.creationTime)
		if elapsed > cache.expirationTime {
			cache.removeKey(key)
			return "E"
		}

		cache.removeNodeFromList(foundNode)
		cache.insertNodeAtListFront(foundNode)
		return foundNode.value
	} else {
		return ""
	}
}

// Places a key value pairing in the cache by creating a node, inserting it at the front of the linked list,
// and mapping the key to the new node in key2ElementMap.
func (cache *cache) putInCache(key, value string) {
	if key == "" {
		return
	}

	newNode := newNode(key, value)
	cache.insertNodeAtListFront(newNode)
	cache.key2ElementMap[key] = newNode

	if len(cache.key2ElementMap) > cache.capacity {
		lastNode := cache.tail
		if lastNode != nil {
			cache.removeKey(lastNode.key)
		}
	}
}

// Removes all trace of the key value pairing associated with the input key. Removes from both linked list and map.
func (cache *cache) removeKey(key string) {
	targetNode := cache.key2ElementMap[key]
	cache.removeNodeFromList(targetNode)
	delete(cache.key2ElementMap, key)
}

// Inserts a linked list node at the start of the list.
func (cache *cache) insertNodeAtListFront(newNode *node) {
	newNode.prev = nil
	newNode.next = cache.head
	if newNode.next != nil {
		newNode.next.prev = newNode
	}

	cache.head = newNode
	if cache.tail == nil {
		cache.tail = newNode
	}
}

// Removes a linked list node from the list.
func (cache *cache) removeNodeFromList(targetNode *node) (*node) {
	if targetNode.prev != nil {
		targetNode.prev.next = targetNode.next
	}

	if targetNode.next != nil {
		targetNode.next.prev = targetNode.prev
	}

	if targetNode == cache.head {
		cache.head = targetNode.next
	}

	if targetNode == cache.tail {
		cache.tail = targetNode.prev
	}

	return targetNode
}

// Logs contents of the cache in order from most to least recently used entry.
func (cache *cache) logContents() {
	curNode := cache.head
	var b bytes.Buffer
	for curNode != nil {
		b.WriteString(fmt.Sprintf("(%s, %s) -> ", curNode.key, curNode.value))
		curNode = curNode.next
	}
	log.Println(b.String())
}

