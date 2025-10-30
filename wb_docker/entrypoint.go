package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	defaultJanusRoot  = "/opt/janus"
	defaultAPIConfDir = "/etc/janus/apis"
	defaultJanusConf  = "/etc/janus/janus.toml"
	defaultLogLevel   = "info"
)

var (
	janusRoot  = getEnv("JANUS_ROOT", defaultJanusRoot)
	janusConf  = getEnv("JANUS_CONF", defaultJanusConf)
	apiConfDir = getEnv("API_CONF_DIR", defaultAPIConfDir)
	logLevel   = getEnv("LOG_LEVEL", defaultLogLevel)

	defaultAPIConfTemplate = janusRoot + "/api_template.json"
	defaultJanusBin        = janusRoot + "/janus"

	apiConfTemplate = getEnv("API_CONF_TEMPLATE", defaultAPIConfTemplate)
	janusBin        = getEnv("JANUS_BIN", defaultJanusBin)
)

func initLog() {
	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
	})
	level, _ := log.ParseLevel(strings.ToLower(logLevel))
	log.SetLevel(level)
}

func main() {
	initLog()
	log.Debug("Preparing Janus configuration...")
	if err := prepareJanusConfiguration(janusConf); err != nil {
		log.Error("Error preparing Janus configuration: ", err)
		os.Exit(1)
	}

	log.Debug("Cleaning old API configurations...")
	if err := cleanOldAPIConfigs(apiConfDir); err != nil {
		log.Error("Error cleaning old API configurations: ", err)
		os.Exit(1)
	}

	log.Debug("Preparing new API configurations...")
	if err := prepareNewAPIConfigs(apiConfDir, apiConfTemplate); err != nil {
		log.Error("Error preparing new API configurations: ", err)
		os.Exit(1)
	}

	log.Info("Starting janus...")
	if err := startService(janusBin, janusConf); err != nil {
		log.Error("Error starting service: ", err)
		os.Exit(1)
	}
}

func prepareJanusConfiguration(janusConf string) error {
	replacements := map[string]string{
		"@HTTP_PORT@":        getEnv("HTTP_PORT", "8080"),
		"@LOG_LEVEL@":        getEnv("LOG_LEVEL", "info"),
		"@ADMIN_HTTP_PORT@":  getEnv("ADMIN_HTTP_PORT", "8081"),
		"@ADMIN_JWT_SECRET@": os.Getenv("ADMIN_JWT_SECRET"),
		"@ADMIN_BASIC_PASS@": os.Getenv("ADMIN_BASIC_PASS"),

		// Tracing configuration
		"@TRACING_EXPORTER@":       getEnv("TRACING_EXPORTER", ""),
		"@TRACING_SERVICE_NAME@":   getEnv("TRACING_SERVICE_NAME", "janus"),
		"@TRACING_SAMPLING_PARAM@": getEnv("TRACING_SAMPLING_PARAM", "1.0"),

		// OTLP configuration (when TRACING_EXPORTER=otlp)
		"@TRACING_OTLP_ENDPOINT@": getEnv("TRACING_OTLP_ENDPOINT", ""),
		"@TRACING_OTLP_PROTOCOL@": getEnv("TRACING_OTLP_PROTOCOL", "grpc"),
		"@TRACING_OTLP_INSECURE@": getEnv("TRACING_OTLP_INSECURE", "true"),
	}

	return replaceInFileMultiple(janusConf, replacements)
}

func cleanOldAPIConfigs(apiConfDir string) error {
	if err := os.RemoveAll(apiConfDir); err != nil {
		return err
	}

	return os.MkdirAll(apiConfDir, os.ModePerm)
}

func prepareNewAPIConfigs(apiConfDir string, apiConfTemplate string) error {
	for i := 0; ; i++ {
		apiEnv := fmt.Sprintf("WB_API_%d", i)
		apiName := os.Getenv(apiEnv)
		if apiName == "" {
			break
		}

		apiConf := fmt.Sprintf("%s/%s.json", apiConfDir, apiName)
		if err := prepareAPIConfig(apiConf, apiConfTemplate, apiEnv, apiName); err != nil {
			return err
		}
	}

	return nil
}

func prepareAPIConfig(apiConf string, apiConfTemplate string, apiEnv string, apiName string) error {
	if err := copyFile(apiConfTemplate, apiConf); err != nil {
		return err
	}

	replacements := map[string]string{
		"@NAME@":                       apiName,
		"@ENABLED@":                    getEnv(apiEnv+"_ENABLED", "true"),
		"@PRESERVE_HOST@":              getEnv(apiEnv+"_PRESERVE_HOST", "false"),
		"@LISTEN_PATH@":                os.Getenv(apiEnv + "_LISTEN_PATH"),
		"@UPSTREAM_TARGET@":            os.Getenv(apiEnv + "_UPSTREAM_TARGET"),
		"@STRIP_PATH@":                 getEnv(apiEnv+"_STRIP_PATH", "true"),
		"@APPEND_PATH@":                getEnv(apiEnv+"_APPEND_PATH", "true"),
		"@HTTP_METHODS@":               getEnv(apiEnv+"_HTTP_METHODS", "\"GET\",\"POST\",\"PUT\",\"DELETE\""),
		"@RATE_LIMIT_ENABLED@":         getEnv(apiEnv+"_RATE_LIMIT_ENABLED", "true"),
		"@RATE_LIMIT_VALUE@":           getEnv(apiEnv+"_RATE_LIMIT_VALUE", "5-M"),
		"@WB_AUTH_PLUGIN@":             os.Getenv(apiEnv + "_WB_AUTH_PLUGIN"),
		"@WB_AUTH_ENABLED@":            getEnv(apiEnv+"_WB_AUTH_ENABLED", "true"),
		"@WB_AUTH_LOGIN_ENDPOINT@":     os.Getenv(apiEnv + "_WB_AUTH_LOGIN_ENDPOINT"),
		"@WB_AUTH_CACHE_TTL_SECS@":     getEnv(apiEnv+"_WB_AUTH_CACHE_TTL_SECS", "30"),
		"@WB_AUTH_CACHE_CLEANUP_SECS@": getEnv(apiEnv+"_WB_AUTH_CACHE_CLEANUP_SECS", "60"),
		"@WB_AUTH_ACCESS_KEY_HEADER@":  getEnv(apiEnv+"_WB_AUTH_ACCESS_KEY_HEADER", "Wb-Access-Key"),
		"@WB_AUTH_SECRET_KEY_HEADER@":  getEnv(apiEnv+"_WB_AUTH_SECRET_KEY_HEADER", "Wb-Secret-Key"),
		"@WB_AUTH_TOKEN_HEADER@":       getEnv(apiEnv+"_WB_AUTH_TOKEN_HEADER", "Authorization"),
		"@WB_AUTH_CLIENT_ID_HEADER@":   getEnv(apiEnv+"_WB_AUTH_CLIENT_ID_HEADER", "Wb-Client-Id"),
		"@WB_AUTH_USER_ID_HEADER@":     getEnv(apiEnv+"_WB_AUTH_USER_ID_HEADER", "Wb-User-Id"),
	}

	return replaceInFileMultiple(apiConf, replacements)
}

func startService(janusBin string, janusConf string) error {
	cmd := exec.Command(janusBin, "-c", janusConf, "start")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = defaultValue
	}
	return value
}

func replaceInFile(filename, search, replace string) error {
	read, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	newContents := strings.Replace(string(read), search, replace, -1)

	return os.WriteFile(filename, []byte(newContents), 0)
}

func replaceInFileMultiple(filename string, replacements map[string]string) error {
	for search, replace := range replacements {
		if err := replaceInFile(filename, search, replace); err != nil {
			return err
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, input, 0o644)
}
