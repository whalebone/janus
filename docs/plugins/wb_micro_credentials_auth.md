# Authentication using WB credentials micro-service

Authenticates each request against Whalebone's microCredentials service. Plugin expects user credentials in provided `Wb-Access-Key` and `Wb-Secret-Key` headers. If authentication is successful `Wb-Client-Id` and `Wb-User-Id` headers are injected into the request acording to provided credentials. In-memmory cache is used to prevent microCredentials service overload.

## Configuration

The plain rate limit config:

```json
"wb_micro_credentials_auth": {
    "enabled": true,
    "config": {
        "login_endpoint": "http://host:port/path/to/login/endpoint",
        "cache_ttl_secs": 30,
        "cache_cleanup_secs": 300
    }
}
```

| Configuration        | Description |
|----------------------|-------------|
| login_endpoint       | The URL of WB credentials micro service Login endpoint (mandatory) |
| cache_ttl_secs       | The expiration interval in seconds for credentials in-memmory cache (optional, must be greater than or equal to 0, if not specified or set to 0 the cache won't be used) |
| cache_cleanup_secs   | The cache clean up interval in seconds (mandatory if cache_ttl_secs is specified and greater than 0 then must be greater than 0, otherwise ignored) |
