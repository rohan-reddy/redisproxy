FROM golang:onbuild

# All of these ENV variables are configurable.
# Note that if your Redis server is hosted locally, such as localhost:6379, you need to use "host.docker.internal"
# instead of "localhost". This is not guaranteed to work outside of Mac OS X, according to Docker.
ENV redisServer="host.docker.internal:6379"
ENV capacity=2
ENV expiryTime=60
ENV maxConnections=3
# If you change localhostPort, make sure to also change it in Makefile.
ENV localhostPort=8080
EXPOSE ${localhostPort}