package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/fivetwenty-io/osbapi/v2/internal/constants"
	"github.com/fivetwenty-io/osbapi/v2/pkg/osbapi"
)

// Bind requests creation of a new service binding.
func (c *Client) Bind(
	ctx context.Context,
	instanceID, bindingID string,
	req osbapi.BindRequest,
	async bool,
) (osbapi.BindResponse, bool, error) {
	query := url.Values{}
	if async {
		query.Set("accepts_incomplete", "true")
	}

	body, err := json.Marshal(req)
	if err != nil {
		return osbapi.BindResponse{}, false, fmt.Errorf("marshaling bind request: %w", err)
	}

	path := fmt.Sprintf("/v2/service_instances/%s/service_bindings/%s", instanceID, bindingID)

	resp, err := c.doRequest(ctx, http.MethodPut, path, query, body)
	if err != nil {
		return osbapi.BindResponse{}, false, fmt.Errorf("sending bind request: %w", err)
	}
	defer resp.Body.Close()

	var result osbapi.BindResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return osbapi.BindResponse{}, false, err
	}

	isAsync := resp.StatusCode == http.StatusAccepted

	return result, isAsync, nil
}

// Unbind requests deletion of a service binding.
func (c *Client) Unbind(
	ctx context.Context,
	instanceID, bindingID string,
	req osbapi.UnbindRequest,
	async bool,
) (osbapi.UnbindResponse, bool, error) {
	query := url.Values{}
	query.Set("service_id", req.ServiceID)
	query.Set("plan_id", req.PlanID)

	if async {
		query.Set("accepts_incomplete", "true")
	}

	path := fmt.Sprintf("/v2/service_instances/%s/service_bindings/%s", instanceID, bindingID)

	resp, err := c.doRequest(ctx, http.MethodDelete, path, query, nil)
	if err != nil {
		return osbapi.UnbindResponse{}, false, fmt.Errorf("sending unbind request: %w", err)
	}
	defer resp.Body.Close()

	var result osbapi.UnbindResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return osbapi.UnbindResponse{}, false, err
	}

	isAsync := resp.StatusCode == http.StatusAccepted

	return result, isAsync, nil
}

// GetBinding retrieves a service binding.
func (c *Client) GetBinding(
	ctx context.Context,
	instanceID, bindingID string,
	req osbapi.FetchBindingRequest,
) (osbapi.FetchBindingResponse, error) {
	query := url.Values{}

	if req.ServiceID != "" {
		query.Set("service_id", req.ServiceID)
	}

	if req.PlanID != "" {
		query.Set("plan_id", req.PlanID)
	}

	path := fmt.Sprintf("/v2/service_instances/%s/service_bindings/%s", instanceID, bindingID)

	resp, err := c.doRequest(ctx, http.MethodGet, path, query, nil)
	if err != nil {
		return osbapi.FetchBindingResponse{}, fmt.Errorf("fetching binding: %w", err)
	}
	defer resp.Body.Close()

	var result osbapi.FetchBindingResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return osbapi.FetchBindingResponse{}, err
	}

	return result, nil
}

// PollBindingLastOperation polls the status of an async binding operation.
func (c *Client) PollBindingLastOperation(
	ctx context.Context,
	instanceID, bindingID string,
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

	path := fmt.Sprintf(
		"/v2/service_instances/%s/service_bindings/%s/last_operation",
		instanceID, bindingID,
	)

	resp, err := c.doRequest(ctx, http.MethodGet, path, query, nil)
	if err != nil {
		return osbapi.LastOperationResponse{}, fmt.Errorf("polling binding last operation: %w", err)
	}
	defer resp.Body.Close()

	var result osbapi.LastOperationResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return osbapi.LastOperationResponse{}, err
	}

	return result, nil
}

// PollBindingUntilComplete polls a binding last operation endpoint until the
// operation succeeds, fails, or the timeout is reached.
func (c *Client) PollBindingUntilComplete(
	ctx context.Context,
	instanceID, bindingID string,
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
		result, err := c.PollBindingLastOperation(ctx, instanceID, bindingID, req)
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
