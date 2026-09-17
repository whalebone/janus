# AGENTS.md
This file provides guidance to AI coding agents working with code in this repository.

## Service identity

Janus is a lightweight, single-binary API Gateway written in Go (originally open-sourced by
HelloFresh at [hellofresh/janus](https://github.com/hellofresh/janus), now maintained as a fork at
`whalebone/janus` for the MotivLabs/Impulse product). It sits in front of upstream APIs and handles
routing, auth (JWT/OAuth2/Basic), rate limiting, circuit breaking, CORS, request/response
transformation, and distributed tracing, driven by API definitions stored in Mongo, Cassandra, or a
local file tree.

## Build / run / test

- `make build` — builds `dist/janus` (`CGO_ENABLED=0`).
- `make test-unit` — `go test ./...`. Single test: `go test ./pkg/proxy/... -run TestName`.
- `make test-integration` — runs `build/mocks.sh` (uploads WireMock fixtures from
  `assets/stubs/{auth-service,upstreams}` to local WireMock instances) then `go test -tags=integration ./...`.
  Requires `DYNAMIC_UPSTREAMS_PORT` / `DYNAMIC_AUTH_PORT` pointing at running WireMock containers.
- `make test-features` — builds the binary, uploads mocks, then runs `build/features.sh`: starts two
  Janus instances (primary + secondary, see `PORT_SECONDARY`/`API_PORT_SECONDARY`) against a real
  MongoDB and drives them with godog (Cucumber) scenarios under `features/*.feature`. Requires
  `DYNAMIC_MONGO_PORT`.
- `make lint` — runs `golangci-lint` (v1.30.0, pinned) in Docker; enabled linters: `gofmt`, `golint`,
  `goimports` (see `.golangci.yml`).
- Run locally: `go run . start -c janus.toml` (copy `janus.sample.toml` as a starting point), or
  `./dist/janus start` after `make build`. Config can come from a TOML file or entirely from env vars
  (`config.LoadEnv`, see `pkg/config/specification.go` for the `envconfig` tag names).
- `./dist/janus check <config-file>` validates a config file without starting the server.
- CI (`.github/workflows/testing.yml`) runs lint → unit → integration → features on every push/PR to
  `master`, against real `mongo`, and two `wiremock` service containers (`upstreams`, `auth_service`).

## Architecture

- Entrypoint: `main.go` → `cmd.NewRootCmd` (`cmd/root.go`) wires the `start` (`cmd/server.go`) and
  `check` (`cmd/check.go`) subcommands. `cmd/init.go` bootstraps logging, stats client/exporter, and
  tracing exporter before the server starts.
- `pkg/server` owns the `Server` lifecycle: builds the chi router (`pkg/router`), the proxy register
  (`pkg/proxy`), the admin/API management HTTP server (`pkg/web`), and the API definition loader
  (`pkg/loader`), then emits `plugin.StartupEvent` / `plugin.AdminAPIStartupEvent` once everything is
  wired up.
- **Storage backend is pluggable via the `DATABASE_DSN` scheme** (`pkg/api/repository.go:BuildRepository`):
  `mongodb://…` → `MongoRepository`, `cassandra://…` → `CassandraRepository`, `file://…` → a directory
  of JSON API definitions. Whichever one is active is exposed to plugins on `plugin.OnStartup`
  (`MongoDB *mongo.Database` or `Cassandra wrapper.Holder`) so a plugin can share the same connection
  instead of opening its own.
- Cassandra session handling lives in `cassandra/` (`cassandra/session.go`, `cassandra/wrapper/*`):
  connection params come from env vars (`CASSANDRA_USERNAME`, `CASSANDRA_PASSWORD`, `CASSANDRA_SSL_CERT`,
  `CLUSTER_CONSISTENCY`, `CASSANDRA_SCHEMA_PATH`/`CASSANDRA_SCHEMA_FILE_NAME`), and keyspace/table
  creation runs from `cassandra/schema.sql` on startup if present.
- **Plugin system** (`pkg/plugin/plugin.go`): plugins self-register via `init()` with
  `plugin.RegisterPlugin(name, Plugin{Action, Validate})`, where `Action` is invoked per-route with the
  route's JSON config (`plugin.Decode`) to attach middleware to a `proxy.RouterDefinition`. Cross-cutting
  setup (DB handles, admin router) is delivered via `plugin.RegisterEventHook` on `StartupEvent` /
  `AdminAPIStartupEvent`. Existing plugins live under `pkg/plugin/<name>/` and must be blank-imported in
  `cmd/server.go` to register. `pkg/plugin/organization` is the newest plugin (Cassandra-only; Mongo and
  in-memory backends are stubbed but `unimplemented`) — see `docs/plugins/organization_auth.md`.
- Auth providers for the admin API register dynamically the same way, under `pkg/jwt/<provider>/`
  (currently `basic`, `github`), blank-imported in `cmd/server.go`.
- BDD tests: `features/*.feature` + `features/bootstrap/*.go`, run through `main_test.go`
  (`InitializeTestSuite`/`InitializeScenario`) via godog — this is the only place both a primary and a
  secondary Janus instance run side by side (for hot-reload / clustering scenarios).

## Service Map

- No `docs/service-map.md` exists yet; run the `project-docs` skill to generate one, including
  downstream dependents (not derivable from this repo alone).
- **Upstream dependencies** (from `pkg/config/specification.go` / `DATABASE_DSN`): one of MongoDB,
  Cassandra, or a local file tree for API definitions; optional Jaeger (`tracing.jaeger.*`) for traces;
  optional Statsd/Prometheus/Stackdriver (`stats.*`) for metrics; optional GitHub OAuth org/team for
  admin-API auth (`web.credentials.github`).
- **Deployment**: this fork does **not** deploy through `whalebone/k8s-wb`. `.circleci/config.yml`
  builds and pushes a `motivlabs/janus` image to Docker Hub on every push to `master`, then
  `.circleci/update-impulse.sh` patches the image tag into two external repos:
  `motiv-labs/impulse` (`docker-tools/compose-template/config/astro.conf`, docker-compose based) and
  `motiv-labs/impulse-googlecloud` (`impulse/janus-deployment.yaml`, k8s manifest). The `janus/` Helm
  chart in this repo and `.github/workflows/release.yml` (GoReleaser, `hellofreshtech` Docker Hub) are
  leftovers from the upstream `hellofresh/janus` project and are not part of the active deploy path.
- When you add or remove a dependency — an env var pointing at another service, a new client, a new
  queue subject — or the deployment path changes, update this Service Map section (or
  `docs/service-map.md` if it's created later) in the same change.

## Documentation

`docs/` is the upstream GitBook-style manual (see `docs/SUMMARY.md` for the full index) — mostly
generic Janus usage docs, not Whalebone-specific. Notable pages:
- `docs/plugins/` — one page per plugin, including `organization_auth.md` for the Cassandra-backed
  organization plugin.
- `docs/install/configuration.md` — full config reference (mirrors `janus.sample.toml`).
- `docs/clustering/clustering.md` — hot-reload/clustering behavior exercised by `features/HotReload.feature`.

## Gotchas

- Adding a plugin or auth provider requires two edits: the package under `pkg/plugin/<name>/` (or
  `pkg/jwt/<name>/`) with its own `init()`, **and** a blank import added to `cmd/server.go` — otherwise
  it silently never registers.
- `pkg/plugin/plugin.Decode` marshals-then-unmarshals JSON instead of using `mapstructure.Decode`
  directly, worked around because `mapstructure` was producing empty arrays for plugin config fields
  (see the `FIXME` in `pkg/plugin/plugin.go`) — don't "simplify" this back to `mapstructure.Decode`.
- `pkg/api.CassandraRepository` and `cassandra/wrapper` both call `wrapper.Initialize`/`New` — Cassandra
  connection setup runs once from the main repository build and is then handed to plugins via
  `plugin.OnStartup.Cassandra`; don't open a second, independent Cassandra session in a new plugin.
- `.circleci/config.yml` currently has the unit-test step commented out — CI-on-push image builds do
  **not** re-run `make all`; correctness is enforced solely by the separate GitHub Actions workflow.
- The commit message `#nosmoke` tag has meaning in `.circleci/update-impulse.sh`: it changes which
  commit message is written back to the `motiv-labs/impulse` repo and skips its CI (`[skip ci]`).
- `CODEOWNERS` still lists `@hellofresh/janus-owners` — a leftover from the upstream repo, not the
  current Whalebone/MotivLabs ownership.

## Team

- Shape: full-stack
- Slice terminus: UI
