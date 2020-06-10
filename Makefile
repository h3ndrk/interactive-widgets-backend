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
docker-backend:
	docker build -t docker.pkg.github.com/h3ndrk/inter-md/backend -f docker/backend/Dockerfile .

.PHONY: docker-monitor-write
docker-monitor-write:
	docker build -t docker.pkg.github.com/h3ndrk/inter-md/monitor-write -f docker/monitor-write/Dockerfile .

.PHONY: docker-docker-build
docker-docker-build:
	docker build -t docker.pkg.github.com/h3ndrk/inter-md/docker-build -f docker/docker-build/Dockerfile .

