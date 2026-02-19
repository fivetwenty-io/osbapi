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

func TestNew_RequiresURL(t *testing.T) {
	t.Parallel()

	_, err := client.New(osbapi.ClientConfig{})

	require.Error(t, err)
	assert.ErrorIs(t, err, client.ErrURLRequired)
}

func TestNew_DefaultConfig(t *testing.T) {
	t.Parallel()

	c, err := client.New(osbapi.ClientConfig{
		URL: "https://broker.example.com",
	})

	require.NoError(t, err)
	assert.NotNil(t, c)
}

func TestGetCatalog(t *testing.T) {
	t.Parallel()

	catalog := osbapi.Catalog{
		Services: []osbapi.Service{
			{
				ID:          "svc-1",
				Name:        "my-service",
				Description: "A test service",
				Bindable:    true,
				Plans: []osbapi.Plan{
					{
						ID:          "plan-1",
						Name:        "default",
						Description: "Default plan",
					},
				},
			},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v2/catalog", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(catalog)
	}))
	defer srv.Close()

	c, err := client.NewWithBasicAuth(srv.URL, "user", "pass")
	require.NoError(t, err)

	result, err := c.GetCatalog(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Services, 1)
	assert.Equal(t, "svc-1", result.Services[0].ID)
	assert.Equal(t, "my-service", result.Services[0].Name)
	require.Len(t, result.Services[0].Plans, 1)
	assert.Equal(t, "plan-1", result.Services[0].Plans[0].ID)
}

func TestProvision_Sync(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/v2/service_instances/inst-1", r.URL.Path)

		var req osbapi.ProvisionRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "svc-1", req.ServiceID)
		assert.Equal(t, "plan-1", req.PlanID)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(osbapi.ProvisionResponse{
			DashboardURL: "https://dashboard.example.com",
		})
	}))
	defer srv.Close()

	c, err := client.NewWithBasicAuth(srv.URL, "user", "pass")
	require.NoError(t, err)

	resp, isAsync, err := c.Provision(context.Background(), "inst-1", osbapi.ProvisionRequest{
		ServiceID: "svc-1",
		PlanID:    "plan-1",
	}, false)

	require.NoError(t, err)
	assert.False(t, isAsync)
	assert.Equal(t, "https://dashboard.example.com", resp.DashboardURL)
}

func TestProvision_Async(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "true", r.URL.Query().Get("accepts_incomplete"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(osbapi.ProvisionResponse{
			Operation: "provision-op-1",
		})
	}))
	defer srv.Close()

	c, err := client.NewWithBasicAuth(srv.URL, "user", "pass")
	require.NoError(t, err)

	resp, isAsync, err := c.Provision(context.Background(), "inst-1", osbapi.ProvisionRequest{
		ServiceID: "svc-1",
		PlanID:    "plan-1",
	}, true)

	require.NoError(t, err)
	assert.True(t, isAsync)
	assert.Equal(t, "provision-op-1", resp.Operation)
}

func TestDeprovision_Sync(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/v2/service_instances/inst-1", r.URL.Path)
		assert.Equal(t, "svc-1", r.URL.Query().Get("service_id"))
		assert.Equal(t, "plan-1", r.URL.Query().Get("plan_id"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(osbapi.DeprovisionResponse{})
	}))
	defer srv.Close()

	c, err := client.NewWithBasicAuth(srv.URL, "user", "pass")
	require.NoError(t, err)

	resp, isAsync, err := c.Deprovision(context.Background(), "inst-1", osbapi.DeprovisionRequest{
		ServiceID: "svc-1",
		PlanID:    "plan-1",
	}, false)

	require.NoError(t, err)
	assert.False(t, isAsync)
	assert.Equal(t, osbapi.DeprovisionResponse{}, resp)
}

func TestGetInstance(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v2/service_instances/inst-1", r.URL.Path)
		assert.Equal(t, "svc-1", r.URL.Query().Get("service_id"))
		assert.Equal(t, "plan-1", r.URL.Query().Get("plan_id"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(osbapi.FetchInstanceResponse{
			ServiceID:    "svc-1",
			PlanID:       "plan-1",
			DashboardURL: "https://dashboard.example.com",
		})
	}))
	defer srv.Close()

	c, err := client.NewWithBasicAuth(srv.URL, "user", "pass")
	require.NoError(t, err)

	resp, err := c.GetInstance(context.Background(), "inst-1", osbapi.FetchInstanceRequest{
		ServiceID: "svc-1",
		PlanID:    "plan-1",
	})

	require.NoError(t, err)
	assert.Equal(t, "svc-1", resp.ServiceID)
	assert.Equal(t, "plan-1", resp.PlanID)
	assert.Equal(t, "https://dashboard.example.com", resp.DashboardURL)
}

func TestUpdate_Sync(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Equal(t, "/v2/service_instances/inst-1", r.URL.Path)

		var req osbapi.UpdateRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "svc-1", req.ServiceID)
		assert.Equal(t, "plan-2", req.PlanID)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(osbapi.UpdateResponse{
			DashboardURL: "https://updated-dashboard.example.com",
		})
	}))
	defer srv.Close()

	c, err := client.NewWithBasicAuth(srv.URL, "user", "pass")
	require.NoError(t, err)

	resp, isAsync, err := c.Update(context.Background(), "inst-1", osbapi.UpdateRequest{
		ServiceID: "svc-1",
		PlanID:    "plan-2",
	}, false)

	require.NoError(t, err)
	assert.False(t, isAsync)
	assert.Equal(t, "https://updated-dashboard.example.com", resp.DashboardURL)
}

func TestPollLastOperation(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v2/service_instances/inst-1/last_operation", r.URL.Path)
		assert.Equal(t, "svc-1", r.URL.Query().Get("service_id"))
		assert.Equal(t, "plan-1", r.URL.Query().Get("plan_id"))
		assert.Equal(t, "provision-op", r.URL.Query().Get("operation"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(osbapi.LastOperationResponse{
			State:       osbapi.StateSucceeded,
			Description: "provisioning complete",
		})
	}))
	defer srv.Close()

	c, err := client.NewWithBasicAuth(srv.URL, "user", "pass")
	require.NoError(t, err)

	resp, err := c.PollLastOperation(context.Background(), "inst-1", osbapi.LastOperationRequest{
		ServiceID: "svc-1",
		PlanID:    "plan-1",
		Operation: "provision-op",
	})

	require.NoError(t, err)
	assert.Equal(t, osbapi.StateSucceeded, resp.State)
	assert.Equal(t, "provisioning complete", resp.Description)
}

func TestErrorResponse(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(osbapi.OSBError{
			ErrorCode:   "AsyncRequired",
			Description: "This service plan requires client support for asynchronous operations.",
		})
	}))
	defer srv.Close()

	c, err := client.NewWithBasicAuth(srv.URL, "user", "pass")
	require.NoError(t, err)

	_, err = c.GetCatalog(context.Background())

	require.Error(t, err)

	var osbErr *osbapi.OSBError

	require.ErrorAs(t, err, &osbErr)
	assert.Equal(t, "AsyncRequired", osbErr.ErrorCode)
	assert.Equal(t, http.StatusUnprocessableEntity, osbErr.StatusCode)
	assert.Contains(t, osbErr.Description, "asynchronous")
}

func TestBasicAuth_Sent(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()

		assert.True(t, ok, "Basic auth should be present")
		assert.Equal(t, "admin", user)
		assert.Equal(t, "secret", pass)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(osbapi.Catalog{Services: []osbapi.Service{}})
	}))
	defer srv.Close()

	c, err := client.NewWithBasicAuth(srv.URL, "admin", "secret")
	require.NoError(t, err)

	_, err = c.GetCatalog(context.Background())
	require.NoError(t, err)
}

func TestAPIVersionHeader_Sent(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "2.17", r.Header.Get("X-Broker-API-Version"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(osbapi.Catalog{Services: []osbapi.Service{}})
	}))
	defer srv.Close()

	c, err := client.NewWithBasicAuth(srv.URL, "user", "pass")
	require.NoError(t, err)

	_, err = c.GetCatalog(context.Background())
	require.NoError(t, err)
}
