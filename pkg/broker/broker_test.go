package broker_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fivetwenty-io/osbapi/v2/pkg/broker"
	"github.com/fivetwenty-io/osbapi/v2/pkg/osbapi"
)

// ---------------------------------------------------------------------------
// Mock broker
// ---------------------------------------------------------------------------

// mockBroker is a minimal ServiceBroker implementation for testing the public
// broker wrapper. It returns a static catalog and zero-value responses for
// all other endpoints.
type mockBroker struct {
	catalog *osbapi.Catalog
}

func newMockBroker() *mockBroker {
	return &mockBroker{
		catalog: &osbapi.Catalog{
			Services: []osbapi.Service{
				{
					ID:          "test-service-id",
					Name:        "test-service",
					Description: "A test service",
					Bindable:    true,
					Plans: []osbapi.Plan{
						{
							ID:          "test-plan-id",
							Name:        "default",
							Description: "Default plan",
						},
					},
				},
			},
		},
	}
}

func (m *mockBroker) GetCatalog(_ context.Context) (*osbapi.Catalog, error) {
	return m.catalog, nil
}

func (m *mockBroker) Provision(_ context.Context, _ string, _ osbapi.ProvisionRequest, _ bool) (osbapi.ProvisionResponse, bool, error) {
	return osbapi.ProvisionResponse{}, false, nil
}

func (m *mockBroker) Deprovision(_ context.Context, _ string, _ osbapi.DeprovisionRequest, _ bool) (osbapi.DeprovisionResponse, bool, error) {
	return osbapi.DeprovisionResponse{}, false, nil
}

func (m *mockBroker) GetInstance(_ context.Context, _ string, _ osbapi.FetchInstanceRequest) (osbapi.FetchInstanceResponse, error) {
	return osbapi.FetchInstanceResponse{}, nil
}

func (m *mockBroker) Update(_ context.Context, _ string, _ osbapi.UpdateRequest, _ bool) (osbapi.UpdateResponse, bool, error) {
	return osbapi.UpdateResponse{}, false, nil
}

func (m *mockBroker) LastOperation(_ context.Context, _ string, _ osbapi.LastOperationRequest) (osbapi.LastOperationResponse, error) {
	return osbapi.LastOperationResponse{State: osbapi.StateSucceeded}, nil
}

func (m *mockBroker) Bind(_ context.Context, _, _ string, _ osbapi.BindRequest, _ bool) (osbapi.BindResponse, bool, error) {
	return osbapi.BindResponse{}, false, nil
}

func (m *mockBroker) Unbind(_ context.Context, _, _ string, _ osbapi.UnbindRequest, _ bool) (osbapi.UnbindResponse, bool, error) {
	return osbapi.UnbindResponse{}, false, nil
}

func (m *mockBroker) GetBinding(_ context.Context, _, _ string, _ osbapi.FetchBindingRequest) (osbapi.FetchBindingResponse, error) {
	return osbapi.FetchBindingResponse{}, nil
}

func (m *mockBroker) LastBindingOperation(_ context.Context, _, _ string, _ osbapi.LastOperationRequest) (osbapi.LastOperationResponse, error) {
	return osbapi.LastOperationResponse{State: osbapi.StateSucceeded}, nil
}

// ---------------------------------------------------------------------------
// Mock logger
// ---------------------------------------------------------------------------

type mockLogger struct {
	called bool
}

func (l *mockLogger) Debug(_ string, _ map[string]any) { l.called = true }
func (l *mockLogger) Info(_ string, _ map[string]any)  { l.called = true }
func (l *mockLogger) Warn(_ string, _ map[string]any)  { l.called = true }
func (l *mockLogger) Error(_ string, _ map[string]any) { l.called = true }

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestNewHandler_ReturnsHandler(t *testing.T) {
	t.Parallel()

	h := broker.NewHandler(newMockBroker())
	assert.NotNil(t, h, "NewHandler must return a non-nil http.Handler")

	// Verify it satisfies the http.Handler interface at compile time.
	var _ http.Handler = h
}

func TestNewHandler_ServesEndpoints(t *testing.T) {
	t.Parallel()

	h := broker.NewHandler(newMockBroker())
	srv := httptest.NewServer(h)
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/v2/catalog", nil)
	require.NoError(t, err)
	req.Header.Set("X-Broker-API-Version", "2.17")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "GET /v2/catalog should return 200")
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
}

func TestWithBasicAuth(t *testing.T) {
	t.Parallel()

	h := broker.NewHandler(newMockBroker(),
		broker.WithBasicAuth("admin", "secret"),
	)
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	t.Run("rejects unauthenticated request", func(t *testing.T) {
		t.Parallel()

		req, err := http.NewRequest(http.MethodGet, srv.URL+"/v2/catalog", nil)
		require.NoError(t, err)
		req.Header.Set("X-Broker-API-Version", "2.17")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode,
			"request without credentials should return 401")
	})

	t.Run("accepts authenticated request", func(t *testing.T) {
		t.Parallel()

		req, err := http.NewRequest(http.MethodGet, srv.URL+"/v2/catalog", nil)
		require.NoError(t, err)
		req.Header.Set("X-Broker-API-Version", "2.17")
		req.SetBasicAuth("admin", "secret")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode,
			"request with valid credentials should return 200")
	})
}

func TestWithLogger(t *testing.T) {
	t.Parallel()

	logger := &mockLogger{}

	// Smoke test: creating a handler with a logger must not panic.
	assert.NotPanics(t, func() {
		h := broker.NewHandler(newMockBroker(), broker.WithLogger(logger))
		assert.NotNil(t, h)
	})
}
