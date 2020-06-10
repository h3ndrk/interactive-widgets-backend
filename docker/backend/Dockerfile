FROM golang:1.14 AS builder

WORKDIR /usr/src
COPY . .
RUN strip /usr/src/docker/dind/docker
RUN make build-backend

FROM scratch

COPY --from=builder /usr/src/bin/backend /usr/src/docker/dind/docker /
ENV PATH=/
ENTRYPOINT [ "/backend" ]
