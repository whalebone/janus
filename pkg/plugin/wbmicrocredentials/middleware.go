package wbmicrocredentials

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/hellofresh/janus/pkg/errors"
	"github.com/hellofresh/janus/pkg/observability"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// NewWBMicroCredAuth is a HTTP basic auth middleware
func NewWBMicroCredAuth(
	wbClient *WBMicroCredClient,
	cache *CredentialsCache,
	accessKeyHeader,
	secretKeyHeader,
	clientIDHeader,
	userIDHeader string,

) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Debug("Starting wb_micro_credentials auth middleware")

			tracer := otel.GetTracerProvider().Tracer("janus/plugin/wbmicrocredentials")
			ctx, span := tracer.Start(r.Context(), "wbmicrocredentials.Handler")
			defer span.End()
			r = r.WithContext(ctx)

			wbAccessKey := r.Header.Get(accessKeyHeader)
			wbSecretKey := r.Header.Get(secretKeyHeader)
			if wbAccessKey == "" || wbSecretKey == "" {
				span.RecordError(ErrNotAuthorized)
				span.SetStatus(codes.Error, "missing credentials")
				errors.Handler(w, r, ErrNotAuthorized)
				return
			}

			span.SetAttributes(
				attribute.String("request.id", observability.RequestIDFromContext(r.Context())),
			)

			// hashed access key and secret key to be used as cache key
			hashedCred := hashCredentials(wbAccessKey, wbSecretKey)
			// client id as returned from WB microCredentials service after login call
			var clientID string
			var userID string
			foundInCache := false
			// cache doesn't have to be used, if not then it is nil
			if cache != nil {
				var cachedCred *CachedCredentials
				if cachedCred, foundInCache = cache.Get(hashedCred); foundInCache {
					if !cachedCred.LoginSuccess {
						span.RecordError(ErrInvalidCredentials)
						span.SetStatus(codes.Error, "cached authentication failure")
						errors.Handler(w, r, ErrInvalidCredentials)
						return
					}
					clientID = cachedCred.ClientID
					userID = cachedCred.UserID
				}
			}
			if !foundInCache {
				var success bool
				var err error
				clientID, userID, success, err = wbClient.Login(r.Context(), wbAccessKey, wbSecretKey)
				if err != nil {
					span.RecordError(err)
					span.SetStatus(codes.Error, err.Error())
					errors.Handler(w, r, ErrInvalidCredentials)
					return
				}
				if cache != nil {
					cache.Put(hashedCred, NewCachedCredentials(clientID, userID, success))
				}
				if !success {
					span.RecordError(ErrInvalidCredentials)
					span.SetStatus(codes.Error, "authentication rejected")
					errors.Handler(w, r, ErrInvalidCredentials)
					return
				}
			}

			span.SetStatus(codes.Ok, "authentication successful")

			// add used identification headers
			r.Header.Set(clientIDHeader, clientID)
			r.Header.Set(userIDHeader, userID)
			// remove secret key from the request header
			r.Header.Del(secretKeyHeader)
			handler.ServeHTTP(w, r)
		})
	}
}

func hashCredentials(wbAccessKey, wbSecretKey string) string {
	sha256Bytes := sha256.Sum256([]byte(wbAccessKey + wbSecretKey))
	return hex.EncodeToString(sha256Bytes[:])
}
