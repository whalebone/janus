package wbapicredentials

import (
	"errors"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/hellofresh/janus/pkg/plugin"
	"github.com/hellofresh/janus/pkg/proxy"
)

const (
	pluginName = "wb_api_credentials_auth"

	accessKeyHeaderDefault = "Wb-Access-Key"
	secretKeyHeaderDefault = "Wb-Secret-Key"

	tokenHeaderDefault = "Authorization"
)

// Config represents a rate limit config
type Config struct {
	LoginEndpoint            string `valid:"url" json:"login_endpoint"`
	CacheTTLSecs             int    `json:"cache_ttl_secs"`
	CacheCleanupIntervalSecs int    `json:"cache_cleanup_secs"`
	AccessKeyHeader          string `json:"access_key_header"`
	SecretKeyHeader          string `json:"secret_key_header"`
	TokenHeader              string `json:"token_header"`
}

func init() {
	plugin.RegisterPlugin(pluginName, plugin.Plugin{
		Action:   setup,
		Validate: validateConfig,
	})
}

func setup(def *proxy.RouterDefinition, rawConfig plugin.Config) error {
	var config Config
	err := plugin.Decode(rawConfig, &config)
	if err != nil {
		return err
	}

	if config.AccessKeyHeader == "" {
		config.AccessKeyHeader = accessKeyHeaderDefault
	}
	if config.SecretKeyHeader == "" {
		config.SecretKeyHeader = secretKeyHeaderDefault
	}
	if config.TokenHeader == "" {
		config.TokenHeader = tokenHeaderDefault
	}

	var cache *Cache
	if config.CacheTTLSecs != 0 {
		cache = NewCache(time.Duration(config.CacheTTLSecs)*time.Second,
			time.Duration(config.CacheCleanupIntervalSecs)*time.Second)
	}

	client := &WBAPICredClient{LoginEndpoint: config.LoginEndpoint}
	def.AddMiddleware(NewWBAPICredAuth(
		client,
		cache,
		config.AccessKeyHeader,
		config.SecretKeyHeader,
		config.TokenHeader,
	))
	return nil
}

func validateConfig(rawConfig plugin.Config) (bool, error) {
	var config Config
	err := plugin.Decode(rawConfig, &config)
	if err != nil {
		return false, err
	}
	if strings.TrimSpace(config.LoginEndpoint) == "" {
		return false, errors.New("login_endpoint is missing")
	}
	if config.CacheTTLSecs < 0 {
		return false, errors.New("cache_ttl_secs must be greater than or equal 0")
	}
	if config.CacheTTLSecs > 0 && config.CacheCleanupIntervalSecs <= 0 {
		return false, errors.New("cache_cleanup_secs must be specified and greater than 0")
	}
	if strings.TrimSpace(config.AccessKeyHeader) == "" {
		return false, errors.New("access_key_header must be set")
	}
	if strings.TrimSpace(config.SecretKeyHeader) == "" {
		return false, errors.New("secret_key_header must be set")
	}
	if strings.TrimSpace(config.TokenHeader) == "" {
		return false, errors.New("token_header must be set")
	}

	return govalidator.ValidateStruct(config)
}
