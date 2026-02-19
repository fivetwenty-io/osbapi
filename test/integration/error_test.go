//go:build integration

package integration

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/fivetwenty-io/osbapi/v2/pkg/osbapi"
	"github.com/fivetwenty-io/osbapi/v2/pkg/osbclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestError_MissingVersionHeader(t *testing.T) {
	t.Parallel()

	srv, _ := StartTestBroker(t)

	// Make a raw HTTP request without the X-Broker-API-Version header.
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		srv.URL+"/v2/catalog",
		nil,
	)
	require.NoError(t, err)
	req.SetBasicAuth(testUsername, testPassword)
	// Deliberately do NOT set X-Broker-API-Version header.

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	assert.Equal(t, http.StatusPreconditionFailed, resp.StatusCode,
		"missing version header should return 412 Precondition Failed")
}

func TestError_InvalidAuth(t *testing.T) {
	t.Parallel()

	srv, _ := StartTestBroker(t)

	// Create a client with wrong credentials.
	badClient, err := osbclient.NewWithBasicAuth(srv.URL, "wrong-user", "wrong-pass")
	require.NoError(t, err)

	_, err = badClient.GetCatalog(context.Background())
	require.Error(t, err)

	var osbErr *osbapi.OSBError
	if assert.ErrorAs(t, err, &osbErr) {
		assert.Equal(t, http.StatusUnauthorized, osbErr.StatusCode,
			"invalid auth should return 401 Unauthorized")
	}
}

func TestError_GetNonExistentInstance(t *testing.T) {
	t.Parallel()

	_, client := StartTestBroker(t)
	ctx := context.Background()

	_, err := client.GetInstance(ctx, "does-not-exist", osbapi.FetchInstanceRequest{})
	require.Error(t, err)
	assert.True(t, osbapi.IsNotFound(err),
		"fetching non-existent instance should return not found, got: %v", err)
}

func TestError_GetNonExistentBinding(t *testing.T) {
	t.Parallel()

	_, client := StartTestBroker(t)
	ctx := context.Background()
	instanceID := "error-binding-inst-001"

	// Provision an instance so the binding fetch has a valid instance context.
	_, _, err := client.Provision(ctx, instanceID, osbapi.ProvisionRequest{
		ServiceID: testServiceID,
		PlanID:    freePlanID,
	}, false)
	require.NoError(t, err)

	_, err = client.GetBinding(ctx, instanceID, "does-not-exist", osbapi.FetchBindingRequest{})
	require.Error(t, err)
	assert.True(t, osbapi.IsNotFound(err),
		"fetching non-existent binding should return not found, got: %v", err)

	// Cleanup
	_, _, err = client.Deprovision(ctx, instanceID, osbapi.DeprovisionRequest{
		ServiceID: testServiceID,
		PlanID:    freePlanID,
	}, false)
	require.NoError(t, err)
}
