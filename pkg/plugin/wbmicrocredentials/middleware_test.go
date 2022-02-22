package wbmicrocredentials

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/hellofresh/janus/pkg/test"
	"github.com/stretchr/testify/assert"
)

func TestAuthorizedAccess(t *testing.T) {
	const (
		accessKeyHeader = "Wb-Access-Key"
		secretKeyHeader = "Wb-Secret-Key"
		clientIDHeader  = "Wb-Client-Id"
		userIDHeader    = "Wb-User-Id"

		accessKey = "accessKey"
		secretKey = "secretKey"
		clientID  = "clientId"
		userID    = "userId"
	)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)
		require.Equal(t, accessKey, body["access_key"])
		require.Equal(t, secretKey, body["secret_key"])
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"user_id": "%s","client_id":"%s"}`, userID, clientID)))
	}))
	defer ts.Close()

	mw := NewWBMicroCredAuth(
		wbClient(ts.URL),
		NewCredentialsCache(time.Minute, time.Minute),
		accessKeyHeader,
		secretKeyHeader,
		clientIDHeader,
		userIDHeader,
	)

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(accessKeyHeader, accessKey)
	req.Header.Set(secretKeyHeader, secretKey)
	w := httptest.NewRecorder()
	mw(http.HandlerFunc(test.Ping)).ServeHTTP(w, req)

	// headers were added into the request
	assert.Equal(t, clientID, req.Header.Get(clientIDHeader))
	assert.Equal(t, userID, req.Header.Get(userIDHeader))
	// access key remained in the request (for tracking purposes)
	assert.Equal(t, accessKey, req.Header.Get(accessKeyHeader))
	// secret key removed from the request (from security reasons)
	assert.Empty(t, req.Header.Get(secretKeyHeader))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestAuthorizedAccessUsingCache(t *testing.T) {
	const (
		accessKeyHeader = "Access-Key"
		secretKeyHeader = "Secret-Key"
		clientIDHeader  = "Client-Id"
		userIDHeader    = "User-Id"

		accessKey = "access-key"
		secretKey = "secret-key"
		clientID  = "client-id"
		userID    = "user-id"
	)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Middleware is not supposed to call the Login endpoint when credentials are cached")
	}))
	defer ts.Close()

	c := NewCredentialsCache(time.Minute, time.Minute)
	c.Put(hashCredentials(accessKey, secretKey), NewCachedCredentials(clientID, userID, true))

	mw := NewWBMicroCredAuth(
		wbClient(ts.URL),
		c,
		accessKeyHeader,
		secretKeyHeader,
		clientIDHeader,
		userIDHeader,
	)

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(accessKeyHeader, accessKey)
	req.Header.Set(secretKeyHeader, secretKey)
	w := httptest.NewRecorder()
	mw(http.HandlerFunc(test.Ping)).ServeHTTP(w, req)

	// the client id and user id header should be injected into the request
	assert.Equal(t, clientID, req.Header.Get(clientIDHeader))
	assert.Equal(t, userID, req.Header.Get(userIDHeader))
	// access key remained in the request (for tracking purposes)
	assert.Equal(t, accessKey, req.Header.Get(accessKeyHeader))
	// secret key removed from the request (from security reasons)
	assert.Empty(t, req.Header.Get(secretKeyHeader))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestUnauthorizedMissingAccessKey(t *testing.T) {
	const (
		accessKeyHeader = "Access"
		secretKeyHeader = "Secret"
		clientIDHeader  = "Client"
		userIDHeader    = "User"
	)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Middleware is not supposed to call the Login endpoint when credentials not specified correctly")
	}))
	defer ts.Close()

	mw := NewWBMicroCredAuth(
		wbClient(ts.URL),
		NewCredentialsCache(time.Minute, time.Minute),
		accessKeyHeader,
		secretKeyHeader,
		clientIDHeader,
		userIDHeader,
	)

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(secretKeyHeader, "secretKey")
	w := httptest.NewRecorder()
	mw(http.HandlerFunc(test.Ping)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestUnauthorizedMissingSecretKey(t *testing.T) {
	const (
		accessKeyHeader = "Access"
		secretKeyHeader = "Secret"
		clientIDHeader  = "Client"
		userIDHeader    = "User"

		accessKey = "access"
	)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Middleware is not supposed to call the Login endpoint when credentials not specified correctly")
	}))
	defer ts.Close()

	mw := NewWBMicroCredAuth(
		wbClient(ts.URL),
		NewCredentialsCache(time.Minute, time.Minute),
		accessKeyHeader,
		secretKeyHeader,
		clientIDHeader,
		userIDHeader,
	)

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(accessKeyHeader, accessKey)
	w := httptest.NewRecorder()
	mw(http.HandlerFunc(test.Ping)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestUnauthorizedMissingCredentialKeys(t *testing.T) {
	const (
		accessKeyHeader = "Access"
		secretKeyHeader = "Secret"
		clientIDHeader  = "Client"
		userIDHeader    = "User"
	)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Middleware is not supposed to call the Login endpoint when credentials not specified correctly")
	}))
	defer ts.Close()

	mw := NewWBMicroCredAuth(
		wbClient(ts.URL),
		NewCredentialsCache(time.Minute, time.Minute),
		accessKeyHeader,
		secretKeyHeader,
		clientIDHeader,
		userIDHeader,
	)

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mw(http.HandlerFunc(test.Ping)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestUnauthorizedWrongCredentials(t *testing.T) {
	const (
		accessKeyHeader = "Access-Key"
		secretKeyHeader = "Secret-Key"
		clientIDHeader  = "Client-Id"
		userIDHeader    = "User-Id"

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
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error_code": 21,"error_message": "Invalid credentials"}`))
	}))
	defer ts.Close()

	mw := NewWBMicroCredAuth(
		wbClient(ts.URL),
		NewCredentialsCache(time.Minute, time.Minute),
		accessKeyHeader,
		secretKeyHeader,
		clientIDHeader,
		userIDHeader,
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

func TestUnauthorizedWrongCredentialsCached(t *testing.T) {
	const (
		accessKeyHeader = "Access-Key"
		secretKeyHeader = "Secret-Key"
		clientIDHeader  = "Client-Id"
		userIDHeader    = "User-Id"

		accessKey = "access-key"
		secretKey = "secret-key"
	)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Middleware is not supposed to call the Login endpoint when credentials cached")
	}))
	defer ts.Close()

	c := NewCredentialsCache(time.Minute, time.Minute)
	c.Put(hashCredentials(accessKey, secretKey), NewCachedCredentials("", "", false))

	mw := NewWBMicroCredAuth(
		wbClient(ts.URL),
		c,
		accessKeyHeader,
		secretKeyHeader,
		clientIDHeader,
		userIDHeader,
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
		accessKeyHeader = "Access-Key"
		secretKeyHeader = "Secret-Key"
		clientIDHeader  = "Client-Id"
		userIDHeader    = "User-Id"

		accessKey = "access-key"
		secretKey = "secret-key"
		clientID  = "client"
		userID    = "user"
	)

	// cache with 1s expiration
	c := NewCredentialsCache(time.Second, time.Minute)
	c.Put(hashCredentials(accessKey, secretKey), NewCachedCredentials("", "", false))

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)
		require.Equal(t, accessKey, body["access_key"])
		require.Equal(t, secretKey, body["secret_key"])
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"user_id":"%s","client_id":"%s"}`, userID, clientID)))
	}))
	defer ts.Close()

	mw := NewWBMicroCredAuth(
		wbClient(ts.URL),
		c,
		accessKeyHeader,
		secretKeyHeader,
		clientIDHeader,
		userIDHeader,
	)

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(accessKeyHeader, accessKey)
	req.Header.Set(secretKeyHeader, secretKey)
	wFail := httptest.NewRecorder()
	mw(http.HandlerFunc(test.Ping)).ServeHTTP(wFail, req)

	// first call must be rejected as unauthorized due to cached credentials
	assert.Equal(t, http.StatusUnauthorized, wFail.Code)
	assert.Equal(t, "application/json", wFail.Header().Get("Content-Type"))

	// let the cached credentials expire
	time.Sleep(time.Second)

	// second call shoud be ok accepted because the call to Login endpoint returns OK
	wSuccess := httptest.NewRecorder()
	mw(http.HandlerFunc(test.Ping)).ServeHTTP(wSuccess, req)
	// the headers should be injected into the request
	assert.Equal(t, clientID, req.Header.Get(clientIDHeader))
	assert.Equal(t, userID, req.Header.Get(userIDHeader))
	// access key remained in the request (for tracking purposes)
	assert.Equal(t, accessKey, req.Header.Get(accessKeyHeader))
	// secret key removed from the request (from security reasons)
	assert.Empty(t, req.Header.Get(secretKeyHeader))
	assert.Equal(t, http.StatusOK, wSuccess.Code)
	assert.Equal(t, "application/json", wSuccess.Header().Get("Content-Type"))
}

func TestLoginEndpointReturnsNoClientId(t *testing.T) {
	const (
		accessKeyHeader = "Access-Key"
		secretKeyHeader = "Secret-Key"
		clientIDHeader  = "Client-Id"
		userIDHeader    = "User-Id"

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
		w.Write([]byte(`{"user_id": "user"}`))
	}))
	defer ts.Close()

	mw := NewWBMicroCredAuth(
		wbClient(ts.URL),
		NewCredentialsCache(time.Minute, time.Minute),
		accessKeyHeader,
		secretKeyHeader,
		clientIDHeader,
		userIDHeader,
	)

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(accessKeyHeader, accessKey)
	req.Header.Set(secretKeyHeader, secretKey)
	w := httptest.NewRecorder()
	mw(http.HandlerFunc(test.Ping)).ServeHTTP(w, req)

	// since client id is not known the request must be rejected even if the credentials were ok
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestLoginEndpointReturnsNoUserId(t *testing.T) {
	const (
		accessKeyHeader = "accessKey"
		secretKeyHeader = "secretKey"
		clientIDHeader  = "clientId"
		userIDHeader    = "userId"

		accessKey = "accessKey"
		secretKey = "secretKey"
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
		w.Write([]byte(`{"client_id": "client"}`))
	}))
	defer ts.Close()

	mw := NewWBMicroCredAuth(
		wbClient(ts.URL),
		NewCredentialsCache(time.Minute, time.Minute),
		accessKeyHeader,
		secretKeyHeader,
		clientIDHeader,
		userIDHeader,
	)

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(accessKeyHeader, accessKey)
	req.Header.Set(secretKeyHeader, secretKey)
	w := httptest.NewRecorder()
	mw(http.HandlerFunc(test.Ping)).ServeHTTP(w, req)

	// since client id is not known the request must be rejected even if the credentials were ok
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestLoginEndpointUnreachable(t *testing.T) {
	const (
		accessKeyHeader = "accessKeyHeader"
		secretKeyHeader = "secretKeyHeader"
		clientIDHeader  = "clientIdHeader"
		userIDHeader    = "userIdHeader"

		accessKey = "accessKey"
		secretKey = "secretKey"
	)

	mw := NewWBMicroCredAuth(
		wbClient("http://enpoint:8080/doesnt/exits"),
		NewCredentialsCache(time.Minute, time.Minute),
		accessKeyHeader,
		secretKeyHeader,
		clientIDHeader,
		userIDHeader,
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
		accessKeyHeader = "Wb-Access-Key"
		secretKeyHeader = "Wb-Secret-Key"
		clientIDHeader  = "Wb-Client-Id"
		userIDHeader    = "Wb-User-Id"

		accessKey = "accessKey"
		secretKey = "secretKey"
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

	mw := NewWBMicroCredAuth(
		wbClient(ts.URL),
		NewCredentialsCache(time.Minute, time.Minute),
		accessKeyHeader,
		secretKeyHeader,
		clientIDHeader,
		userIDHeader,
	)

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(accessKeyHeader, accessKey)
	req.Header.Set(secretKeyHeader, secretKey)
	w := httptest.NewRecorder()
	mw(http.HandlerFunc(test.Ping)).ServeHTTP(w, req)

	// any code but OK(200) from Login endpoint must result in Unauthorized status
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func wbClient(loginEndpoint string) *WBMicroCredClient {
	return &WBMicroCredClient{LoginEndpoint: loginEndpoint}
}
