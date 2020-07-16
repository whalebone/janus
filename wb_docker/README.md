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

##### Upstream WB api requests authentications using micro credentials service

- `WB_API_<i>_WB_AUTH_ENABLED` - authentication enabled or disabled (optional; true or false; default true)
- `WB_API_<i>_WB_AUTH_LOGIN_ENDPOINT` - URL of microCredentials [Login](https://app.swaggerhub.com/apis-docs/whalebone/microCredentials/1.0.0#/Credentials/post_login) endpoint
- `WB_API_<i>_WB_AUTH_CACHE_TTL_SECS` - cache expiration interval in seconds (if set to 0 cache will not be used)
- `WB_API_<i>_WB_AUTH_CACHE_CLEANUP_SECS` - expired cached records cleanup interval in secods (default )
