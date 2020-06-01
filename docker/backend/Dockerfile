FROM golang:1.14 AS builder

WORKDIR /usr/src
COPY . .
RUN make build-backend

FROM scratch

COPY --from=builder /usr/src/bin/backend /
ENTRYPOINT [ "/backend" ]
