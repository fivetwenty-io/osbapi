//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/fivetwenty-io/osbapi/v2/pkg/osbapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBinding_BindUnbind(t *testing.T) {
	t.Parallel()

	_, client := StartTestBroker(t)
	ctx := context.Background()
	instanceID := "binding-bind-unbind-001"
	bindingID := "binding-001"

	// Provision
	_, _, err := client.Provision(ctx, instanceID, osbapi.ProvisionRequest{
		ServiceID: testServiceID,
		PlanID:    freePlanID,
	}, false)
	require.NoError(t, err)

	// Bind
	bindResp, isAsync, err := client.Bind(ctx, instanceID, bindingID, osbapi.BindRequest{
		ServiceID: testServiceID,
		PlanID:    freePlanID,
	}, false)
	require.NoError(t, err)
	assert.False(t, isAsync, "sync broker should not return async")
	assert.NotEmpty(t, bindResp.Credentials, "bind should return credentials")

	// Unbind
	_, isAsync, err = client.Unbind(ctx, instanceID, bindingID, osbapi.UnbindRequest{
		ServiceID: testServiceID,
		PlanID:    freePlanID,
	}, false)
	require.NoError(t, err)
	assert.False(t, isAsync)

	// Deprovision
	_, _, err = client.Deprovision(ctx, instanceID, osbapi.DeprovisionRequest{
		ServiceID: testServiceID,
		PlanID:    freePlanID,
	}, false)
	require.NoError(t, err)
}

func TestBinding_BindGetBindingUnbind(t *testing.T) {
	t.Parallel()

	_, client := StartTestBroker(t)
	ctx := context.Background()
	instanceID := "binding-get-001"
	bindingID := "binding-get-b-001"

	// Provision
	_, _, err := client.Provision(ctx, instanceID, osbapi.ProvisionRequest{
		ServiceID: testServiceID,
		PlanID:    freePlanID,
	}, false)
	require.NoError(t, err)

	// Bind
	bindResp, _, err := client.Bind(ctx, instanceID, bindingID, osbapi.BindRequest{
		ServiceID: testServiceID,
		PlanID:    freePlanID,
	}, false)
	require.NoError(t, err)
	require.NotEmpty(t, bindResp.Credentials)

	// Get Binding
	fetchResp, err := client.GetBinding(ctx, instanceID, bindingID, osbapi.FetchBindingRequest{})
	require.NoError(t, err)
	assert.Equal(t, bindResp.Credentials, fetchResp.Credentials,
		"fetched binding credentials should match bind response")

	// Unbind
	_, _, err = client.Unbind(ctx, instanceID, bindingID, osbapi.UnbindRequest{
		ServiceID: testServiceID,
		PlanID:    freePlanID,
	}, false)
	require.NoError(t, err)

	// Deprovision
	_, _, err = client.Deprovision(ctx, instanceID, osbapi.DeprovisionRequest{
		ServiceID: testServiceID,
		PlanID:    freePlanID,
	}, false)
	require.NoError(t, err)
}

func TestBinding_BindCredentials(t *testing.T) {
	t.Parallel()

	_, client := StartTestBroker(t)
	ctx := context.Background()
	instanceID := "binding-creds-001"
	bindingID := "binding-creds-b-001"

	// Provision
	_, _, err := client.Provision(ctx, instanceID, osbapi.ProvisionRequest{
		ServiceID: testServiceID,
		PlanID:    freePlanID,
	}, false)
	require.NoError(t, err)

	// Bind
	bindResp, _, err := client.Bind(ctx, instanceID, bindingID, osbapi.BindRequest{
		ServiceID: testServiceID,
		PlanID:    freePlanID,
	}, false)
	require.NoError(t, err)

	// Verify credentials contain expected keys
	creds := bindResp.Credentials
	require.NotEmpty(t, creds, "credentials should not be empty")
	assert.Contains(t, creds, "uri", "credentials should contain uri")
	assert.Contains(t, creds, "username", "credentials should contain username")
	assert.Contains(t, creds, "password", "credentials should contain password")
	assert.Contains(t, creds, "database", "credentials should contain database")

	// Verify credential values are non-empty strings
	assert.NotEmpty(t, creds["uri"])
	assert.NotEmpty(t, creds["username"])
	assert.NotEmpty(t, creds["password"])
	assert.NotEmpty(t, creds["database"])

	// Cleanup: unbind + deprovision
	_, _, err = client.Unbind(ctx, instanceID, bindingID, osbapi.UnbindRequest{
		ServiceID: testServiceID,
		PlanID:    freePlanID,
	}, false)
	require.NoError(t, err)

	_, _, err = client.Deprovision(ctx, instanceID, osbapi.DeprovisionRequest{
		ServiceID: testServiceID,
		PlanID:    freePlanID,
	}, false)
	require.NoError(t, err)
}

func TestBinding_UnbindNonExistent(t *testing.T) {
	t.Parallel()

	_, client := StartTestBroker(t)
	ctx := context.Background()
	instanceID := "binding-unbind-gone-001"
	bindingID := "binding-gone-b-001"

	// Provision an instance first (unbind needs a valid instance path)
	_, _, err := client.Provision(ctx, instanceID, osbapi.ProvisionRequest{
		ServiceID: testServiceID,
		PlanID:    freePlanID,
	}, false)
	require.NoError(t, err)

	// Unbind a non-existent binding
	_, _, err = client.Unbind(ctx, instanceID, bindingID, osbapi.UnbindRequest{
		ServiceID: testServiceID,
		PlanID:    freePlanID,
	}, false)
	require.Error(t, err)
	assert.True(t, osbapi.IsGone(err), "unbinding non-existent binding should return 410 Gone, got: %v", err)

	// Cleanup
	_, _, err = client.Deprovision(ctx, instanceID, osbapi.DeprovisionRequest{
		ServiceID: testServiceID,
		PlanID:    freePlanID,
	}, false)
	require.NoError(t, err)
}
