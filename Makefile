.PHONY: all
all: build-backend build-monitor-write build-docker-build

.PHONY: build-backend
build-backend:
	rm -Rf bin/backend
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/backend -v -trimpath ./cmd/backend
	strip bin/backend

.PHONY: build-monitor-write
build-monitor-write:
	rm -Rf bin/monitor-write
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/monitor-write -v -trimpath ./cmd/monitor_write
	strip bin/monitor-write

.PHONY: build-docker-build
build-docker-build:
	rm -Rf bin/docker-build
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/docker-build -v -trimpath ./cmd/docker_build
	strip bin/docker-build


.PHONY: docker-backend
docker-backend: docker/dind/docker
	docker build -t inter-md-backend -f docker/backend/Dockerfile .

.PHONY: docker-monitor-write
docker-monitor-write:
	docker build -t inter-md-monitor-write -f docker/monitor-write/Dockerfile .

.PHONY: docker-docker-build
docker-docker-build:
	docker build -t inter-md-docker-build -f docker/docker-build/Dockerfile .

.PHONY: docker-nginx-frontend
docker-nginx-frontend:
	docker build -t inter-md-nginx-frontend -f docker/nginx-frontend/Dockerfile .

docker/dind/docker:
	mkdir -p /tmp
	cd /tmp && wget --output-document docker.tgz https://download.docker.com/linux/static/stable/x86_64/docker-$(firstword $(subst -, ,$(shell docker version --format '{{.Client.Version}}'))).tgz
	cd /tmp && tar xzf docker.tgz
	mkdir -p docker/dind
	mv /tmp/docker/docker docker/dind/docker
	rm -Rf /tmp/docker /tmp/docker.tgz


.PHONY: docker-images
docker-images: docker-backend docker-monitor-write docker-docker-build docker-nginx-frontend
	docker run --rm -v "/var/run/docker.sock:/var/run/docker.sock" -v "$(shell pwd)/pages:/pages" inter-md-docker-build

.PHONY: up
up:
	docker-compose --file docker/docker-compose.yml --project-name inter-md up && docker-compose --file docker/docker-compose.yml --project-name inter-md rm --force --stop -v
