#!/bin/bash

set -e -u -x -o pipefail

export GO111MODULE=on
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64

rm -Rf bin/

go build \
    -o bin/backend \
    -v \
    -trimpath \
    ./cmd/backend

strip bin/backend

go build \
    -o bin/monitor-write \
    -v \
    -trimpath \
    ./cmd/monitor_write

strip bin/monitor-write
