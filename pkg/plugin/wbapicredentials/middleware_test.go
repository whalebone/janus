package wbapicredentials

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hellofresh/janus/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthorizedAccess(t *testing.T) {
	const (
		accessKeyHeader = "Wb-Access-Key"
		secretKeyHeader = "Wb-Secret-Key"
		tokenHeader     = "Authorization"

		accessKey = "accessKey"
		secretKey = "secretKey"
		token     = "token"
	)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)
		require.Equal(t, accessKey, body["access_key"])
		require.Equal(t, secretKey, body["secret_key"])
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"access_token": "%s", "irrelevant": "some val"}`, token)))
	}))
	defer ts.Close()

	mw := NewWBAPICredAuth(
		wbClient(ts.URL),
		NewCache(time.Minute, time.Minute),
		accessKeyHeader,
		secretKeyHeader,
		tokenHeader,
	)

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(accessKeyHeader, accessKey)
	req.Header.Set(secretKeyHeader, secretKey)
	w := httptest.NewRecorder()
	mw(http.HandlerFunc(test.Ping)).ServeHTTP(w, req)

	// headers were added into the request
	assert.Equal(t, token, req.Header.Get(tokenHeader))
	// access key remained in the request (for tracking purposes)
	assert.Equal(t, accessKey, req.Header.Get(accessKeyHeader))
	// secret key removed from the request (from security reasons)
	assert.Empty(t, req.Header.Get(secretKeyHeader))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestUnauthorizedAccessUsingCache(t *testing.T) {
	const (
		accessKeyHeader = "Access-Key"
		secretKeyHeader = "Secret-Key"
		tokenHeader     = "Irrelevant"

		accessKey = "access-key"
		secretKey = "access-key"
	)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Middleware is not supposed to call the Login endpoint when credentials have been cached")
	}))
	defer ts.Close()

	c := NewCache(time.Minute, time.Minute)
	c.Put(hashCredentials(accessKey, secretKey))

	mw := NewWBAPICredAuth(wbClient(ts.URL), c, accessKeyHeader, secretKeyHeader, tokenHeader)

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(accessKeyHeader, accessKey)
	req.Header.Set(secretKeyHeader, secretKey)
	w := httptest.NewRecorder()
	mw(http.HandlerFunc(test.Ping)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestUnauthorizedMissingAccessKey(t *testing.T) {
	const (
		accessKeyHeader = "AccessKey"
		secretKeyHeader = "SecretKey"
		tokenHeader     = "Irrelevant"

		secretKey = "secret_key"
	)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Middleware is not supposed to call the Login endpoint when credentials not specified correctly")
	}))
	defer ts.Close()

	mw := NewWBAPICredAuth(wbClient(ts.URL), NewCache(time.Minute, time.Minute), accessKeyHeader, secretKeyHeader, tokenHeader)

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(secretKeyHeader, secretKey)
	w := httptest.NewRecorder()
	mw(http.HandlerFunc(test.Ping)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestUnauthorizedMissingSecretKey(t *testing.T) {
	const (
		accessKeyHeader = "AccessKey"
		secretKeyHeader = "SecretKey"
		tokenHeader     = "Irrelevant"

		accessKey = "access_key"
	)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Middleware is not supposed to call the Login endpoint when credentials not specified correctly")
	}))
	defer ts.Close()

	mw := NewWBAPICredAuth(wbClient(ts.URL), NewCache(time.Minute, time.Minute), accessKeyHeader, secretKeyHeader, tokenHeader)

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add(accessKeyHeader, accessKey)
	w := httptest.NewRecorder()
	mw(http.HandlerFunc(test.Ping)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestUnauthorizedMissingCredentialKeys(t *testing.T) {
	const (
		accessKeyHeader = "WBAccessKey"
		secretKeyHeader = "WBSecretKey"
		tokenHeader     = "Irrelevant"
	)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Middleware is not supposed to call the Login endpoint when credentials not specified correctly")
	}))
	defer ts.Close()

	mw := NewWBAPICredAuth(
		wbClient(ts.URL),
		NewCache(time.Minute, time.Minute),
		accessKeyHeader,
		secretKeyHeader,
		tokenHeader,
	)

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	req.Header.Add("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mw(http.HandlerFunc(test.Ping)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestUnauthorizedWrongCredentials(t *testing.T) {
	const (
		accessKeyHeader = "WBAccessKey"
		secretKeyHeader = "WBSecretKey"
		tokenHeader     = "Irrelevant"

		accessKey = "access_key"
		secretKey = "secret_key"
	)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)
		require.Equal(t, accessKey, body["access_key"])
		require.Equal(t, secretKey, body["secret_key"])
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error_code": "INVALID_CREDENTIALS", "errors": ["invalid credentials"]}`))
	}))
	defer ts.Close()

	mw := NewWBAPICredAuth(
		wbClient(ts.URL),
		NewCache(time.Minute, time.Minute),
		accessKeyHeader,
		secretKeyHeader,
		tokenHeader,
	)

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(accessKeyHeader, accessKey)
	req.Header.Set(secretKeyHeader, secretKey)
	w := httptest.NewRecorder()
	mw(http.HandlerFunc(test.Ping)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestCacheExpires(t *testing.T) {
	const (
		accessKeyHeader = "WBAccessKey"
		secretKeyHeader = "WBSecretKey"
		tokenHeader     = "WBToken"

		accessKey = "access-key"
		secretKey = "secret-key"
		token     = "authToken"
	)

	// cache with 1s expiration
	c := NewCache(time.Second, time.Minute)
	c.Put(hashCredentials(accessKey, secretKey))

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)
		require.Equal(t, accessKey, body["access_key"])
		require.Equal(t, secretKey, body["secret_key"])
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"access_token": "%s"}`, token)))
	}))
	defer ts.Close()

	mw := NewWBAPICredAuth(
		wbClient(ts.URL),
		c,
		accessKeyHeader,
		secretKeyHeader,
		tokenHeader,
	)

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(accessKeyHeader, accessKey)
	req.Header.Set(secretKeyHeader, secretKey)
	wFail := httptest.NewRecorder()
	mw(http.HandlerFunc(test.Ping)).ServeHTTP(wFail, req)

	// first call must be rejected as unauthorized due to cache
	assert.Equal(t, http.StatusUnauthorized, wFail.Code)
	assert.Equal(t, "application/json", wFail.Header().Get("Content-Type"))

	// let the cached credentials expire
	time.Sleep(time.Second)

	// second call shoud be ok accepted because the call to Login endpoint returns OK
	wSuccess := httptest.NewRecorder()
	mw(http.HandlerFunc(test.Ping)).ServeHTTP(wSuccess, req)
	// the Wb-* headers should be injected into the request
	assert.Equal(t, token, req.Header.Get(tokenHeader))
	assert.Equal(t, http.StatusOK, wSuccess.Code)
	assert.Equal(t, "application/json", wSuccess.Header().Get("Content-Type"))
}

func TestLoginEndpointReturnsNoToken(t *testing.T) {
	const (
		accessKeyHeader = "AccessKey"
		secretKeyHeader = "SecretKey"
		tokenHeader     = "Token"

		accessKey = "access-key"
		secretKey = "secret-key"
	)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)
		require.Equal(t, accessKey, body["access_key"])
		require.Equal(t, secretKey, body["secret_key"])
		w.WriteHeader(http.StatusOK)
		// client id is missng in the response
		w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	mw := NewWBAPICredAuth(
		wbClient(ts.URL),
		NewCache(time.Minute, time.Minute),
		accessKeyHeader,
		secretKeyHeader,
		tokenHeader,
	)

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(accessKeyHeader, accessKey)
	req.Header.Set(secretKeyHeader, secretKey)
	w := httptest.NewRecorder()
	mw(http.HandlerFunc(test.Ping)).ServeHTTP(w, req)

	// since token is not known the request must be rejected even if the credentials were ok
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestLoginEndpointUnreachable(t *testing.T) {
	const (
		accessKeyHeader = "AccessKey"
		secretKeyHeader = "SecretKey"
		tokenHeader     = "Token"

		accessKey = "access-key"
		secretKey = "secret-key"
	)

	mw := NewWBAPICredAuth(
		wbClient("http://enpoint:8080/doesnt/exits"),
		NewCache(time.Minute, time.Minute),
		accessKeyHeader,
		secretKeyHeader,
		tokenHeader,
	)

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add(accessKeyHeader, accessKey)
	req.Header.Add(secretKeyHeader, secretKey)
	w := httptest.NewRecorder()
	mw(http.HandlerFunc(test.Ping)).ServeHTTP(w, req)

	// since client id is not known the request must be rejected even if the credentials were ok
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestLoginEndpointReturns500(t *testing.T) {
	const (
		accessKeyHeader = "AccessKey"
		secretKeyHeader = "SecretKey"
		tokenHeader     = "Token"

		accessKey = "access-key"
		secretKey = "secret-key"
	)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)
		require.Equal(t, accessKey, body["access_key"])
		require.Equal(t, secretKey, body["secret_key"])
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	mw := NewWBAPICredAuth(
		wbClient(ts.URL),
		NewCache(time.Minute, time.Minute),
		accessKeyHeader,
		secretKeyHeader,
		tokenHeader,
	)

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add(accessKeyHeader, accessKey)
	req.Header.Add(secretKeyHeader, secretKey)
	w := httptest.NewRecorder()
	mw(http.HandlerFunc(test.Ping)).ServeHTTP(w, req)

	// any code but OK(200) from Login endpoint must result in Unauthorized status
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func wbClient(loginEndpoint string) *WBAPICredClient {
	return &WBAPICredClient{LoginEndpoint: loginEndpoint}
}
