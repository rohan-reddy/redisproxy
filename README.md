# redisproxy

### High-Level Overview
Everything is contained in the root directory redisproxy, including:
- Makefile (used for 1-click build & test, and 1-click run)
- Dockerfile (used for building the app)
- docker-compose.yml (used for running end to end tests)
- main.go (boots up the HTTP service that listens on the user's chosen port)
- cache.go (defines all operations related to the underlying cache)
- cache_test.go (unit and integration tests for the cache)
- vendor (directory containing 3rd party libraries "mux" and "redigo")

##### How do these files fit together at runtime?
1. The main() function in main.go is called at startup. A cache is initialized 
per the configuration settings in Dockerfile, and its GetValue() function is attached
to HTTP GET requests at localhost (the port number is set in Dockerfile).
2. At this point in time, the cache connects to the Redis server at the address defined 
in Dockerfile. Other configuration settings in Dockerfile include simultaneous connections
to Redis, the expiry time of entries in the cache, and the maximum capacity of the cache.
3. When a GET request is made to localhost at the specified port, the key string is parsed
from the request header. If the cache currently contains an entry with the given key,
the app returns an HTTP response of the associated value string. If not, the app retrieves
the value from Redis, querying the linked Redis server with a "GET" command. If this also returns
nothing, an empty string is returned. 

##### Why are the files not contained within dedicated "src" and "tst" folders?
I played around with Dockerfile configurations for a while to get the app to build 
successfully with my original structure. In the end, I couldn't get any directory structure
working with "docker build" besides lumping all the files into the root folder.

##### What are the 3rd party libraries for?
"mux" handles starting up the HTTP service and "redigo" is a Redis client.

##### What is docker-compose used for?
If you examine Makefile, you'll see the sequence of events that occurs when you run
"make test". First, any currently active containers started with docker-compose are shut down.
Then, we build the app and boot up an app instance at port 8080 along with a lightweight
Alpine Linux based Redis image at port 6379. In cache_test.go, some tests make requests
to the app instance directly to check the health of the HTTP service, while others connect directly
to the booted up Redis store to get and retrieve values for testing. The former category
makes use of the Redis store and essentially simulates an end-to-end environment for the application,
since it includes a Redis store similar to what the user may connect the app to.

### Cache Operation Design and Algorithmic Complexity
##### Design
I created a struct for doubly linked list nodes, where each one contains a key, value, pointers to the next and previous nodes, time of creation. This 
creation time value is used for calculating whether a cache entry has expired. The cache struct contains a Redis
connection, capacity and expiration time, pointers to the head and tail of the linked list of 
cache entries (represented as nodes), and a string to node map, keyed by each node's key. Considering this is
an LRU cache, we have to move entries around frequently in the list, which is why a linked list was the best choice.
The only downside of a linked list is finding individual nodes, and our map solves this problem for us.

##### Algorithmic Complexity
All operations are constant time.
- get(key string): O(1). This encompasses fetching from the cache and fetching from Redis if necessary.
- putInCache(key, value string): O(1). Inserts node at front of linked list and adds mapping to map.
- removeKey(key string): O(1). Retrieves node via the map, removes it by adjusting pointers in the surrounding nodes,
and removes the mapping from the map.

### Running the Proxy 
In Dockerfile: adjust target Redis server address, cache capacity and expiry time, max connections to the Redis server, 
and port on which to host the proxy. If you modify the localhost port, make sure to also modify it in Makefile. Then simply open up a terminal, navigate to the project root directory, and run:
<br/>`make build`<br/>`make run`

### Testing the Proxy
In the project root directory, run:
<br/>`make test`

### Project Steps Duration
All durations are approximate.
- Go tutorial before diving in - 30 minutes. I've never coded in Go before, but I just looked at some basic syntax and figured out the rest along the way.
- Setting up HTTP service in main.go - 5 minutes, was pretty self-explanatory with mux.
- Implementing cache.go - 90 minutes due to debugging time, I knew how I wanted to implement the cache but didn't initially understand how Go pointers worked, so I spent a while debugging those issues.
- Configuring Docker - ~5 hrs, I've never used Docker before and just couldn't get the build set up the way I wanted it. Even in its current rendition, I'm not happy with the app's structure. Included in this amount of time was setting up my integration testing environment
using Docker compose
- Writing Makefile - 5 min, I've never used Makefile before either but it was pretty self-explanatory.
- Writing tests in cache_test.go - 2 hrs, 1.75 of which was figuring out how to get docker-compose to set up the environment I wanted for testing.
- Adding Redis multiple connection pool - 5 min, was pretty self-explanatory with redigo library.
- Documentation - 45 min including this README.

I'd clock the total time I spent on this project at around 10 hours. The main factors influencing this:
- Lack of familiarity with Go, which added varying amounts of time throughout reading documentation.
- Lack of familiarity with Docker and docker-compose. This was the big one, I'd attribute around 7 hours to trying to get these components working.

This was a good learning experience. Go is a very clean language and I will definitely use it in the future. Docker is also incredibly useful, I learned a lot about how to use it
but still have quite a ways to go. I spent a lot of time trying to create an extremely lightweight image for the app by using multi-stage Docker builds with Alpine-linux based
Go images, but they didn't pan out.

### Unimplemented Requirements 
- Sequential concurrent processing - I wasn't sure exactly what this meant in the context of an HTTP service, but that being said,
this is the first time I've ever written one. It seems to me that requests would not be thrown out if they hit the service while
it's processing another one, at least from what I read about Go. I did add an explicit pool, with max connections configurable by the user,
for the target Redis server in the app. So it's possible that the proxy satisfies this requirement. I didn't have much of an idea about how to
go about testing this capability, since I didn't fully understand it. This is the only basic requirement with no explicit test in 
cache_test.go.
- Bonus requirements - Looked at them briefly, but understanding sequential concurrent processing is a prerequisite to 
implementing parallel processing. The RESP protocol looked viable given enough time, but I didn't have the extra time 
this week to commit to this project aside from what I already put in.








