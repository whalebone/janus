package wbmicrocredentials

import (
	"net/http"

	"github.com/hellofresh/janus/pkg/errors"
)

var (
	// ErrNotAuthorized is used when the the access is not permisted
	ErrNotAuthorized = errors.New(http.StatusUnauthorized, "not authorized")
	// ErrInvalidCredentials is used when provided credentials are wrong
	ErrInvalidCredentials = errors.New(http.StatusUnauthorized, "invalid combination of access key and secret key")
)
