package wbmicrocredentials

import (
	"testing"

	"github.com/hellofresh/janus/pkg/plugin"
	"github.com/hellofresh/janus/pkg/proxy"
	"github.com/stretchr/testify/require"
)

func TestSetup(t *testing.T) {
	def := proxy.NewRouterDefinition(proxy.NewDefinition())

	conf := make(plugin.Config)
	conf["test"] = "asda"

	err := setupMicroCredentials(def, conf)
	require.NoError(t, err)
	middleware := def.Middleware()
	require.Len(t, middleware, 1)
}

func TestValidateConfig(t *testing.T) {
	conf := make(plugin.Config)
	conf["login_endpoint"] = "http://endpoint:8080/path/to/login"

	valid, err := validateConfig(conf)
	require.NoError(t, err)
	require.True(t, valid)
}

func TestValidateConfigMissingLoginEndpoint(t *testing.T) {
	conf := make(plugin.Config)

	valid, err := validateConfig(conf)
	require.False(t, valid)
	require.Error(t, err)
	require.EqualError(t, err, "login_endpoint is missing")
}

func TestValidateConfigInvalidCacheTTL(t *testing.T) {
	conf := make(plugin.Config)
	conf["login_endpoint"] = "http://endpoint:8080/path/to/login"
	conf["cache_ttl_secs"] = -1

	valid, err := validateConfig(conf)
	require.False(t, valid)
	require.Error(t, err)
	require.EqualError(t, err, "cache_ttl_secs must be greater than or equal 0")
}

func TestValidateConfigMissingCacheCleanupInterval(t *testing.T) {
	conf := make(plugin.Config)
	conf["login_endpoint"] = "http://endpoint:8080/path/to/login"
	conf["cache_ttl_secs"] = 1

	valid, err := validateConfig(conf)
	require.False(t, valid)
	require.Error(t, err)
	require.EqualError(t, err, "cache_cleanup_secs must be specified and greater than 0")
}

func TestValidateConfigInvalidCacheCleanupInterval(t *testing.T) {
	conf := make(plugin.Config)
	conf["login_endpoint"] = "http://endpoint:8080/path/to/login"
	conf["cache_ttl_secs"] = 1
	conf["cache_cleanup_secs"] = -1

	valid, err := validateConfig(conf)
	require.False(t, valid)
	require.Error(t, err)
	require.EqualError(t, err, "cache_cleanup_secs must be specified and greater than 0")
}
