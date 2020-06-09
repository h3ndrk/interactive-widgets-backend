FROM golang:1.14 AS builder

WORKDIR /usr/src
COPY . .
RUN make build-docker-build

FROM scratch

COPY --from=builder /usr/src/bin/docker-build /
ENTRYPOINT [ "/docker-build" ]
