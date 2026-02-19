package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fivetwenty-io/osbapi/v2/internal/client"
	"github.com/fivetwenty-io/osbapi/v2/pkg/osbapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBind_Sync(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/v2/service_instances/inst-1/service_bindings/bind-1", r.URL.Path)

		var req osbapi.BindRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "svc-1", req.ServiceID)
		assert.Equal(t, "plan-1", req.PlanID)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(osbapi.BindResponse{
			Credentials: map[string]any{
				"username": "admin",
				"password": "secret",
			},
		})
	}))
	defer srv.Close()

	c, err := client.NewWithBasicAuth(srv.URL, "user", "pass")
	require.NoError(t, err)

	resp, isAsync, err := c.Bind(context.Background(), "inst-1", "bind-1", osbapi.BindRequest{
		ServiceID: "svc-1",
		PlanID:    "plan-1",
	}, false)

	require.NoError(t, err)
	assert.False(t, isAsync)
	assert.Equal(t, "admin", resp.Credentials["username"])
	assert.Equal(t, "secret", resp.Credentials["password"])
}

func TestBind_Async(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/v2/service_instances/inst-1/service_bindings/bind-1", r.URL.Path)
		assert.Equal(t, "true", r.URL.Query().Get("accepts_incomplete"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(osbapi.BindResponse{
			Operation: "bind-op-1",
		})
	}))
	defer srv.Close()

	c, err := client.NewWithBasicAuth(srv.URL, "user", "pass")
	require.NoError(t, err)

	resp, isAsync, err := c.Bind(context.Background(), "inst-1", "bind-1", osbapi.BindRequest{
		ServiceID: "svc-1",
		PlanID:    "plan-1",
	}, true)

	require.NoError(t, err)
	assert.True(t, isAsync)
	assert.Equal(t, "bind-op-1", resp.Operation)
}

func TestUnbind_Sync(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/v2/service_instances/inst-1/service_bindings/bind-1", r.URL.Path)
		assert.Equal(t, "svc-1", r.URL.Query().Get("service_id"))
		assert.Equal(t, "plan-1", r.URL.Query().Get("plan_id"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(osbapi.UnbindResponse{})
	}))
	defer srv.Close()

	c, err := client.NewWithBasicAuth(srv.URL, "user", "pass")
	require.NoError(t, err)

	resp, isAsync, err := c.Unbind(context.Background(), "inst-1", "bind-1", osbapi.UnbindRequest{
		ServiceID: "svc-1",
		PlanID:    "plan-1",
	}, false)

	require.NoError(t, err)
	assert.False(t, isAsync)
	assert.Equal(t, osbapi.UnbindResponse{}, resp)
}

func TestGetBinding(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v2/service_instances/inst-1/service_bindings/bind-1", r.URL.Path)
		assert.Equal(t, "svc-1", r.URL.Query().Get("service_id"))
		assert.Equal(t, "plan-1", r.URL.Query().Get("plan_id"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(osbapi.FetchBindingResponse{
			Credentials: map[string]any{
				"uri": "postgres://host:5432/db",
			},
			SyslogDrainURL: "syslog://logs.example.com",
		})
	}))
	defer srv.Close()

	c, err := client.NewWithBasicAuth(srv.URL, "user", "pass")
	require.NoError(t, err)

	resp, err := c.GetBinding(context.Background(), "inst-1", "bind-1", osbapi.FetchBindingRequest{
		ServiceID: "svc-1",
		PlanID:    "plan-1",
	})

	require.NoError(t, err)
	assert.Equal(t, "postgres://host:5432/db", resp.Credentials["uri"])
	assert.Equal(t, "syslog://logs.example.com", resp.SyslogDrainURL)
}

func TestPollBindingLastOperation(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t,
			"/v2/service_instances/inst-1/service_bindings/bind-1/last_operation",
			r.URL.Path,
		)
		assert.Equal(t, "svc-1", r.URL.Query().Get("service_id"))
		assert.Equal(t, "plan-1", r.URL.Query().Get("plan_id"))
		assert.Equal(t, "bind-op", r.URL.Query().Get("operation"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(osbapi.LastOperationResponse{
			State:       osbapi.StateSucceeded,
			Description: "binding complete",
		})
	}))
	defer srv.Close()

	c, err := client.NewWithBasicAuth(srv.URL, "user", "pass")
	require.NoError(t, err)

	resp, err := c.PollBindingLastOperation(
		context.Background(), "inst-1", "bind-1", osbapi.LastOperationRequest{
			ServiceID: "svc-1",
			PlanID:    "plan-1",
			Operation: "bind-op",
		},
	)

	require.NoError(t, err)
	assert.Equal(t, osbapi.StateSucceeded, resp.State)
	assert.Equal(t, "binding complete", resp.Description)
}

func TestBind_ErrorResponse(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(osbapi.OSBError{
			ErrorCode:   "RequiresApp",
			Description: "This service supports generation of credentials through binding an application only.",
		})
	}))
	defer srv.Close()

	c, err := client.NewWithBasicAuth(srv.URL, "user", "pass")
	require.NoError(t, err)

	_, _, err = c.Bind(context.Background(), "inst-1", "bind-1", osbapi.BindRequest{
		ServiceID: "svc-1",
		PlanID:    "plan-1",
	}, false)

	require.Error(t, err)

	var osbErr *osbapi.OSBError

	require.ErrorAs(t, err, &osbErr)
	assert.Equal(t, "RequiresApp", osbErr.ErrorCode)
	assert.Equal(t, http.StatusUnprocessableEntity, osbErr.StatusCode)
	assert.Contains(t, osbErr.Description, "binding an application")
}
