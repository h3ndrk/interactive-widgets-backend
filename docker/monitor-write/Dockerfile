FROM golang:1.14 AS builder

WORKDIR /usr/src
COPY . .
RUN make build-monitor-write

FROM scratch

COPY --from=builder /usr/src/bin/monitor-write /
ENTRYPOINT [ "/monitor-write" ]
