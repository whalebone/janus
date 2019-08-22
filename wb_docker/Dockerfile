# build from project root dir using 'docker build -f wb_docker/Dockerfile -t <tag> .'
FROM golang:1.10-alpine AS builder

ARG VERSION='3.8.11-rc.1-wb-0.0.1-docker'

WORKDIR /go/src/github.com/hellofresh/janus

COPY . ./

RUN apk add --update bash make git
RUN export JANUS_BUILD_ONLY_DEFAULT=1 && \
    export VERSION=$VERSION && \
    make build

# ---

FROM alpine

COPY --from=builder /go/src/github.com/hellofresh/janus/dist/janus /

RUN apk add --no-cache ca-certificates
RUN mkdir -p /etc/janus/apis && \
    mkdir -p /etc/janus/auth

ENV JANUS_CONF "/etc/janus/janus.toml"
COPY wb_docker/janus.toml ${JANUS_CONF}

ENV PUBLIC_API_CONF "/etc/janus/apis/public_api.json"
COPY wb_docker/public_api.json ${PUBLIC_API_CONF}

RUN apk add --update curl sed && \
    rm -rf /var/cache/apk/*

COPY wb_docker/janus.sh /
RUN chmod u+x /janus.sh
ENTRYPOINT ["./janus.sh"]
