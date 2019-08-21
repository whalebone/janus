package wbmicrocredentials

import (
	goerrors "errors"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/hellofresh/janus/pkg/plugin"
	"github.com/hellofresh/janus/pkg/proxy"
)

// Config represents a rate limit config
type Config struct {
	LoginEndpoint            string `valid:"url" json:"login_endpoint"`
	CacheTTLSecs             int    `json:"cache_ttl_secs"`
	CacheCleanupIntervalSecs int    `json:"cache_cleanup_secs"`
}

func init() {
	plugin.RegisterPlugin("wb_micro_credentials_auth", plugin.Plugin{
		Action:   setupMicroCredentials,
		Validate: validateConfig,
	})
}

func setupMicroCredentials(def *proxy.RouterDefinition, rawConfig plugin.Config) error {
	var config Config
	err := plugin.Decode(rawConfig, &config)
	if err != nil {
		return err
	}

	var credentialsCache *CredentialsCache
	if config.CacheTTLSecs != 0 {
		credentialsCache = NewCredentialsCache(time.Duration(config.CacheTTLSecs)*time.Second,
			time.Duration(config.CacheCleanupIntervalSecs)*time.Second)
	}

	client := &WBMicroCredClient{LoginEndpoint: config.LoginEndpoint}
	def.AddMiddleware(NewWBMicroCredAuth(client, credentialsCache))
	return nil
}

func validateConfig(rawConfig plugin.Config) (bool, error) {
	var config Config
	err := plugin.Decode(rawConfig, &config)
	if err != nil {
		return false, err
	}
	if config.LoginEndpoint == "" {
		return false, goerrors.New("login_endpoint is missing")
	}
	if config.CacheTTLSecs < 0 {
		return false, goerrors.New("cache_ttl_secs must be greater than or equal 0")
	}
	if config.CacheTTLSecs > 0 && config.CacheCleanupIntervalSecs <= 0 {
		return false, goerrors.New("cache_cleanup_secs must be specified and greater than 0")
	}

	return govalidator.ValidateStruct(config)
}
