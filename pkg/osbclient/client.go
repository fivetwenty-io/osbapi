package osbclient

import (
	internalclient "github.com/fivetwenty-io/osbapi/v2/internal/client"
	"github.com/fivetwenty-io/osbapi/v2/pkg/osbapi"
)

// New creates a new OSB API client with the given configuration.
func New(config osbapi.ClientConfig) (osbapi.Client, error) {
	return internalclient.New(config)
}

// NewWithBasicAuth creates a new OSB API client with HTTP Basic Authentication.
func NewWithBasicAuth(brokerURL, username, password string) (osbapi.Client, error) {
	return internalclient.NewWithBasicAuth(brokerURL, username, password)
}
