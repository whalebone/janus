package wbapicredentials

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/hellofresh/janus/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// NewWBAPICredAuth is a HTTP auth middleware using WB API credentials service
func NewWBAPICredAuth(
	wbClient *WBAPICredClient,
	failuresCache *Cache,
	accessKeyHeader,
	secretKeyHeader,
	tokenHeader string,
) func(http.Handler) http.Handler {
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
			foundInCache := false
			// cache doesn't have to be used, if not then it is nil
			if failuresCache != nil {
				// if credentials found in cache it means they're invalid
				if foundInCache = failuresCache.Contains(hashedCred); foundInCache {
					errors.Handler(w, r, ErrInvalidCredentials)
					return
				}
			}
			if !foundInCache {
				token, success, err := wbClient.Login(r.Context(), wbAccessKey, wbSecretKey)
				if err != nil {
					errors.Handler(w, r, ErrInvalidCredentials)
					return
				}
				if !success {
					if failuresCache != nil {
						failuresCache.Put(hashedCred)
					}
					errors.Handler(w, r, ErrInvalidCredentials)
					return
				}
				// add used identification headers
				r.Header.Set(tokenHeader, token)
				// remove secret key from the request header
				r.Header.Del(secretKeyHeader)
				handler.ServeHTTP(w, r)
			}
		})
	}
}

func hashCredentials(wbAccessKey, wbSecretKey string) string {
	sha256Bytes := sha256.Sum256([]byte(wbAccessKey + wbSecretKey))
	return hex.EncodeToString(sha256Bytes[:])
}
