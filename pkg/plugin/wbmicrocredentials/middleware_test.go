package wbmicrocredentials

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/hellofresh/janus/pkg/test"
	"github.com/stretchr/testify/assert"
)

func TestAuthorizedAccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)
		require.Equal(t, "accessKey", body["access_key"])
		require.Equal(t, "secretKey", body["secret_key"])
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"user_id": "1","client_id":"someClinetId"}`))
	}))
	defer ts.Close()

	mw := NewWBMicroCredAuth(wbClient(ts.URL), NewCredentialsCache(time.Minute, time.Minute))

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Wb-Access-Key", "accessKey")
	req.Header.Add("Wb-Secret-Key", "secretKey")
	w := httptest.NewRecorder()
	mw(http.HandlerFunc(test.Ping)).ServeHTTP(w, req)

	assert.Equal(t, "someClinetId", req.Header.Get("WB-Client-Id"))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestAuthorizedAccessUsingCache(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Middleware is not supposed to call the Login endpoint when credentials are cached")
	}))
	defer ts.Close()

	c := NewCredentialsCache(time.Minute, time.Minute)
	c.Put(hashCredentials("accessKey", "secretKey"), &CachedCredentials{ClientID: "12345", LoginSuccess: true})

	mw := NewWBMicroCredAuth(wbClient(ts.URL), c)

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Wb-Access-Key", "accessKey")
	req.Header.Add("Wb-Secret-Key", "secretKey")
	w := httptest.NewRecorder()
	mw(http.HandlerFunc(test.Ping)).ServeHTTP(w, req)

	// the WB-Client-Id header should be injected into the request
	assert.Equal(t, "12345", req.Header.Get("WB-Client-Id"))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestUnauthorizedMissingAccessKey(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Middleware is not supposed to call the Login endpoint when credentials not specified correctly")
	}))
	defer ts.Close()

	mw := NewWBMicroCredAuth(wbClient(ts.URL), NewCredentialsCache(time.Minute, time.Minute))

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Wb-Secret-Key", "secretKey")
	w := httptest.NewRecorder()
	mw(http.HandlerFunc(test.Ping)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestUnauthorizedMissingSecretKey(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Middleware is not supposed to call the Login endpoint when credentials not specified correctly")
	}))
	defer ts.Close()

	mw := NewWBMicroCredAuth(wbClient(ts.URL), NewCredentialsCache(time.Minute, time.Minute))

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Wb-Access-Key", "accessKey")
	w := httptest.NewRecorder()
	mw(http.HandlerFunc(test.Ping)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestUnauthorizedMissingCredentialKeys(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Middleware is not supposed to call the Login endpoint when credentials not specified correctly")
	}))
	defer ts.Close()

	mw := NewWBMicroCredAuth(wbClient(ts.URL), NewCredentialsCache(time.Minute, time.Minute))

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	req.Header.Add("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mw(http.HandlerFunc(test.Ping)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestUnauthorizedWrongCredentials(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)
		require.Equal(t, "accessKey", body["access_key"])
		require.Equal(t, "secretKey", body["secret_key"])
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error_code": 21,"error_message": "Invalid credentials"}`))
	}))
	defer ts.Close()

	mw := NewWBMicroCredAuth(wbClient(ts.URL), NewCredentialsCache(time.Minute, time.Minute))

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Wb-Access-Key", "accessKey")
	req.Header.Add("Wb-Secret-Key", "secretKey")
	w := httptest.NewRecorder()
	mw(http.HandlerFunc(test.Ping)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestUnauthorizedWrongCredentialsCached(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Middleware is not supposed to call the Login endpoint when credentials cached")
	}))
	defer ts.Close()

	c := NewCredentialsCache(time.Minute, time.Minute)
	c.Put(hashCredentials("accessKey", "secretKey"), &CachedCredentials{ClientID: "", LoginSuccess: false})

	mw := NewWBMicroCredAuth(wbClient(ts.URL), c)

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Wb-Access-Key", "accessKey")
	req.Header.Add("Wb-Secret-Key", "secretKey")
	w := httptest.NewRecorder()
	mw(http.HandlerFunc(test.Ping)).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestCacheExpires(t *testing.T) {
	// cache with 1s expiration
	c := NewCredentialsCache(time.Second, time.Minute)
	c.Put(hashCredentials("accessKey", "secretKey"), &CachedCredentials{ClientID: "", LoginSuccess: false})

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)
		require.Equal(t, "accessKey", body["access_key"])
		require.Equal(t, "secretKey", body["secret_key"])
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"user_id": "1","client_id":"someClinetId"}`))
	}))
	defer ts.Close()

	mw := NewWBMicroCredAuth(wbClient(ts.URL), c)

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Wb-Access-Key", "accessKey")
	req.Header.Add("Wb-Secret-Key", "secretKey")
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
	// the WB-Client-Id header should be injected into the request
	assert.Equal(t, "someClinetId", req.Header.Get("WB-Client-Id"))
	assert.Equal(t, http.StatusOK, wSuccess.Code)
	assert.Equal(t, "application/json", wSuccess.Header().Get("Content-Type"))
}

func TestLoginEndpointReturnsNoClientId(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)
		require.Equal(t, "accessKey", body["access_key"])
		require.Equal(t, "secretKey", body["secret_key"])
		w.WriteHeader(http.StatusOK)
		// client id is missng in the response
		w.Write([]byte(`{"user_id": "1"}`))
	}))
	defer ts.Close()

	mw := NewWBMicroCredAuth(wbClient(ts.URL), NewCredentialsCache(time.Minute, time.Minute))

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Wb-Access-Key", "accessKey")
	req.Header.Add("Wb-Secret-Key", "secretKey")
	w := httptest.NewRecorder()
	mw(http.HandlerFunc(test.Ping)).ServeHTTP(w, req)

	// since client id is not known the request must be rejected even if the credentials were ok
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestLoginEndpointUnreachable(t *testing.T) {
	mw := NewWBMicroCredAuth(wbClient("http://enpoint:8080/doesnt/exits"), NewCredentialsCache(time.Minute, time.Minute))

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Wb-Access-Key", "accessKey")
	req.Header.Add("Wb-Secret-Key", "secretKey")
	w := httptest.NewRecorder()
	mw(http.HandlerFunc(test.Ping)).ServeHTTP(w, req)

	// since client id is not known the request must be rejected even if the credentials were ok
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestLoginEndpointReturns500(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)
		require.Equal(t, "accessKey", body["access_key"])
		require.Equal(t, "secretKey", body["secret_key"])
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	mw := NewWBMicroCredAuth(wbClient(ts.URL), NewCredentialsCache(time.Minute, time.Minute))

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Wb-Access-Key", "accessKey")
	req.Header.Add("Wb-Secret-Key", "secretKey")
	w := httptest.NewRecorder()
	mw(http.HandlerFunc(test.Ping)).ServeHTTP(w, req)

	// any code but OK(200) from Login endpoint must result in Unauthorized status
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func wbClient(loginEndpoint string) *WBMicroCredClient {
	return &WBMicroCredClient{LoginEndpoint: loginEndpoint}
}
