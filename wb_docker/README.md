### Build
`docker build -f wb_docker/Dockerfile -t harbor.whalebone.io/whalebone/janus:<version> .`

**Note:** build must be run from project root, not from `wb_docker` folder

### Docker container configuration using env properties

##### Janus basic config (values are used in [janus.toml](janus.toml))
**Note**: see [janus.sample.toml](../janus.sample.toml) for specification

- `HTTP_PORT` - public port for incomming http requests (optional, default 8080)
- `ADMIN_HTTP_PORT` - admin API port (should'nt be published to the public; optional, default 8081)
- `LOG_LEVEL` - possible values: panic, fatal, error, warn (warning), info, debug; (optional, default info)
- `ADMIN_JWT_SECRET` - secret password for JWT tokens encryption for admin API
- `ADMIN_BASIC_PASS` - password for admin basic auth to admin API

##### Distributed Tracing Configuration (OpenTelemetry/OpenCensus)

- `TRACING_EXPORTER` - tracing backend: "otlp" for OpenTelemetry (recommended), "jaeger" for legacy Jaeger, or empty to disable (optional, default "")
- `TRACING_SERVICE_NAME` - service name in traces (optional, default "janus")
- `TRACING_SAMPLING_PARAM` - sampling rate: 1.0 = 100%, 0.1 = 10%, etc. (optional, default 1.0)

**When TRACING_EXPORTER="otlp" (OpenTelemetry):**

- `TRACING_OTLP_ENDPOINT` - OTLP endpoint, e.g., "otel-collector:4317" for gRPC or "otel-collector:4318" for HTTP (mandatory when using otlp)
- `TRACING_OTLP_PROTOCOL` - protocol: "grpc" or "http" (optional, default "grpc")
- `TRACING_OTLP_INSECURE` - use insecure connection (no TLS): "true" or "false" (optional, default "true")

**Standard OpenTelemetry environment variables are also supported:**
- `OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_SERVICE_NAME`, `OTEL_TRACES_SAMPLER_ARG`, `OTEL_RESOURCE_ATTRIBUTES`, etc.

##### Metrics Configuration (Prometheus)

- `STATS_EXPORTER` - metrics backend: "prometheus" to enable Prometheus metrics, or empty to disable (optional, default "")

**When STATS_EXPORTER="prometheus":**
- Metrics are exposed at `http://<ADMIN_HTTP_PORT>/metrics`
- Available metrics include:
  - `http_server_response_count_by_path_code_and_method` - request counts by path, status code, and method
  - `http_server_request_latency_by_path_and_method` - request latency histograms
  - `http_server_request_size` - request size distribution
  - `http_proxy_request_count_by_path` - upstream request counts
  - `http_proxy_request_latency_by_path` - upstream request latency
  - `plugin_oauth2_*` - OAuth2 authentication metrics
  - `plugin_jwt_manager_validation_error_total` - JWT validation errors

##### Whalebone API upstream endpoints configuration (values used in [api\_template.json](api\_template.json))

`i` - is whole number (0, 1, 2...) that allows to specify more WB api upstreams each

- `WB_API_<i>` - the name of the upstream WB api (no spaces) (mandatory)
- `WB_API_<i>_ENABLED` - enable or disable requests forwarding to upstream WB api (optional; true or false; default true)
- `WB_API_<i>_PRESERVE_HOST` - see [preserve_host property spec](../docs/proxy/preserve_host_property.md) (optional; true or false; default true)
- `WB_API_<i>_LISTEN_PATH` - requests targeting this path would be forwarded to upstream WB api (see [request uri spec](../docs/proxy/request_uri.md))
- `WB_API_<i>_UPSTREAM_TARGET` - upstream WB api base url
- `WB_API_<i>_STRIP_PATH` - see [strip_path property spec](../docs/proxy/strip_uri_property.md) (optional; true or false; default true)
- `WB_API_<i>_APPEND_PATH` - see [append_path property spec](../docs/proxy/append_uri_property.md) (optional; true or false; default true)
- `WB_API_<i>_HTTP_METHODS` - see [methods property spec](../docs/proxy/request_http_method.md) request http methods must be comma separted, each method must be quoted, eg "GET", "POST" (optional; by default: "GET", "POST", "PUT", "DELETE")

##### Upstream WB api rate limiting configuration

- `WB_API_<i>_RATE_LIMIT_ENABLED` - rate limiting enabled or disabled for WB api (optional; true or false; default true)
- `WB_API_<i>_RATE_LIMIT_VALUE` - see [rate limit property](../docs/plugins/rate_limit.md)

##### Upstream WB api requests authentications using wb auth services

- `WB_API_<i>_WB_AUTH_PLUGIN` - authentication plugin to be used for auth process (mandatory, values: wb_micro_credentials_auth, wb_api_credentials_auth)
- `WB_API_<i>_WB_AUTH_ENABLED` - authentication enabled or disabled (optional; true or false; default true)
- `WB_API_<i>_WB_AUTH_LOGIN_ENDPOINT` - URL of the service login POST endpoint (mandatory)
- `WB_API_<i>_WB_AUTH_CACHE_TTL_SECS` - cache expiration interval in seconds (if set to 0 cache will not be used, default is 30s)
- `WB_API_<i>_WB_AUTH_CACHE_CLEANUP_SECS` - expired cached records cleanup interval in secods (default is 60s)
- `WB_API_<i>_WB_AUTH_ACCESS_KEY_HEADER` - name of the request header where access key is expected (default is Wb-Access-Key)
- `WB_API_<i>_WB_AUTH_SECRET_KEY_HEADER` - name of the request header where secret key is expected (default is Wb-Secret-Key)
- `WB_API_<i>_WB_AUTH_CLIENT_ID_HEADER` - name of the request header where client id will be injected to (default is Wb-Client-Id, valid option for wb_micro_credentials_auth plugin only)
- `WB_API_<i>_WB_AUTH_USER_ID_HEADER` name of the request header where user id will be injected to (default is Wb-User-Id, valid option for wb_micro_credentials_auth plugin only)
- `WB_API_<i>_WB_AUTH_TOKEN_HEADER` - name of header where auth token will be injected (default is Authorization, valid option for wb_api_credentials_auth plugin only)
