package wbapicredentials

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
			log.Debug("Starting wb_api_credentials auth middleware")

			tracer := otel.GetTracerProvider().Tracer("janus/plugin/wbapicredentials")
			ctx, span := tracer.Start(r.Context(), "wbapicredentials.Handler")
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
			foundInCache := false
			// cache doesn't have to be used, if not then it is nil
			if failuresCache != nil {
				// if credentials found in cache it means they're invalid
				if foundInCache = failuresCache.Contains(hashedCred); foundInCache {
					span.RecordError(ErrInvalidCredentials)
					span.SetStatus(codes.Error, "cached authentication failure")
					errors.Handler(w, r, ErrInvalidCredentials)
					return
				}
			}
			if !foundInCache {
				token, success, err := wbClient.Login(r.Context(), wbAccessKey, wbSecretKey)
				if err != nil {
					span.RecordError(err)
					span.SetStatus(codes.Error, err.Error())
					errors.Handler(w, r, ErrInvalidCredentials)
					return
				}
				if !success {
					if failuresCache != nil {
						failuresCache.Put(hashedCred)
					}
					span.RecordError(ErrInvalidCredentials)
					span.SetStatus(codes.Error, "authentication failed")
					errors.Handler(w, r, ErrInvalidCredentials)
					return
				}

				span.SetStatus(codes.Ok, "authentication successful")

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
