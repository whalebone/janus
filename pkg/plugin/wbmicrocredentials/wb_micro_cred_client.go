package wbmicrocredentials

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/hellofresh/janus/pkg/observability/otel"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// WBCredentials is returned from call of Login endpoint service after login call
type WBCredentials struct {
	ClientID string `json:"client_id"`
	UserID   string `json:"user_id"`
}

// WBLoginRequestBody is json struct for micro credentials login request body representation
type WBLoginRequestBody struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
}

// WBMicroCredClient connects to WB micro credentials service
type WBMicroCredClient struct {
	LoginEndpoint string
}

// Login calls WB micro credentials service Login endpoint
func (wbClient *WBMicroCredClient) Login(ctx context.Context, wbAccessKey, wbSecretKey string) (clientID, userID string, success bool, err error) {
	loginReqBody, err := json.Marshal(&WBLoginRequestBody{AccessKey: wbAccessKey, SecretKey: wbSecretKey})
	if err != nil {
		log.WithError(err).Error("Cannot marshall access key and secret key to json")
		return "", "", false, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, wbClient.LoginEndpoint, strings.NewReader(string(loginReqBody)))
	if err != nil {
		log.WithError(err).Error("Cannot create login request")
		return "", "", false, err
	}
	req.Header.Set("Content-Type", "application/json")

	// Chain transports: X-Request-Id propagation -> OTel instrumentation -> default transport
	client := &http.Client{
		Transport: &otel.RequestIDPropagatingTransport{
			RoundTripper: otelhttp.NewTransport(http.DefaultTransport),
		},
	}
	response, err := client.Do(req)
	if err != nil {
		log.WithError(err).Error("Cannot perform login request")
		return "", "", false, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		log.Debugf("Login endpoint returned %d status code", response.StatusCode)
		return "", "", false, nil
	}
	var cred WBCredentials
	err = json.NewDecoder(response.Body).Decode(&cred)
	if err != nil {
		log.WithError(err).Errorf("Cannot parse login response")
		return "", "", false, nil
	}
	if cred.ClientID == "" {
		log.Errorf("Login endpoint returned credentials with no client id")
		return "", "", false, nil
	}

	if cred.UserID == "" {
		log.Errorf("Login endpoint returned credentials with no user id")
		return "", "", false, nil
	}

	return cred.ClientID, cred.UserID, true, nil
}
