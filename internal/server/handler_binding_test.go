package server_test

import (
	"net/http"
	"testing"

	osbapi "github.com/fivetwenty-io/osbapi/v2/pkg/osbapi"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Bind
// ---------------------------------------------------------------------------

func TestBindHandler_Sync(t *testing.T) {
	t.Parallel()

	broker := &mockBroker{
		bindResp: osbapi.BindResponse{
			Credentials: map[string]any{
				"username": "admin",
				"password": "secret",
			},
		},
		bindAsync: false,
	}

	ts := newTestServer(t, broker)
	defer ts.Close()

	body := osbapi.BindRequest{
		ServiceID: "svc-1",
		PlanID:    "plan-1",
	}

	resp := doRequest(t, http.MethodPut,
		ts.URL+"/v2/service_instances/inst-abc/service_bindings/bind-123", body)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	result := decodeBody[osbapi.BindResponse](t, resp)
	assert.Equal(t, "admin", result.Credentials["username"])
	assert.Equal(t, "secret", result.Credentials["password"])
	assert.Equal(t, "inst-abc", broker.lastInstanceID)
	assert.Equal(t, "bind-123", broker.lastBindingID)
	assert.Equal(t, "svc-1", broker.lastBindReq.ServiceID)
	assert.Equal(t, "plan-1", broker.lastBindReq.PlanID)
}

func TestBindHandler_Async(t *testing.T) {
	t.Parallel()

	broker := &mockBroker{
		bindResp: osbapi.BindResponse{
			Operation: "bind-op-1",
		},
		bindAsync: true,
	}

	ts := newTestServer(t, broker)
	defer ts.Close()

	body := osbapi.BindRequest{
		ServiceID: "svc-1",
		PlanID:    "plan-1",
	}

	resp := doRequest(t, http.MethodPut,
		ts.URL+"/v2/service_instances/inst-abc/service_bindings/bind-123?accepts_incomplete=true", body)
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)

	result := decodeBody[osbapi.BindResponse](t, resp)
	assert.Equal(t, "bind-op-1", result.Operation)
}

func TestBindHandler_AlreadyExists(t *testing.T) {
	t.Parallel()

	broker := &mockBroker{
		bindResp: osbapi.BindResponse{
			Credentials: map[string]any{
				"username": "admin",
			},
		},
		bindErr: osbapi.ErrBindingAlreadyExists,
	}

	ts := newTestServer(t, broker)
	defer ts.Close()

	body := osbapi.BindRequest{
		ServiceID: "svc-1",
		PlanID:    "plan-1",
	}

	resp := doRequest(t, http.MethodPut,
		ts.URL+"/v2/service_instances/inst-abc/service_bindings/bind-123", body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	result := decodeBody[osbapi.BindResponse](t, resp)
	assert.Equal(t, "admin", result.Credentials["username"])
}

// ---------------------------------------------------------------------------
// Fetch Binding
// ---------------------------------------------------------------------------

func TestFetchBindingHandler(t *testing.T) {
	t.Parallel()

	broker := &mockBroker{
		getBindingResp: osbapi.FetchBindingResponse{
			Credentials: map[string]any{
				"uri": "postgres://host/db",
			},
		},
	}

	ts := newTestServer(t, broker)
	defer ts.Close()

	resp := doRequest(t, http.MethodGet,
		ts.URL+"/v2/service_instances/inst-abc/service_bindings/bind-123?service_id=svc-1&plan_id=plan-1", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	result := decodeBody[osbapi.FetchBindingResponse](t, resp)
	assert.Equal(t, "postgres://host/db", result.Credentials["uri"])
	assert.Equal(t, "inst-abc", broker.lastInstanceID)
	assert.Equal(t, "bind-123", broker.lastBindingID)
	assert.Equal(t, "svc-1", broker.lastFetchBindReq.ServiceID)
	assert.Equal(t, "plan-1", broker.lastFetchBindReq.PlanID)
}

// ---------------------------------------------------------------------------
// Unbind
// ---------------------------------------------------------------------------

func TestUnbindHandler_Sync(t *testing.T) {
	t.Parallel()

	broker := &mockBroker{
		unbindAsync: false,
	}

	ts := newTestServer(t, broker)
	defer ts.Close()

	resp := doRequest(t, http.MethodDelete,
		ts.URL+"/v2/service_instances/inst-abc/service_bindings/bind-123?service_id=svc-1&plan_id=plan-1", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	assert.Equal(t, "inst-abc", broker.lastInstanceID)
	assert.Equal(t, "bind-123", broker.lastBindingID)
	assert.Equal(t, "svc-1", broker.lastUnbindReq.ServiceID)
	assert.Equal(t, "plan-1", broker.lastUnbindReq.PlanID)

	// Consume the body to avoid leaking.
	_ = decodeBody[osbapi.UnbindResponse](t, resp)
}

func TestUnbindHandler_Async(t *testing.T) {
	t.Parallel()

	broker := &mockBroker{
		unbindResp: osbapi.UnbindResponse{
			Operation: "unbind-op-1",
		},
		unbindAsync: true,
	}

	ts := newTestServer(t, broker)
	defer ts.Close()

	resp := doRequest(t, http.MethodDelete,
		ts.URL+"/v2/service_instances/inst-abc/service_bindings/bind-123?service_id=svc-1&plan_id=plan-1&accepts_incomplete=true", nil)
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)

	result := decodeBody[osbapi.UnbindResponse](t, resp)
	assert.Equal(t, "unbind-op-1", result.Operation)
}

// ---------------------------------------------------------------------------
// Binding Last Operation
// ---------------------------------------------------------------------------

func TestBindingLastOperationHandler(t *testing.T) {
	t.Parallel()

	broker := &mockBroker{
		lastBindingOpResp: osbapi.LastOperationResponse{
			State:       osbapi.StateInProgress,
			Description: "binding in progress",
		},
	}

	ts := newTestServer(t, broker)
	defer ts.Close()

	resp := doRequest(t, http.MethodGet,
		ts.URL+"/v2/service_instances/inst-abc/service_bindings/bind-123/last_operation?service_id=svc-1&plan_id=plan-1&operation=bind-op-1",
		nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	result := decodeBody[osbapi.LastOperationResponse](t, resp)
	assert.Equal(t, osbapi.StateInProgress, result.State)
	assert.Equal(t, "binding in progress", result.Description)
	assert.Equal(t, "inst-abc", broker.lastInstanceID)
	assert.Equal(t, "bind-123", broker.lastBindingID)
	assert.Equal(t, "bind-op-1", broker.lastLastBindingOpReq.Operation)

	// Verify service_id and plan_id were also passed through.
	require.Equal(t, "svc-1", broker.lastLastBindingOpReq.ServiceID)
	require.Equal(t, "plan-1", broker.lastLastBindingOpReq.PlanID)
}
