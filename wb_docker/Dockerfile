# build from project root dir using 'docker build -f wb_docker/Dockerfile -t <tag> .'
FROM golang:1.20-alpine AS builder

ARG VERSION="dev-4.0.0-wb-0.0.0"

WORKDIR /janus

# Create appuser.
ENV USER=appuser
ENV UID=10001
# See https://stackoverflow.com/a/55757473/12429735RUN
RUN adduser \
    --disabled-password \
    --gecos "" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

# copy go mod and sum files
COPY go.mod go.sum ./

RUN apk add --no-cache bash make git ca-certificates && \
    go mod download && go mod verify && \
    rm -f ~/.gitconfig

# copy sources
COPY . ./

RUN export JANUS_BUILD_ONLY_DEFAULT=1 && \
    export VERSION=$VERSION && \
    make build


# runtime image
FROM alpine

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

RUN apk add --no-cache curl sed

ENV JANUS_ROOT="/opt/janus"
ENV API_CONF_DIR "/etc/janus/apis"
ENV JANUS_CONF "/etc/janus/janus.toml"
ENV API_CONF_TEMPLATE "${JANUS_ROOT}/api_template.json"
ENV JANUS_BIN "${JANUS_ROOT}/janus"

RUN mkdir -p ${JANUS_ROOT} && \
    mkdir -p ${API_CONF_DIR} && \
    mkdir -p /etc/janus/auth && \
    chown appuser:appuser /etc/janus

COPY --from=builder /janus/dist/janus ${JANUS_BIN}
RUN chmod u+x ${JANUS_BIN}
COPY wb_docker/janus.toml ${JANUS_CONF}
COPY wb_docker/janus.sh ${JANUS_ROOT}/janus.sh
RUN chmod u+x ${JANUS_ROOT}/janus.sh
COPY wb_docker/api_template.json ${API_CONF_TEMPLATE}

USER appuser:appuser

ENTRYPOINT ["/opt/janus/janus.sh"]
