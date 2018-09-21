FROM golang:onbuild
ENV redisServer="host.docker.internal:6379"
ENV capacity=2
ENV expiryTime=60
ENV localhostPort=8080
EXPOSE ${localhostPort}