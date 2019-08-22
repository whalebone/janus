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

##### Public-api upstream endpoint configuration (values used in [public\_api.json](public\_api.json))

- `PUBLIC_API_ENABLED` - enable or disable requests forwarding to public-api service (optional; true or false; default true)
- `PUBLIC_API_PRESERVE_HOST` - see [preserve_host property spec](../docs/proxy/preserve_host_property.md) (optional; true or false; default true)
- `PUBLIC_API_LISTEN_PATH` - requests targeting this path would be forwarded to publi-api service (see [request uri spec](../docs/proxy/request_uri.md))
- `PUBLIC_API_UPSTREAM_TARGET` - pulblic-api base url
- `PUBLIC_API_STRIP_PATH` - see [strip_path property spec](../docs/proxy/strip_uri_property.md) (optional; true or false; default true)
- `PUBLIC_API_APPEND_PATH` - see [append_path property spec](../docs/proxy/append_uri_property.md) (optional; true or false; default true)

##### Public-api rate limiting configuration

- `PUBLIC_API_RATE_LIMIT_ENABLED` - rate limiting enabled or disabled for public-api service (optional; true or false; default true)
- `PUBLIC_API_RATE_LIMIT_VALUE` - see 

##### Public-api requests authentication using micro credentials service configuration

- `PUBLIC_API_WB_AUTH_ENABLED` - authentication enabled or disabled for public-api service (optional; true or false; default true)
- `PUBLIC_API_WB_AUTH_LOGIN_ENDPOINT` - URL of microCredentials [Login](https://app.swaggerhub.com/apis-docs/whalebone/microCredentials/1.0.0#/Credentials/post_login) endpoint
- `PUBLIC_API_WB_AUTH_CACHE_TTL_SECS` - cache expiration interval in seconds (if set to 0 cache will not be used)
- `PUBLIC_API_WB_AUTH_CACHE_CLEANUP_SECS` - expired cached records cleanup interval in secods
