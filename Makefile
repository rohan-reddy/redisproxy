# Go, Docker, Redis parameters
GOCMD=go
USER_NAME=reddyroh
APP_NAME=redisproxy
TAG_NAME=v1
IMAGE_NAME=${USER_NAME}/${APP_NAME}:${TAG_NAME}
LOCALHOST_PORT=8080

build:
		docker build -t redisproxy .
clean:
		docker-compose down
test:
		docker-compose down
		docker-compose build
		docker-compose up -d
		go test
dummy:
		docker-compose down
		docker-compose build
		docker-compose up
		redis-cli set k1 v1
		redis-cli set k2 v2
		redis-cli set k3 v3
run:
		docker run -p ${LOCALHOST_PORT}:${LOCALHOST_PORT} ${APP_NAME}
publish:
		make build
		make test
		docker tag ${APP_NAME} ${IMAGE_NAME}
		docker push ${IMAGE_NAME}
