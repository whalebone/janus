package wbmicrocredentials

import (
	"errors"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/hellofresh/janus/pkg/plugin"
	"github.com/hellofresh/janus/pkg/proxy"
)

const (
	pluginName = "wb_micro_credentials_auth"

	accessKeyHeaderDefault = "Wb-Access-Key"
	secretKeyHeaderDefault = "Wb-Secret-Key"
	clientIDHeaderDefault  = "Wb-Client-Id"
	userIDHeaderDefault    = "Wb-User-Id"
)

// Config represents a rate limit config
type Config struct {
	LoginEndpoint            string `valid:"url" json:"login_endpoint"`
	CacheTTLSecs             int    `json:"cache_ttl_secs"`
	CacheCleanupIntervalSecs int    `json:"cache_cleanup_secs"`
	AccessKeyHeader          string `json:"access_key_header"`
	SecretKeyHeader          string `json:"secret_key_header"`
	ClientIDHeader           string `json:"client_id_header"`
	UserIDHeader             string `json:"user_id_header"`
}

func init() {
	plugin.RegisterPlugin(pluginName, plugin.Plugin{
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

	if config.AccessKeyHeader == "" {
		config.AccessKeyHeader = accessKeyHeaderDefault
	}
	if config.SecretKeyHeader == "" {
		config.SecretKeyHeader = secretKeyHeaderDefault
	}
	if config.ClientIDHeader == "" {
		config.ClientIDHeader = clientIDHeaderDefault
	}
	if config.UserIDHeader == "" {
		config.UserIDHeader = userIDHeaderDefault
	}

	var credentialsCache *CredentialsCache
	if config.CacheTTLSecs != 0 {
		credentialsCache = NewCredentialsCache(time.Duration(config.CacheTTLSecs)*time.Second,
			time.Duration(config.CacheCleanupIntervalSecs)*time.Second)
	}

	client := &WBMicroCredClient{LoginEndpoint: config.LoginEndpoint}
	def.AddMiddleware(NewWBMicroCredAuth(
		client,
		credentialsCache,
		config.AccessKeyHeader,
		config.SecretKeyHeader,
		config.ClientIDHeader,
		config.UserIDHeader,
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
	if strings.TrimSpace(config.ClientIDHeader) == "" {
		return false, errors.New("client_id_header must be set")
	}
	if strings.TrimSpace(config.UserIDHeader) == "" {
		return false, errors.New("user_id_header must be set")
	}

	return govalidator.ValidateStruct(config)
}
