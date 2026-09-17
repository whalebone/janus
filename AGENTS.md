# AGENTS.md
This file provides guidance to AI coding agents working with code in this repository.

## Service identity

Janus is a lightweight API Gateway written in Go: it proxies/routes requests to upstream services and applies
cross-cutting concerns (auth, rate limiting, circuit breaking, CORS, request/response transforms, retries) via a
plugin system. It exposes two HTTP surfaces: a public gateway port (proxying to upstreams) and a separate admin API
for managing API definitions, credentials and plugins.

This repo (`whalebone/janus`) is Whalebone's fork of [hellofresh/janus](https://github.com/hellofresh/janus), with
Whalebone-specific additions on top (Cassandra-backed repository, the `organization_auth` plugin). Version tags
follow `<upstream-version>-wb-<whalebone-patch>`, e.g. `4.0.0-wb-1.0.6`.

## Build / run / test

- `make build` — builds the `dist/janus` binary (`CGO_ENABLED=0`).
- `make test-unit` — `go test ./...`.
- `make test-integration` — runs `build/mocks.sh` (uploads WireMock stub fixtures for `auth_service`/`upstreams`,
  needs `DYNAMIC_AUTH_PORT`/`DYNAMIC_UPSTREAMS_PORT` or default ports 9088/9089) then `go test -tags=integration ./...`.
- `make test-features` — builds the binary, uploads mocks, then runs the Godog BDD suite in `features/` via
  `build/features.sh` (needs a MongoDB instance; starts two Janus instances on primary/secondary ports).
- `make lint` — runs `golangci-lint` (gofmt, golint, goimports) in Docker, pinned to `golangci/golangci-lint:v1.30.0`.
- Single test: `go test ./pkg/<package>/... -run TestName`.
- CI (`.github/workflows/testing.yml`) runs lint → unit → integration → features against `mongo:3` and two
  `wiremock` service containers, then uploads coverage to Codecov.
- Release (`.github/workflows/release.yml` + `.goreleaser.yml`) builds cross-platform binaries and a Docker image on
  GitHub release creation; this is the upstream hellofresh flow (pushes to `hellofreshtech/janus` on Docker Hub) —
  Whalebone's own image (`harbor.whalebone.io/whalebone/janus`) is built/tagged separately, see `.circleci/config.yml`.

## Architecture overview

- `main.go` → `cmd/` (Cobra commands: `root.go`, `server.go` wires `start`, `init.go` loads config/logging/stats/tracing).
- `cmd/server.go` blank-imports every plugin package (`pkg/plugin/*`) and auth provider (`pkg/jwt/*`) purely for their
  `init()` side-effect registration — a plugin/provider not imported there is invisible to the gateway even if its
  package compiles.
- `pkg/config` — `Specification` struct loaded via Viper (TOML/YAML/JSON file) or `envconfig` (env vars), see
  `janus.sample.toml` for the annotated file format and `pkg/config/specification.go` for the env var names
  (`envconfig` tags) and defaults.
- `pkg/api` — API `Definition`/`Plugin` domain types plus the `Repository` interface with three backends selected by
  the `database.dsn` scheme: `file://`, `mongodb://`, `cassandra://` (dispatch in `pkg/api/repository.go:BuildRepository`).
- `pkg/proxy`, `pkg/router`, `pkg/loader` — build the live router/reverse-proxy from `api.Definition`s;
  `loader.APILoader.RegisterAPI` walks each definition's plugins and invokes their registered `plugin.SetupFunc`.
- `pkg/plugin` — plugin registry (`RegisterPlugin`) and a pub/sub-style event system (`RegisterEventHook` /
  `EmitEvent`) with events `startup`, `admin_startup`, `reload`, `shutdown`, `setup` (see `pkg/plugin/events.go`).
  Each plugin under `pkg/plugin/<name>/` registers itself in an `init()` and typically hooks `StartupEvent` to wire
  its admin API routes and pick a repository backend (see `pkg/plugin/organization/setup.go` for the pattern).
- `pkg/web` — the admin/management HTTP API (definitions CRUD, credentials, health).
- `pkg/server` — top-level `Server` that owns both HTTP surfaces, the config-change channel, and emits plugin
  lifecycle events on startup/hot-reload.
- `cassandra/` — the Cassandra session holder (`cassandra/session.go`) and connection/keyspace-bootstrap wrapper
  (`cassandra/wrapper/`); `cassandra/schema.sql` is the keyspace schema, copied into the Docker image.
- `features/*.feature` + `features/bootstrap/` — Godog BDD specs exercising the running binary end-to-end (auth,
  hot-reload, proxying, health).

## Service Map

- Deployment manifests: [`whalebone/k8s-wb` → `deployments/jules-janus/`](https://github.com/whalebone/k8s-wb/tree/master/deployments/jules-janus)
  (namespace `jules-janus`, k8s app label `jules-janus-api`). Image: `harbor.whalebone.io/whalebone/janus`.
- Known upstream dependency from that deployment: `jules-api` (`http://jules-api.jules:8080/api/v1/`), fronted here
  as a proxied API, with login/credential checks against `jules-apicreds`.
- **Gotcha**: the `jules-janus-api-configmap` in that deployment sets env vars (`WB_API_0*`, `TRACING_OTLP_ENDPOINT`,
  the `wb_api_credentials_auth` plugin name) that don't exist anywhere in this repo's `pkg/config/specification.go`
  or plugin registry as of this version — confirm with the team whether that configmap is stale/for a different
  Janus variant before trusting it as this service's current runtime configuration.
- Full map (including what depends on this service) isn't derivable from this repo alone — run the `project-docs`
  skill for the authoritative version.
- When you add or remove a dependency — an env var pointing at another service, a new client, a new queue subject —
  or the deployment manifests move, update the Service Map (`docs/service-map.md` if it exists, otherwise this
  section) in the same change.

## Documentation

`docs/` is a GitBook-style manual (index: `docs/SUMMARY.md`); most relevant for agents:
- `docs/install/configuration.md` — config file formats (TOML/YAML/JSON) vs. env vars.
- `docs/plugins/*.md` — one page per plugin, including `organization_auth.md` (Whalebone-added).
- `docs/proxy/*.md` — routing/proxy semantics (host header, path stripping, priorities, load balancing).
- `docs/clustering/clustering.md`, `docs/misc/{tracing,monitoring,health_checks}.md`.
- `docs/upgrade/*.md` — breaking-change notes between major versions.

## Gotchas

- Plugins and JWT auth providers must be blank-imported in `cmd/server.go` to register themselves — adding a new
  package under `pkg/plugin/` or `pkg/jwt/` does nothing until it's added there.
- `Dockerfile`'s builder stage pins `golang:1.13.6-buster` while `go.mod` declares `go 1.15`; `Dockerfile.dev` uses
  `golang:1.14-alpine`. Verify the build toolchain version if bumping Go or dependencies.
- The Cassandra repository (`pkg/api/cassandra_repository.go`) expects the DSN path to encode
  `clusterHost/systemKeyspace/appKeyspace/timeout` (parsed by `parseDSN`); the `cassandra://` scheme is Whalebone-only
  and not documented in `docs/install/configuration.md` (which only covers upstream file/mongo options).
- `make test-features` builds the binary and starts two real Janus processes (primary + secondary ports) against a
  live MongoDB — it is not a fast/unit-level loop.

## Team

- Shape: split-fe-be
- Slice terminus: API contract / event
- E2E: QA-owned in wb-test-automation
