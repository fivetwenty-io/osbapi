// Package client provides an HTTP-based implementation of the osbapi.Client
// interface for communicating with Open Service Broker API v2.17 brokers.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/fivetwenty-io/osbapi/v2/internal/constants"
	"github.com/fivetwenty-io/osbapi/v2/pkg/osbapi"
)

// Static errors for err113 compliance.
var (
	ErrURLRequired = errors.New("broker URL is required")
	ErrPollFailed  = errors.New("polling failed")
	ErrPollTimeout = errors.New("polling timed out")
)

// maxErrorBodySnippet is the maximum number of bytes from a response body to
// include in a generic error message when the body cannot be parsed as JSON.
const maxErrorBodySnippet = 200

// Client implements the osbapi.Client interface using HTTP.
type Client struct {
	baseURL    string
	username   string
	password   string
	apiVersion string
	httpClient *http.Client
	logger     osbapi.Logger
	userAgent  string
	verbose    bool
}

// Compile-time interface check.
var _ osbapi.Client = (*Client)(nil)

// New creates a new Client from the given config.
func New(config osbapi.ClientConfig) (*Client, error) {
	if config.URL == "" {
		return nil, ErrURLRequired
	}

	baseURL := strings.TrimSuffix(config.URL, "/")

	apiVersion := config.APIVersion
	if apiVersion == "" {
		apiVersion = osbapi.APIVersion
	}

	httpTimeout := config.HTTPTimeout
	if httpTimeout == 0 {
		httpTimeout = constants.DefaultHTTPTimeout
	}

	return &Client{
		baseURL:    baseURL,
		username:   config.Username,
		password:   config.Password,
		apiVersion: apiVersion,
		httpClient: &http.Client{Timeout: httpTimeout},
		logger:     config.Logger,
		userAgent:  config.UserAgent,
		verbose:    config.Verbose,
	}, nil
}

// NewWithBasicAuth is a convenience constructor.
func NewWithBasicAuth(brokerURL, username, password string) (*Client, error) {
	return New(osbapi.ClientConfig{
		URL:      brokerURL,
		Username: username,
		Password: password,
	})
}

// GetCatalog retrieves the service broker's catalog of services and plans.
func (c *Client) GetCatalog(ctx context.Context) (*osbapi.Catalog, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/v2/catalog", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("requesting catalog: %w", err)
	}
	defer resp.Body.Close()

	var catalog osbapi.Catalog
	if err := c.handleResponse(resp, &catalog); err != nil {
		return nil, err
	}

	return &catalog, nil
}

// Provision requests provisioning of a new service instance.
func (c *Client) Provision(
	ctx context.Context,
	instanceID string,
	req osbapi.ProvisionRequest,
	async bool,
) (osbapi.ProvisionResponse, bool, error) {
	query := url.Values{}
	if async {
		query.Set("accepts_incomplete", "true")
	}

	body, err := json.Marshal(req)
	if err != nil {
		return osbapi.ProvisionResponse{}, false, fmt.Errorf("marshaling provision request: %w", err)
	}

	path := fmt.Sprintf("/v2/service_instances/%s", instanceID)

	resp, err := c.doRequest(ctx, http.MethodPut, path, query, body)
	if err != nil {
		return osbapi.ProvisionResponse{}, false, fmt.Errorf("sending provision request: %w", err)
	}
	defer resp.Body.Close()

	var result osbapi.ProvisionResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return osbapi.ProvisionResponse{}, false, err
	}

	isAsync := resp.StatusCode == http.StatusAccepted

	return result, isAsync, nil
}

// Deprovision requests deletion of a service instance.
func (c *Client) Deprovision(
	ctx context.Context,
	instanceID string,
	req osbapi.DeprovisionRequest,
	async bool,
) (osbapi.DeprovisionResponse, bool, error) {
	query := url.Values{}
	query.Set("service_id", req.ServiceID)
	query.Set("plan_id", req.PlanID)

	if async {
		query.Set("accepts_incomplete", "true")
	}

	path := fmt.Sprintf("/v2/service_instances/%s", instanceID)

	resp, err := c.doRequest(ctx, http.MethodDelete, path, query, nil)
	if err != nil {
		return osbapi.DeprovisionResponse{}, false, fmt.Errorf("sending deprovision request: %w", err)
	}
	defer resp.Body.Close()

	var result osbapi.DeprovisionResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return osbapi.DeprovisionResponse{}, false, err
	}

	isAsync := resp.StatusCode == http.StatusAccepted

	return result, isAsync, nil
}

// GetInstance retrieves a service instance.
func (c *Client) GetInstance(
	ctx context.Context,
	instanceID string,
	req osbapi.FetchInstanceRequest,
) (osbapi.FetchInstanceResponse, error) {
	query := url.Values{}

	if req.ServiceID != "" {
		query.Set("service_id", req.ServiceID)
	}

	if req.PlanID != "" {
		query.Set("plan_id", req.PlanID)
	}

	path := fmt.Sprintf("/v2/service_instances/%s", instanceID)

	resp, err := c.doRequest(ctx, http.MethodGet, path, query, nil)
	if err != nil {
		return osbapi.FetchInstanceResponse{}, fmt.Errorf("fetching instance: %w", err)
	}
	defer resp.Body.Close()

	var result osbapi.FetchInstanceResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return osbapi.FetchInstanceResponse{}, err
	}

	return result, nil
}

// Update requests modification of a service instance.
func (c *Client) Update(
	ctx context.Context,
	instanceID string,
	req osbapi.UpdateRequest,
	async bool,
) (osbapi.UpdateResponse, bool, error) {
	query := url.Values{}
	if async {
		query.Set("accepts_incomplete", "true")
	}

	body, err := json.Marshal(req)
	if err != nil {
		return osbapi.UpdateResponse{}, false, fmt.Errorf("marshaling update request: %w", err)
	}

	path := fmt.Sprintf("/v2/service_instances/%s", instanceID)

	resp, err := c.doRequest(ctx, http.MethodPatch, path, query, body)
	if err != nil {
		return osbapi.UpdateResponse{}, false, fmt.Errorf("sending update request: %w", err)
	}
	defer resp.Body.Close()

	var result osbapi.UpdateResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return osbapi.UpdateResponse{}, false, err
	}

	isAsync := resp.StatusCode == http.StatusAccepted

	return result, isAsync, nil
}

// PollLastOperation polls the status of an async instance operation.
func (c *Client) PollLastOperation(
	ctx context.Context,
	instanceID string,
	req osbapi.LastOperationRequest,
) (osbapi.LastOperationResponse, error) {
	query := url.Values{}

	if req.ServiceID != "" {
		query.Set("service_id", req.ServiceID)
	}

	if req.PlanID != "" {
		query.Set("plan_id", req.PlanID)
	}

	if req.Operation != "" {
		query.Set("operation", req.Operation)
	}

	path := fmt.Sprintf("/v2/service_instances/%s/last_operation", instanceID)

	resp, err := c.doRequest(ctx, http.MethodGet, path, query, nil)
	if err != nil {
		return osbapi.LastOperationResponse{}, fmt.Errorf("polling last operation: %w", err)
	}
	defer resp.Body.Close()

	var result osbapi.LastOperationResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return osbapi.LastOperationResponse{}, err
	}

	return result, nil
}

// PollInstanceUntilComplete polls a last operation endpoint until the
// operation succeeds, fails, or the timeout is reached.
func (c *Client) PollInstanceUntilComplete(
	ctx context.Context,
	instanceID string,
	opts osbapi.PollConfig,
) (osbapi.LastOperationResponse, error) {
	interval := opts.Interval
	if interval == 0 {
		interval = constants.DefaultPollInterval
	}

	timeout := opts.Timeout
	if timeout == 0 {
		timeout = constants.DefaultPollTimeout
	}

	deadline := time.Now().Add(timeout)
	req := osbapi.LastOperationRequest{}

	for {
		result, err := c.PollLastOperation(ctx, instanceID, req)
		if err != nil {
			return osbapi.LastOperationResponse{}, fmt.Errorf("%w: %w", ErrPollFailed, err)
		}

		switch result.State {
		case osbapi.StateSucceeded:
			return result, nil
		case osbapi.StateFailed:
			return result, fmt.Errorf("%w: operation failed: %s", ErrPollFailed, result.Description)
		}

		if time.Now().Add(interval).After(deadline) {
			return osbapi.LastOperationResponse{}, ErrPollTimeout
		}

		select {
		case <-ctx.Done():
			return osbapi.LastOperationResponse{}, fmt.Errorf("polling cancelled: %w", ctx.Err())
		case <-time.After(interval):
		}
	}
}

// doRequest builds and executes an HTTP request with authentication headers,
// API version, content-type, and user-agent.
func (c *Client) doRequest(
	ctx context.Context,
	method, path string,
	query url.Values,
	body []byte,
) (*http.Response, error) {
	requestURL := c.baseURL + path

	if len(query) > 0 {
		requestURL += "?" + query.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, requestURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set(osbapi.HeaderAPIVersion, c.apiVersion)
	req.Header.Set("Content-Type", "application/json")

	if c.username != "" || c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	if c.verbose && c.logger != nil {
		c.logger.Debug("OSB API request", map[string]any{
			"method": method,
			"url":    requestURL,
		})
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	if c.verbose && c.logger != nil {
		c.logger.Debug("OSB API response", map[string]any{
			"status": resp.StatusCode,
			"method": method,
			"url":    requestURL,
		})
	}

	return resp, nil
}

// handleResponse reads the response body, checks the status code, and
// unmarshals JSON into the result. For error status codes, it returns an
// *osbapi.OSBError or a generic error.
func (c *Client) handleResponse(resp *http.Response, result any) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode >= constants.HTTPStatusBadRequest {
		return c.handleErrorResponse(resp, body)
	}

	if len(body) > 0 && result != nil {
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("unmarshaling response: %w", err)
		}
	}

	return nil
}

// handleErrorResponse tries to unmarshal the body as an OSBError. If that
// fails, it returns a generic error with the status code and a snippet of
// the response body.
func (c *Client) handleErrorResponse(resp *http.Response, body []byte) error {
	var osbErr osbapi.OSBError
	if err := json.Unmarshal(body, &osbErr); err == nil && (osbErr.ErrorCode != "" || osbErr.Description != "") {
		osbErr.StatusCode = resp.StatusCode

		return &osbErr
	}

	snippet := string(body)
	if len(snippet) > maxErrorBodySnippet {
		snippet = snippet[:maxErrorBodySnippet]
	}

	return fmt.Errorf("HTTP %d: %s", resp.StatusCode, snippet)
}
