package server_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	osbapi "github.com/fivetwenty-io/osbapi/v2/pkg/osbapi"
)

// TestOriginatingIdentity_HeaderPresent_PopulatesContext verifies that a
// well-formed X-Broker-API-Originating-Identity header is decoded by the
// middleware and made available via OriginatingIdentityFromContext on the
// broker handler's context.
func TestOriginatingIdentity_HeaderPresent_PopulatesContext(t *testing.T) {
	t.Parallel()

	broker := &mockBroker{
		provisionResp: osbapi.ProvisionResponse{DashboardURL: "https://dash"},
	}

	ts := newTestServer(t, broker)
	defer ts.Close()

	encoded, err := osbapi.EncodeOriginatingIdentity(osbapi.OriginatingIdentity{
		Platform: "cloudfoundry",
		Value: map[string]any{
			"user_id": "user-123",
			"email":   "alice@example.com",
		},
	})
	require.NoError(t, err)

	req := osbapi.ProvisionRequest{ServiceID: "svc-1", PlanID: "plan-1"}

	resp := doRequestWithHeaders(t, http.MethodPut, ts.URL+"/v2/service_instances/inst-1", req, map[string]string{
		osbapi.HeaderOriginatingIdentity: encoded,
	})
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	require.NotNil(t, broker.lastProvisionCtx, "broker should have received a context")

	id, ok := osbapi.OriginatingIdentityFromContext(broker.lastProvisionCtx)
	require.True(t, ok, "context should carry originating identity")
	assert.Equal(t, "cloudfoundry", id.Platform)
	assert.Equal(t, "user-123", id.Value["user_id"])
	assert.Equal(t, "alice@example.com", id.Value["email"])
}

// TestOriginatingIdentity_HeaderAbsent_ContextUnchanged verifies that the
// middleware leaves the context untouched when the header is missing.
func TestOriginatingIdentity_HeaderAbsent_ContextUnchanged(t *testing.T) {
	t.Parallel()

	broker := &mockBroker{
		provisionResp: osbapi.ProvisionResponse{DashboardURL: "https://dash"},
	}

	ts := newTestServer(t, broker)
	defer ts.Close()

	req := osbapi.ProvisionRequest{ServiceID: "svc-1", PlanID: "plan-1"}
	resp := doRequest(t, http.MethodPut, ts.URL+"/v2/service_instances/inst-1", req)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.NotNil(t, broker.lastProvisionCtx)

	_, ok := osbapi.OriginatingIdentityFromContext(broker.lastProvisionCtx)
	assert.False(t, ok, "no header => no identity in context")
}

// TestOriginatingIdentity_HeaderMalformed_ContextUnchanged verifies the
// "lenient" path: a malformed header is dropped (not echoed as 400) so a
// misbehaving platform cannot block otherwise-valid requests. Brokers that
// require the identity must check OriginatingIdentityFromContext themselves.
func TestOriginatingIdentity_HeaderMalformed_ContextUnchanged(t *testing.T) {
	t.Parallel()

	broker := &mockBroker{
		provisionResp: osbapi.ProvisionResponse{DashboardURL: "https://dash"},
	}

	ts := newTestServer(t, broker)
	defer ts.Close()

	req := osbapi.ProvisionRequest{ServiceID: "svc-1", PlanID: "plan-1"}
	resp := doRequestWithHeaders(t, http.MethodPut, ts.URL+"/v2/service_instances/inst-1", req, map[string]string{
		// no space separator => DecodeOriginatingIdentity returns an error
		osbapi.HeaderOriginatingIdentity: "garbage-no-space",
	})
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.NotNil(t, broker.lastProvisionCtx)

	_, ok := osbapi.OriginatingIdentityFromContext(broker.lastProvisionCtx)
	assert.False(t, ok, "malformed header => identity dropped")
}

// TestContextWithOriginatingIdentity_RoundTrip exercises the public helpers
// directly (no HTTP) so brokers using them in tests have a known contract.
func TestContextWithOriginatingIdentity_RoundTrip(t *testing.T) {
	t.Parallel()

	id := osbapi.OriginatingIdentity{
		Platform: "kubernetes",
		Value: map[string]any{
			"username": "k8s-user",
			"groups":   []any{"group-1", "group-2"},
		},
	}

	ctx := osbapi.ContextWithOriginatingIdentity(t.Context(), id)

	got, ok := osbapi.OriginatingIdentityFromContext(ctx)
	require.True(t, ok)
	assert.Equal(t, id.Platform, got.Platform)

	// Compare via JSON to sidestep []any vs []string equality quirks.
	wantJSON, _ := json.Marshal(id.Value)
	gotJSON, _ := json.Marshal(got.Value)
	assert.JSONEq(t, string(wantJSON), string(gotJSON))
}
