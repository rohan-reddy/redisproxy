package main

import (
	"bytes"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"log"
	"net/http"
	"time"
)

type node struct {
	prev, next *node
	key, value string
	creationTime time.Time
}

func NewNode(key, value string) *node {
	n := new(node)
	n.key = key
	n.value = value
	n.creationTime = time.Now()
	return n
}

type cache struct {
	conn redis.Conn
	head, tail *node
	key2ElementMap map[string]*node
	capacity int
	expirationTime time.Duration
}

func NewCache(redisServer string, capacity int, expirationTime int) *cache {
	c := new(cache)
	conn, err := redis.Dial("tcp", redisServer)
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

func (cache *cache) GetValue(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("key")
	value := cache.get(key)
	if value == "" || value == "E" {
		value = cache.getFromRedis(key)
	}

	fmt.Fprint(w, string(value))
	cache.logContents()
}


func (cache *cache) getFromRedis(key string) string {
	var data []byte
	data, err := redis.Bytes(cache.conn.Do("GET", key))
	if err != nil {
		return ""
	} else {
		value := string(data)
		cache.put(key, value)
		return value
	}

}

func (cache *cache) get(key string) string {
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

func (cache *cache) put(key, value string) {
	if key == "" {
		return
	}

	newNode := NewNode(key, value)
	cache.insertNodeAtListFront(newNode)
	cache.key2ElementMap[key] = newNode

	if len(cache.key2ElementMap) > cache.capacity {
		lastNode := cache.tail
		if lastNode != nil {
			cache.removeKey(lastNode.key)
		}
	}
}

func (cache *cache) removeKey(key string) {
	targetNode := cache.key2ElementMap[key]
	cache.removeNodeFromList(targetNode)
	delete(cache.key2ElementMap, key)
}


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

func (cache *cache) logContents() {
	curNode := cache.head
	var b bytes.Buffer
	for curNode != nil {
		b.WriteString(fmt.Sprintf("(%s, %s) -> ", curNode.key, curNode.value))
		curNode = curNode.next
	}
	log.Println(b.String())
}
