// Package constants provides default configuration values for the OSB API
// client and server implementations.
package constants

import "time"

const (
	// DefaultHTTPTimeout is the default timeout for HTTP requests.
	DefaultHTTPTimeout = 30 * time.Second

	// DefaultPollInterval is the default interval between polling requests.
	DefaultPollInterval = 5 * time.Second

	// DefaultPollTimeout is the default maximum time to wait for async operations.
	DefaultPollTimeout = 30 * time.Minute

	// MaxPollInterval is the maximum allowed polling interval.
	MaxPollInterval = 60 * time.Second

	// HTTPStatusBadRequest is the threshold for error status codes.
	HTTPStatusBadRequest = 400
)
