# build from project root dir using 'docker build -f wb_docker/Dockerfile -t <tag> .'
FROM golang:1.12-alpine AS builder

ARG VERSION='3.8.22-wb-0.0.4'

WORKDIR /janus

COPY . ./

RUN apk add --update bash make git
RUN export JANUS_BUILD_ONLY_DEFAULT=1 && \
    export VERSION=$VERSION && \
    make deps build

# ---

FROM alpine

RUN apk add --no-cache ca-certificates

ENV JANUS_ROOT="/opt/janus"
ENV API_CONF_DIR "/etc/janus/apis"
ENV JANUS_CONF "/etc/janus/janus.toml"
ENV API_CONF_TEMPLATE "${JANUS_ROOT}/api_template.json"
ENV JANUS_BIN "${JANUS_ROOT}/janus"

RUN mkdir -p ${JANUS_ROOT} && \
    mkdir -p ${API_CONF_DIR} && \
    mkdir -p /etc/janus/auth

COPY --from=builder /janus/dist/janus ${JANUS_BIN}
RUN chmod u+x ${JANUS_BIN}
COPY wb_docker/janus.toml ${JANUS_CONF}
COPY wb_docker/janus.sh ${JANUS_ROOT}/janus.sh
RUN chmod u+x ${JANUS_ROOT}/janus.sh
COPY wb_docker/api_template.json ${API_CONF_TEMPLATE}

RUN apk add --update curl sed && \
    rm -rf /var/cache/apk/*

ENTRYPOINT ["/opt/janus/janus.sh"]
