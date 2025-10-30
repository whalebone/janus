package wbapicredentials

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/hellofresh/janus/pkg/observability/otel"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// LoginRequestBody is json struct for micro credentials login request body representation
type LoginRequestBody struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
}

// LoginResponseBody is returned from call of Login endpoint service after login call
type LoginResponseBody struct {
	AccessToken string `json:"access_token"`
}

// WBAPICredClient connects to WB micro credentials service
type WBAPICredClient struct {
	LoginEndpoint string
}

// Login calls WB micro credentials service Login endpoint
func (wbClient *WBAPICredClient) Login(ctx context.Context, wbAccessKey, wbSecretKey string) (token string, success bool, err error) {
	loginReqBody, err := json.Marshal(&LoginRequestBody{AccessKey: wbAccessKey, SecretKey: wbSecretKey})
	if err != nil {
		log.WithError(err).Error("Cannot marshall access key and secret key to json")
		return "", false, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, wbClient.LoginEndpoint, bytes.NewBuffer(loginReqBody))
	if err != nil {
		log.WithError(err).Error("Cannot create login request")
		return "", false, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Transport: &otel.RequestIDPropagatingTransport{
			RoundTripper: otelhttp.NewTransport(http.DefaultTransport),
		},
	}
	response, err := client.Do(req)
	if err != nil {
		log.WithError(err).Error("Cannot perform login request")
		return "", false, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		log.Debugf("Login endpoint returned %d status code", response.StatusCode)
		return "", false, nil
	}
	var resp LoginResponseBody
	err = json.NewDecoder(response.Body).Decode(&resp)
	if err != nil {
		log.WithError(err).Errorf("Cannot parse login response")
		return "", false, nil
	}
	if resp.AccessToken == "" {
		log.Errorf("Login endpoint returned empty access token")
		return "", false, nil
	}

	return resp.AccessToken, true, nil
}
