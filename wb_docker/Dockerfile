# build from project root dir using 'docker build -f wb_docker/Dockerfile -t <tag> .'
FROM golang:1.22.5-alpine AS builder

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

# Create directories for runtime image
RUN mkdir -p /opt/janus && \
    mkdir -p /etc/janus/apis && \
    mkdir -p /etc/janus/auth && \
    chown -R ${UID}:${UID} /etc/janus

# copy go mod and sum files
COPY go.mod go.sum ./

RUN apk add --no-cache bash make git ca-certificates && \
    go mod download && go mod verify && \
    rm -f ~/.gitconfig

# copy sources
COPY . ./

RUN export JANUS_BUILD_ONLY_DEFAULT=1 && \
    export VERSION=$VERSION && \
    make build && \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o "dist/entrypoint" ./wb_docker/entrypoint.go

#
RUN chmod u+x dist/janus && \
    chmod u+x dist/entrypoint


# runtime image
FROM scratch

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /opt/janus /opt/janus
COPY --from=builder /etc/janus /etc/janus


COPY --from=builder --chown=${UID}:${UID} /janus/dist/janus /opt/janus/janus
COPY --from=builder --chown=${UID}:${UID} /janus/dist/entrypoint /entrypoint
COPY wb_docker/janus.toml /etc/janus/janus.toml
COPY wb_docker/api_template.json /opt/janus/api_template.json

USER ${UID}:${UID}

ENTRYPOINT ["/entrypoint"]
