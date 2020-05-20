FROM golang:1.14 AS builder

WORKDIR /usr/src
COPY go.mod go.sum main.go read.go wait.go /usr/src/
RUN go build -o monitor-write

FROM scratch

COPY --from=builder /usr/src/monitor-write /
ENTRYPOINT [ "/monitor-write" ]
