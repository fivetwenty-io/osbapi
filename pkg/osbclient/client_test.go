package osbclient_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fivetwenty-io/osbapi/v2/pkg/osbapi"
	"github.com/fivetwenty-io/osbapi/v2/pkg/osbclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_ReturnsClient(t *testing.T) {
	t.Parallel()

	c, err := osbclient.New(osbapi.ClientConfig{
		URL:      "https://broker.example.com",
		Username: "user",
		Password: "pass",
	})

	require.NoError(t, err)
	assert.NotNil(t, c)
}

func TestNew_RequiresURL(t *testing.T) {
	t.Parallel()

	_, err := osbclient.New(osbapi.ClientConfig{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "broker URL is required")
}

func TestNewWithBasicAuth_ReturnsClient(t *testing.T) {
	t.Parallel()

	c, err := osbclient.NewWithBasicAuth("https://broker.example.com", "user", "pass")

	require.NoError(t, err)
	assert.NotNil(t, c)
}

func TestNew_GetCatalog(t *testing.T) {
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
		assert.Equal(t, "2.17", r.Header.Get("X-Broker-API-Version"))

		user, pass, ok := r.BasicAuth()
		assert.True(t, ok, "Basic auth should be present")
		assert.Equal(t, "user", user)
		assert.Equal(t, "pass", pass)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(catalog)
	}))
	defer srv.Close()

	c, err := osbclient.NewWithBasicAuth(srv.URL, "user", "pass")
	require.NoError(t, err)

	result, err := c.GetCatalog(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Services, 1)
	assert.Equal(t, "svc-1", result.Services[0].ID)
	assert.Equal(t, "my-service", result.Services[0].Name)
	assert.True(t, result.Services[0].Bindable)
	require.Len(t, result.Services[0].Plans, 1)
	assert.Equal(t, "plan-1", result.Services[0].Plans[0].ID)
	assert.Equal(t, "default", result.Services[0].Plans[0].Name)
}

func TestNew_Provision(t *testing.T) {
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

	c, err := osbclient.NewWithBasicAuth(srv.URL, "user", "pass")
	require.NoError(t, err)

	resp, isAsync, err := c.Provision(context.Background(), "inst-1", osbapi.ProvisionRequest{
		ServiceID: "svc-1",
		PlanID:    "plan-1",
	}, false)

	require.NoError(t, err)
	assert.False(t, isAsync)
	assert.Equal(t, "https://dashboard.example.com", resp.DashboardURL)
}
