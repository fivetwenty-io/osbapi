//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/fivetwenty-io/osbapi/v2/pkg/osbapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLifecycle_ProvisionDeprovision(t *testing.T) {
	t.Parallel()

	_, client := StartTestBroker(t)
	ctx := context.Background()
	instanceID := "lifecycle-prov-deprov-001"

	// Provision
	provResp, isAsync, err := client.Provision(ctx, instanceID, osbapi.ProvisionRequest{
		ServiceID: testServiceID,
		PlanID:    freePlanID,
	}, false)
	require.NoError(t, err)
	assert.False(t, isAsync, "sync broker should not return async")
	assert.NotEmpty(t, provResp.DashboardURL)

	// Deprovision
	_, isAsync, err = client.Deprovision(ctx, instanceID, osbapi.DeprovisionRequest{
		ServiceID: testServiceID,
		PlanID:    freePlanID,
	}, false)
	require.NoError(t, err)
	assert.False(t, isAsync)
}

func TestLifecycle_ProvisionGetInstanceDeprovision(t *testing.T) {
	t.Parallel()

	_, client := StartTestBroker(t)
	ctx := context.Background()
	instanceID := "lifecycle-get-inst-001"

	// Provision
	provResp, _, err := client.Provision(ctx, instanceID, osbapi.ProvisionRequest{
		ServiceID: testServiceID,
		PlanID:    freePlanID,
	}, false)
	require.NoError(t, err)
	assert.NotEmpty(t, provResp.DashboardURL)

	// Get Instance
	fetchResp, err := client.GetInstance(ctx, instanceID, osbapi.FetchInstanceRequest{})
	require.NoError(t, err)
	assert.Equal(t, testServiceID, fetchResp.ServiceID)
	assert.Equal(t, freePlanID, fetchResp.PlanID)
	assert.Equal(t, provResp.DashboardURL, fetchResp.DashboardURL)

	// Deprovision
	_, _, err = client.Deprovision(ctx, instanceID, osbapi.DeprovisionRequest{
		ServiceID: testServiceID,
		PlanID:    freePlanID,
	}, false)
	require.NoError(t, err)
}

func TestLifecycle_ProvisionUpdateDeprovision(t *testing.T) {
	t.Parallel()

	_, client := StartTestBroker(t)
	ctx := context.Background()
	instanceID := "lifecycle-update-001"

	// Provision with free plan
	_, _, err := client.Provision(ctx, instanceID, osbapi.ProvisionRequest{
		ServiceID: testServiceID,
		PlanID:    freePlanID,
	}, false)
	require.NoError(t, err)

	// Update to paid plan
	updateResp, isAsync, err := client.Update(ctx, instanceID, osbapi.UpdateRequest{
		ServiceID: testServiceID,
		PlanID:    paidPlanID,
	}, false)
	require.NoError(t, err)
	assert.False(t, isAsync)
	assert.NotEmpty(t, updateResp.DashboardURL)

	// Verify the plan was updated
	fetchResp, err := client.GetInstance(ctx, instanceID, osbapi.FetchInstanceRequest{})
	require.NoError(t, err)
	assert.Equal(t, paidPlanID, fetchResp.PlanID)

	// Deprovision
	_, _, err = client.Deprovision(ctx, instanceID, osbapi.DeprovisionRequest{
		ServiceID: testServiceID,
		PlanID:    paidPlanID,
	}, false)
	require.NoError(t, err)
}

func TestLifecycle_ProvisionIdempotent(t *testing.T) {
	t.Parallel()

	_, client := StartTestBroker(t)
	ctx := context.Background()
	instanceID := "lifecycle-idempotent-001"

	req := osbapi.ProvisionRequest{
		ServiceID: testServiceID,
		PlanID:    freePlanID,
	}

	// First provision
	resp1, _, err := client.Provision(ctx, instanceID, req, false)
	require.NoError(t, err)
	assert.NotEmpty(t, resp1.DashboardURL)

	// Second provision with same parameters should succeed (idempotent)
	resp2, _, err := client.Provision(ctx, instanceID, req, false)
	require.NoError(t, err)
	assert.Equal(t, resp1.DashboardURL, resp2.DashboardURL)

	// Cleanup
	_, _, err = client.Deprovision(ctx, instanceID, osbapi.DeprovisionRequest{
		ServiceID: testServiceID,
		PlanID:    freePlanID,
	}, false)
	require.NoError(t, err)
}

func TestLifecycle_DeprovisionNonExistent(t *testing.T) {
	t.Parallel()

	_, client := StartTestBroker(t)
	ctx := context.Background()

	_, _, err := client.Deprovision(ctx, "non-existent-instance", osbapi.DeprovisionRequest{
		ServiceID: testServiceID,
		PlanID:    freePlanID,
	}, false)
	require.Error(t, err)
	assert.True(t, osbapi.IsGone(err), "deprovisioning non-existent instance should return 410 Gone, got: %v", err)
}
