# AGENTS.md

This file provides guidance to AI coding agents working with code in this repository.

## ⚠️ Branch: work on `wb-micro-cred-auth`, not `master`

`master` is behind and not what's in production. All active Whalebone development and the actual
production deployment lineage live on **`origin/wb-micro-cred-auth`** (currently 41 commits ahead of
`master`, most recent Nov 2025). Its tags (e.g. `4.0.0-wb-1.0.6`) are what's built and pushed to
`harbor.whalebone.io/whalebone/janus` and deployed in `k8s-wb` (see Service Map). Before starting any
work here, check out or branch from `wb-micro-cred-auth`, not `master`. If you're reading this from a
`master` checkout, verify with `git log origin/wb-micro-cred-auth --oneline -5` that this is still
true before trusting it — branches get merged/renamed over time.

The rest of this document describes the codebase as it exists on `wb-micro-cred-auth`.

## Service identity

Janus is an HTTP API gateway (Whalebone's fork of [hellofresh/janus](https://github.com/hellofresh/janus)). It sits
in front of upstream Whalebone services, handling routing, auth, rate limiting, CORS, retries/circuit breaking, and
observability, so upstream services don't reimplement that infrastructure. It's deployed multiple times under
different names for different upstream systems (see Service Map).

Whalebone additions on top of upstream Janus:
- `pkg/plugin/wbmicrocredentials` and `pkg/plugin/wbapicredentials` — auth plugins that authenticate requests
  against Whalebone's credentials microservices (see `docs/plugins/wb_micro_credentials_auth.md`).
- `wb_docker/` — Whalebone's docker image build and env-var-driven config generation (`wb_docker/README.md`).
- `janus/` — a Helm chart for the WB deployment.

## Commands

```sh
make build            # CGO_ENABLED=0 go build -> dist/janus
make test-unit         # go test ./...
make test-integration   # requires _mocks; go test -tags=integration -cover ./...
make test-features      # builds the binary, starts _mocks, runs godog features via build/features.sh
make lint               # runs golangci-lint (gofmt + goimports only) in a docker container
```

- Run a single unit test: `go test ./pkg/plugin/wbmicrocredentials/... -run TestName -v`
- Integration tests are gated behind the `integration` build tag. They use a filesystem-backed
  `api.Repository` (`api.NewFileSystemRepository`, see `pkg/loader/api_loader_test.go`,
  `pkg/proxy/register_test.go`) against a WireMock stand-in for the upstream service, wired via
  `DYNAMIC_UPSTREAMS_PORT` (see `build/mocks.sh`). They do **not** need Mongo or the auth-service mock —
  those two only apply to feature tests below.
- Feature tests (`features/*.feature`) use [godog](https://github.com/cucumber/godog); `make test-features` builds
  the binary first, then runs `build/features.sh`, which starts the built binary (configured with a MongoDB-backed
  repository, `DYNAMIC_MONGO_PORT`) against the mocked upstream/auth services (`DYNAMIC_UPSTREAMS_PORT`,
  `DYNAMIC_AUTH_PORT`).
- CI (`.github/workflows/testing.yml`) runs lint → unit → integration → features, in that order, on
  any push that includes a non-Markdown file change (`paths-ignore: ["**/*.md"]`) — a docs-only push
  skips these jobs entirely.

## Architecture

Startup path: `main.go` → `cmd.NewRootCmd` → `cmd.RunServerStart` (`cmd/server.go`). This builds an `api.Repository`
(the config/API-definition store) and a `server.Server` (`pkg/server/server.go`), then blocks until shutdown.

**Plugin registration is side-effect based.** `cmd/server.go` blank-imports every plugin package
(`pkg/plugin/*`) purely so their `init()` functions run and call `plugin.RegisterPlugin(name, ...)`
(`pkg/plugin/plugin.go`). Adding a new plugin requires adding its blank import there, or it will never be loaded.

**Definition → Loader → Register → Router pipeline:**
1. `api.Repository` (`pkg/api/repository.go`) loads `api.Definition`s from one of three backends chosen by DSN
   scheme: `mongodb`, `cassandra`, or `file` (`BuildRepository`). Each backend has its own repository implementation
   (`mongodb_repository.go`, `cassandra_repository.go`, `file_repository.go`).
2. `loader.APILoader.RegisterAPI` (`pkg/loader/api_loader.go`) validates each definition, builds a
   `proxy.RouterDefinition`, and for every configured plugin entry looks up its `SetupFunc` via
   `plugin.DirectiveAction(name)` and calls it with the definition and raw plugin config — this is how a plugin
   attaches middleware to a specific route.
3. `proxy.Register` (`pkg/proxy/register.go`) wires the resulting router definitions into the `router.Router`
   (`pkg/router`, a chi-based router).
4. `server.Server` (`pkg/server/server.go`) owns the HTTP listener, the admin web API (`pkg/web`), and hot-reload:
   configuration changes arrive via `configurationChan`/`webServer.ConfigurationChan`, and `handleEvent` rebuilds the
   router and re-runs `APILoader.RegisterAPIs` without a process restart.

**Plugin contract** (`pkg/plugin/plugin.go`): a plugin registers a `Plugin{Action: SetupFunc, Validate: ValidateFunc}`
under a unique name. `SetupFunc` receives the `proxy.RouterDefinition` and raw JSON-like config
(`plugin.Config = map[string]interface{}`) and attaches middleware via `def.AddMiddleware(...)`. `plugin.Decode`
round-trips the raw config through JSON to populate a typed `Config` struct (a documented workaround for
`mapstructure.Decode` producing empty arrays). Plugins also register auth providers dynamically under
`pkg/jwt/{basic,github}` and can hook startup/reload events via `plugin.RegisterEventHook` /
`plugin.EmitEvent(plugin.StartupEvent | plugin.ReloadEvent, ...)`.

**Two parallel HTTP surfaces:** the main gateway/proxy listener (`s.globalConfig.Port`, built in
`createRouter`/`listenAndServe`) and a separate admin API (`pkg/web`, `s.globalConfig.Web.Port`) used to
manage API definitions and expose health/metrics/profiling. They are started and reloaded independently.

**Observability:** the project is mid-migration from OpenCensus to OpenTelemetry — `pkg/observability` and
`pkg/observability/otel` bridge the two (see `go.opencensus.io` + `go.opentelemetry.io/otel/bridge/opencensus` in
`go.mod`). Tracing exporter (`otlp` or legacy `jaeger`) and Prometheus stats exporter are configured in
`cmd/init.go`'s `initTracingExporter`/`initStatsExporter` (invoked from `cmd/server.go`'s
`RunServerStart`), driven by `config.Specification`
(`pkg/config/specification.go`), which is populated from `janus.toml` / env vars via `envconfig`.

**Config loading is file-or-env fallback, not a merge:** `cmd/init.go`'s `initConfig` calls
`config.Load(configFile)` (reads `janus.toml`, see `janus.sample.toml` for the full annotated spec);
only if that fails does it fall back to `config.LoadEnv()` (`envconfig.Process` over
`pkg/config/specification.go`'s struct tags). If the file loads successfully, env vars for these
`Specification` fields are **not** consulted — don't document them as overriding a present config file.
The WB docker image (`wb_docker/`) sidesteps this by generating `janus.toml` and `api_template.json`
from `WB_API_<i>_*` env vars *before* Janus starts — see `wb_docker/README.md` for the full variable
reference, including which are valid only for `wb_micro_credentials_auth` vs `wb_api_credentials_auth`.

## Writing a new plugin

Follow an existing plugin under `pkg/plugin/` (e.g. `wbmicrocredentials` or `rate`) as the template:
- `setup.go`: `Config` struct with `valid:"..."` govalidator tags, an `init()` that calls
  `plugin.RegisterPlugin(name, plugin.Plugin{Action, Validate})`, a `SetupFunc` that decodes config and calls
  `def.AddMiddleware(...)`, and a `ValidateFunc`.
- `middleware.go`: the actual `http.Handler`/middleware logic.
- Register the new package's blank import in `cmd/server.go`.
- Add a config doc under `docs/plugins/`.

## Documentation

- `docs/plugins/` — meant to hold one config-reference file per plugin (`docs/plugins/wb_micro_credentials_auth.md`
  for the `wbmicrocredentials` plugin), but coverage isn't complete — e.g. `pkg/plugin/wbapicredentials` has no
  corresponding doc yet. Don't assume a doc exists just because a plugin package does; check the directory.
- `docs/proxy/` — proxy-level route properties (`listen_path`, `strip_path`, `append_path`, `preserve_host`, etc.).
- `docs/auth/`, `docs/clustering/`, `docs/config/`, `docs/install/`, `docs/quick_start/`, `docs/upgrade/` — upstream
  Janus docs (GitBook source, `docs/SUMMARY.md` is the table of contents).
- `wb_docker/README.md` — Whalebone docker image build and env-var reference (`WB_API_<i>_*`).
- `janus.sample.toml` — fully annotated example of every `janus.toml` config key.

## Service Map

No `docs/service-map.md` exists yet in this repo; the map below is a best-effort summary derived from this repo and
a local read of `whalebone/k8s-wb`. Run the `project-docs` skill to generate a full, maintained map (including
"depended on by").

- **Deployment manifests**: Janus is deployed multiple times under different app names, each pointing at a
  different upstream — there is no single fixed manifest path. Found instances in
  [`whalebone/k8s-wb`](https://github.com/whalebone/k8s-wb) under `deployments/`:
  - `deployments/central-auth/base/central-auth-janus-api.yaml` (+ `central-auth-janus-fe.yaml`)
  - `deployments/partner-portal/base/janus-api.yaml` (+ `janus-fe.yaml`)
  - `deployments/jules-janus/base/jules-janus-api.yaml`
  - Alerting: `deployments/central-monitoring/base/alert-rules/janus-5xx-prometheusrule.yaml`
- **Upstream dependencies** (per deployment, configured via env vars — see `wb_docker/README.md`): each Janus
  instance proxies to one or more upstream Whalebone APIs (`WB_API_<i>_UPSTREAM_TARGET`) and authenticates against a
  Whalebone credentials service (`WB_API_<i>_WB_AUTH_LOGIN_ENDPOINT`, via `wb_micro_credentials_auth` or
  `wb_api_credentials_auth`). Example from `central-auth`: upstream target `central-auth-api`, auth endpoint
  `central-auth-apicreds`. `DATABASE_DSN` selects the API-definition store (`file://`, `mongodb://`, `cassandra://`).
- **Downstream dependents**: not determinable from this repo alone — run `project-docs` for the full picture.

When you add or remove a dependency — an env var pointing at another service, a new client, a new queue subject —
or the deployment manifests move, update the Service Map (`docs/service-map.md` if it exists, otherwise this
section) in the same change.

## Team

- Shape: full-stack
- Slice terminus: UI
