package otel

import (
	"net/http"

	"github.com/hellofresh/janus/pkg/observability"
)

type RequestIDPropagatingTransport struct {
	http.RoundTripper
}

func (t *RequestIDPropagatingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Extract X-Request-Id from context if present
	if requestID := observability.RequestIDFromContext(req.Context()); requestID != "" {
		req.Header.Set("X-Request-Id", requestID)
	}
	return t.RoundTripper.RoundTrip(req)
}
