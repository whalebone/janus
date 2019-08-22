package wbmicrocredentials

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/hellofresh/janus/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	accessKeyHeader = "Wb-Access-Key"
	secretKeyHeader = "Wb-Secret-Key"
	clientIDHeader  = "Wb-Client-Id"
)

// NewWBMicroCredAuth is a HTTP basic auth middleware
func NewWBMicroCredAuth(wbClient *WBMicroCredClient, cache *CredentialsCache) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Debug("Starting wb_micro_credentials auth middleware")
			wbAccessKey := r.Header.Get(accessKeyHeader)
			wbSecretKey := r.Header.Get(secretKeyHeader)
			if wbAccessKey == "" || wbSecretKey == "" {
				errors.Handler(w, r, ErrNotAuthorized)
				return
			}
			// hashed access key and secret key to be used as cache key
			hashedCred := hashCredentials(wbAccessKey, wbSecretKey)
			// client id as returned from WB microCredentials service after login call
			var clientID string
			foundInCache := false
			// cache doesn't have to be used, if not then it is nil
			if cache != nil {
				var cachedCred *CachedCredentials
				if cachedCred, foundInCache = cache.Get(hashedCred); foundInCache {
					if !cachedCred.LoginSuccess {
						errors.Handler(w, r, ErrInvalidCredentials)
						return
					}
					clientID = cachedCred.ClientID
				}
			}
			if !foundInCache {
				var success bool
				var err error
				clientID, success, err = wbClient.Login(wbAccessKey, wbSecretKey)
				if err != nil {
					errors.Handler(w, r, ErrInvalidCredentials)
					return
				}
				if cache != nil {
					cache.Put(hashedCred, &CachedCredentials{ClientID: clientID, LoginSuccess: success})
				}
				if !success {
					errors.Handler(w, r, ErrInvalidCredentials)
					return
				}
			}
			r.Header.Set(clientIDHeader, clientID)
			handler.ServeHTTP(w, r)
		})
	}
}

func hashCredentials(wbAccessKey, wbSecretKey string) string {
	sha256Bytes := sha256.Sum256([]byte(wbAccessKey + wbSecretKey))
	return hex.EncodeToString(sha256Bytes[:])
}
