FROM golang:1.14 AS builder

WORKDIR /usr/src
COPY . .
RUN strip /usr/src/docker/dind/docker
RUN make build-docker-build

FROM scratch

COPY --from=builder /usr/src/bin/docker-build /usr/src/docker/dind/docker /
ENV PATH=/
ENTRYPOINT [ "/docker-build" ]
