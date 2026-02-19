package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	osbapi "github.com/fivetwenty-io/osbapi/v2/pkg/osbapi"

	"github.com/fivetwenty-io/osbapi/v2/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// newTestServer creates an httptest.Server backed by the given mockBroker and
// any additional handler options.
func newTestServer(t *testing.T, broker *mockBroker, opts ...server.Option) *httptest.Server {
	t.Helper()
	h := server.NewHandler(broker, opts...)
	return httptest.NewServer(h)
}

// doRequest builds and executes an HTTP request against the given server,
// automatically adding the required API version header.
func doRequest(t *testing.T, method, url string, body any) *http.Response {
	t.Helper()
	return doRequestWithHeaders(t, method, url, body, nil)
}

// doRequestWithHeaders is like doRequest but allows extra headers.
func doRequestWithHeaders(t *testing.T, method, url string, body any, headers map[string]string) *http.Response {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}

	req, err := http.NewRequest(method, url, &buf)
	require.NoError(t, err)

	// Default version header; callers can override via headers map.
	req.Header.Set(osbapi.HeaderAPIVersion, osbapi.APIVersion)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	return resp
}

// decodeBody is a generic JSON decoder for response bodies.
func decodeBody[T any](t *testing.T, resp *http.Response) T {
	t.Helper()
	defer resp.Body.Close()

	var v T
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&v))
	return v
}

// ---------------------------------------------------------------------------
// Catalog
// ---------------------------------------------------------------------------

func TestCatalogHandler(t *testing.T) {
	t.Parallel()

	broker := &mockBroker{
		catalog: &osbapi.Catalog{
			Services: []osbapi.Service{
				{
					ID:          "svc-1",
					Name:        "test-service",
					Description: "A test service",
					Bindable:    true,
					Plans: []osbapi.Plan{
						{ID: "plan-1", Name: "default", Description: "Default plan"},
					},
				},
			},
		},
	}

	ts := newTestServer(t, broker)
	defer ts.Close()

	resp := doRequest(t, http.MethodGet, ts.URL+"/v2/catalog", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	catalog := decodeBody[osbapi.Catalog](t, resp)
	require.Len(t, catalog.Services, 1)
	assert.Equal(t, "svc-1", catalog.Services[0].ID)
	assert.Equal(t, "test-service", catalog.Services[0].Name)
}

// ---------------------------------------------------------------------------
// Provision
// ---------------------------------------------------------------------------

func TestProvisionHandler_Sync(t *testing.T) {
	t.Parallel()

	broker := &mockBroker{
		provisionResp: osbapi.ProvisionResponse{
			DashboardURL: "https://dashboard.example.com",
		},
		provisionAsync: false,
	}

	ts := newTestServer(t, broker)
	defer ts.Close()

	body := osbapi.ProvisionRequest{
		ServiceID:        "svc-1",
		PlanID:           "plan-1",
		OrganizationGUID: "org-guid",
		SpaceGUID:        "space-guid",
	}

	resp := doRequest(t, http.MethodPut, ts.URL+"/v2/service_instances/abc", body)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	result := decodeBody[osbapi.ProvisionResponse](t, resp)
	assert.Equal(t, "https://dashboard.example.com", result.DashboardURL)
	assert.Equal(t, "abc", broker.lastInstanceID)
	assert.Equal(t, "svc-1", broker.lastProvisionReq.ServiceID)
}

func TestProvisionHandler_Async(t *testing.T) {
	t.Parallel()

	broker := &mockBroker{
		provisionResp: osbapi.ProvisionResponse{
			Operation: "provision-op-1",
		},
		provisionAsync: true,
	}

	ts := newTestServer(t, broker)
	defer ts.Close()

	body := osbapi.ProvisionRequest{
		ServiceID:        "svc-1",
		PlanID:           "plan-1",
		OrganizationGUID: "org-guid",
		SpaceGUID:        "space-guid",
	}

	resp := doRequest(t, http.MethodPut, ts.URL+"/v2/service_instances/abc?accepts_incomplete=true", body)
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)

	result := decodeBody[osbapi.ProvisionResponse](t, resp)
	assert.Equal(t, "provision-op-1", result.Operation)
}

// ---------------------------------------------------------------------------
// Deprovision
// ---------------------------------------------------------------------------

func TestDeprovisionHandler_Sync(t *testing.T) {
	t.Parallel()

	broker := &mockBroker{
		deprovisionAsync: false,
	}

	ts := newTestServer(t, broker)
	defer ts.Close()

	resp := doRequest(t, http.MethodDelete,
		ts.URL+"/v2/service_instances/abc?service_id=svc-1&plan_id=plan-1", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	assert.Equal(t, "abc", broker.lastInstanceID)
	assert.Equal(t, "svc-1", broker.lastDeprovisionReq.ServiceID)
	assert.Equal(t, "plan-1", broker.lastDeprovisionReq.PlanID)
}

// ---------------------------------------------------------------------------
// FetchInstance
// ---------------------------------------------------------------------------

func TestFetchInstanceHandler(t *testing.T) {
	t.Parallel()

	broker := &mockBroker{
		fetchInstanceResp: osbapi.FetchInstanceResponse{
			ServiceID:    "svc-1",
			PlanID:       "plan-1",
			DashboardURL: "https://dashboard.example.com",
		},
	}

	ts := newTestServer(t, broker)
	defer ts.Close()

	resp := doRequest(t, http.MethodGet, ts.URL+"/v2/service_instances/abc", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	result := decodeBody[osbapi.FetchInstanceResponse](t, resp)
	assert.Equal(t, "svc-1", result.ServiceID)
	assert.Equal(t, "abc", broker.lastInstanceID)
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestUpdateHandler_Sync(t *testing.T) {
	t.Parallel()

	broker := &mockBroker{
		updateResp: osbapi.UpdateResponse{
			DashboardURL: "https://dashboard.example.com/updated",
		},
		updateAsync: false,
	}

	ts := newTestServer(t, broker)
	defer ts.Close()

	body := osbapi.UpdateRequest{
		ServiceID: "svc-1",
		PlanID:    "plan-2",
	}

	resp := doRequest(t, http.MethodPatch, ts.URL+"/v2/service_instances/abc", body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	result := decodeBody[osbapi.UpdateResponse](t, resp)
	assert.Equal(t, "https://dashboard.example.com/updated", result.DashboardURL)
	assert.Equal(t, "abc", broker.lastInstanceID)
}

func TestUpdateHandler_Async(t *testing.T) {
	t.Parallel()

	broker := &mockBroker{
		updateResp: osbapi.UpdateResponse{
			Operation: "update-op-1",
		},
		updateAsync: true,
	}

	ts := newTestServer(t, broker)
	defer ts.Close()

	body := osbapi.UpdateRequest{
		ServiceID: "svc-1",
		PlanID:    "plan-2",
	}

	resp := doRequest(t, http.MethodPatch, ts.URL+"/v2/service_instances/abc?accepts_incomplete=true", body)
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)

	result := decodeBody[osbapi.UpdateResponse](t, resp)
	assert.Equal(t, "update-op-1", result.Operation)
}

// ---------------------------------------------------------------------------
// LastOperation
// ---------------------------------------------------------------------------

func TestLastOperationHandler(t *testing.T) {
	t.Parallel()

	broker := &mockBroker{
		lastOpResp: osbapi.LastOperationResponse{
			State:       osbapi.StateSucceeded,
			Description: "done",
		},
	}

	ts := newTestServer(t, broker)
	defer ts.Close()

	resp := doRequest(t, http.MethodGet,
		ts.URL+"/v2/service_instances/abc/last_operation?service_id=svc-1&plan_id=plan-1&operation=provision-op-1",
		nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	result := decodeBody[osbapi.LastOperationResponse](t, resp)
	assert.Equal(t, osbapi.StateSucceeded, result.State)
	assert.Equal(t, "done", result.Description)
	assert.Equal(t, "abc", broker.lastInstanceID)
	assert.Equal(t, "provision-op-1", broker.lastLastOpReq.Operation)
}

// ---------------------------------------------------------------------------
// Middleware: API version header
// ---------------------------------------------------------------------------

func TestVersionHeaderRequired(t *testing.T) {
	t.Parallel()

	broker := &mockBroker{
		catalog: &osbapi.Catalog{},
	}

	ts := newTestServer(t, broker)
	defer ts.Close()

	// Send a request without the version header.
	req, err := http.NewRequest(http.MethodGet, ts.URL+"/v2/catalog", nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusPreconditionFailed, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// Middleware: Basic auth
// ---------------------------------------------------------------------------

func TestBasicAuthRequired(t *testing.T) {
	t.Parallel()

	broker := &mockBroker{
		catalog: &osbapi.Catalog{},
	}

	ts := newTestServer(t, broker, server.WithBasicAuth("admin", "secret"))
	defer ts.Close()

	// Request with wrong credentials.
	req, err := http.NewRequest(http.MethodGet, ts.URL+"/v2/catalog", nil)
	require.NoError(t, err)
	req.Header.Set(osbapi.HeaderAPIVersion, osbapi.APIVersion)
	req.SetBasicAuth("wrong", "creds")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// Error mapping: ErrAsyncRequired -> 422
// ---------------------------------------------------------------------------

func TestAsyncRequiredError(t *testing.T) {
	t.Parallel()

	broker := &mockBroker{
		provisionErr: osbapi.ErrAsyncRequired,
	}

	ts := newTestServer(t, broker)
	defer ts.Close()

	body := osbapi.ProvisionRequest{
		ServiceID:        "svc-1",
		PlanID:           "plan-1",
		OrganizationGUID: "org",
		SpaceGUID:        "space",
	}

	resp := doRequest(t, http.MethodPut, ts.URL+"/v2/service_instances/abc", body)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)

	osbErr := decodeBody[osbapi.OSBError](t, resp)
	assert.Equal(t, osbapi.ErrorCodeAsyncRequired, osbErr.ErrorCode)
}

// ---------------------------------------------------------------------------
// Error mapping: ErrInstanceNotFound -> 404
// ---------------------------------------------------------------------------

func TestNotFoundError(t *testing.T) {
	t.Parallel()

	broker := &mockBroker{
		fetchInstanceErr: osbapi.ErrInstanceNotFound,
	}

	ts := newTestServer(t, broker)
	defer ts.Close()

	resp := doRequest(t, http.MethodGet, ts.URL+"/v2/service_instances/abc", nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
