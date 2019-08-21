# Authentication using WB credentials micro-service

Rate limit how many HTTP requests a developer can make in a given period of seconds, minutes, hours, days, months or years.

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
| login_endpoint       | The url of WB credentials micro service Login endpoint (mandatory) |
| cache_ttl_secs       | The the expiration interval in seconds for credentials in-memmory cache (optional, must be greater than or equal to 0, if not specified or set to 0 the cache won't be used) |
| cache_cleanup_secs   | The cache clean up interval in seconds (mandatory if cache_ttl_secs is specified and greater than 0, must be greater than 0) |
