.PHONY: all
all:

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


.PHONY: docker-backend
docker-backend:
	docker build -t containerized-playground-backend -f docker/backend/Dockerfile .

.PHONY: docker-monitor-write
docker-monitor-write:
	docker build -t containerized-playground-monitor-write -f docker/monitor-write/Dockerfile .

