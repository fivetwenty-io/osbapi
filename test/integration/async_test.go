//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/fivetwenty-io/osbapi/v2/pkg/osbapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	asyncServiceID = "async-service-1"
	asyncPlanID    = "async-plan-1"
	asyncDelay     = 50 * time.Millisecond
)

// TestAsync_ProvisionAndPoll provisions with async=true, gets 202,
// then polls until the operation succeeds.
func TestAsync_ProvisionAndPoll(t *testing.T) {
	t.Parallel()

	_, client := StartAsyncTestBroker(t, asyncDelay)
	ctx := context.Background()
	instanceID := "async-prov-" + t.Name()

	// Provision with async=true.
	provResp, isAsync, err := client.Provision(ctx, instanceID, osbapi.ProvisionRequest{
		ServiceID:        asyncServiceID,
		PlanID:           asyncPlanID,
		OrganizationGUID: "org-1",
		SpaceGUID:        "space-1",
	}, true)
	require.NoError(t, err)
	assert.True(t, isAsync, "expected async response (202)")
	assert.NotEmpty(t, provResp.Operation, "expected operation token")

	// Immediately poll -- should be in progress.
	pollResp, err := client.PollLastOperation(ctx, instanceID, osbapi.LastOperationRequest{
		Operation: provResp.Operation,
	})
	require.NoError(t, err)
	assert.Equal(t, osbapi.StateInProgress, pollResp.State)

	// Wait for the async operation to complete, then poll again.
	result, err := client.PollInstanceUntilComplete(ctx, instanceID, osbapi.PollConfig{
		Interval: 10 * time.Millisecond,
		Timeout:  2 * time.Second,
	})
	require.NoError(t, err)
	assert.Equal(t, osbapi.StateSucceeded, result.State)

	// Verify instance is now fetchable.
	instResp, err := client.GetInstance(ctx, instanceID, osbapi.FetchInstanceRequest{})
	require.NoError(t, err)
	assert.Equal(t, asyncServiceID, instResp.ServiceID)
	assert.Equal(t, asyncPlanID, instResp.PlanID)
}

// TestAsync_ProvisionFails_WithoutAsync provisions with async=false and
// expects an AsyncRequired error.
func TestAsync_ProvisionFails_WithoutAsync(t *testing.T) {
	t.Parallel()

	_, client := StartAsyncTestBroker(t, asyncDelay)
	ctx := context.Background()
	instanceID := "async-no-async-" + t.Name()

	_, _, err := client.Provision(ctx, instanceID, osbapi.ProvisionRequest{
		ServiceID:        asyncServiceID,
		PlanID:           asyncPlanID,
		OrganizationGUID: "org-1",
		SpaceGUID:        "space-1",
	}, false)
	require.Error(t, err)
	assert.True(t, osbapi.IsAsyncRequired(err), "expected AsyncRequired error, got: %v", err)
}

// TestAsync_DeprovisionAndPoll provisions (async), waits, then deprovisions
// (async) and polls until succeeded.
func TestAsync_DeprovisionAndPoll(t *testing.T) {
	t.Parallel()

	_, client := StartAsyncTestBroker(t, asyncDelay)
	ctx := context.Background()
	instanceID := "async-deprov-" + t.Name()

	// Provision first.
	_, isAsync, err := client.Provision(ctx, instanceID, osbapi.ProvisionRequest{
		ServiceID:        asyncServiceID,
		PlanID:           asyncPlanID,
		OrganizationGUID: "org-1",
		SpaceGUID:        "space-1",
	}, true)
	require.NoError(t, err)
	require.True(t, isAsync)

	// Wait for provisioning to complete.
	_, err = client.PollInstanceUntilComplete(ctx, instanceID, osbapi.PollConfig{
		Interval: 10 * time.Millisecond,
		Timeout:  2 * time.Second,
	})
	require.NoError(t, err)

	// Deprovision with async=true.
	deprovResp, isAsync, err := client.Deprovision(ctx, instanceID, osbapi.DeprovisionRequest{
		ServiceID: asyncServiceID,
		PlanID:    asyncPlanID,
	}, true)
	require.NoError(t, err)
	assert.True(t, isAsync, "expected async deprovision (202)")
	assert.NotEmpty(t, deprovResp.Operation, "expected operation token")

	// Poll deprovision until succeeded.
	// Note: PollInstanceUntilComplete uses empty LastOperationRequest by default,
	// so we poll manually with the operation token.
	pollResp, err := client.PollLastOperation(ctx, instanceID, osbapi.LastOperationRequest{
		Operation: deprovResp.Operation,
	})
	require.NoError(t, err)
	// May still be in progress or already succeeded depending on timing.
	assert.Contains(t,
		[]osbapi.OperationState{osbapi.StateInProgress, osbapi.StateSucceeded},
		pollResp.State,
	)

	// Wait enough time for deprovision to complete.
	time.Sleep(asyncDelay + 20*time.Millisecond)

	pollResp, err = client.PollLastOperation(ctx, instanceID, osbapi.LastOperationRequest{
		Operation: deprovResp.Operation,
	})
	require.NoError(t, err)
	assert.Equal(t, osbapi.StateSucceeded, pollResp.State)

	// Instance should no longer be fetchable.
	_, err = client.GetInstance(ctx, instanceID, osbapi.FetchInstanceRequest{})
	require.Error(t, err)
	assert.True(t, osbapi.IsNotFound(err), "expected not found error, got: %v", err)

	// Deprovision without async should fail.
	_, _, err = client.Deprovision(ctx, instanceID, osbapi.DeprovisionRequest{
		ServiceID: asyncServiceID,
		PlanID:    asyncPlanID,
	}, false)
	require.Error(t, err)
	assert.True(t, osbapi.IsAsyncRequired(err), "expected AsyncRequired error, got: %v", err)
}

// TestAsync_BindAndPoll provisions (async), waits, then binds (async)
// and polls the binding operation until succeeded.
func TestAsync_BindAndPoll(t *testing.T) {
	t.Parallel()

	_, client := StartAsyncTestBroker(t, asyncDelay)
	ctx := context.Background()
	instanceID := "async-bind-" + t.Name()
	bindingID := "binding-1"

	// Provision first.
	_, _, err := client.Provision(ctx, instanceID, osbapi.ProvisionRequest{
		ServiceID:        asyncServiceID,
		PlanID:           asyncPlanID,
		OrganizationGUID: "org-1",
		SpaceGUID:        "space-1",
	}, true)
	require.NoError(t, err)

	_, err = client.PollInstanceUntilComplete(ctx, instanceID, osbapi.PollConfig{
		Interval: 10 * time.Millisecond,
		Timeout:  2 * time.Second,
	})
	require.NoError(t, err)

	// Bind with async=true.
	bindResp, isAsync, err := client.Bind(ctx, instanceID, bindingID, osbapi.BindRequest{
		ServiceID: asyncServiceID,
		PlanID:    asyncPlanID,
	}, true)
	require.NoError(t, err)
	assert.True(t, isAsync, "expected async bind (202)")
	assert.NotEmpty(t, bindResp.Operation, "expected operation token")

	// Poll binding operation until succeeded.
	result, err := client.PollBindingUntilComplete(ctx, instanceID, bindingID, osbapi.PollConfig{
		Interval: 10 * time.Millisecond,
		Timeout:  2 * time.Second,
	})
	require.NoError(t, err)
	assert.Equal(t, osbapi.StateSucceeded, result.State)

	// Verify binding is fetchable with credentials.
	fetchResp, err := client.GetBinding(ctx, instanceID, bindingID, osbapi.FetchBindingRequest{})
	require.NoError(t, err)
	assert.Equal(t, "async-user", fetchResp.Credentials["username"])
	assert.Equal(t, "async-pass", fetchResp.Credentials["password"])

	// Bind without async should fail.
	_, _, err = client.Bind(ctx, instanceID, "binding-no-async", osbapi.BindRequest{
		ServiceID: asyncServiceID,
		PlanID:    asyncPlanID,
	}, false)
	require.Error(t, err)
	assert.True(t, osbapi.IsAsyncRequired(err), "expected AsyncRequired error, got: %v", err)
}

// TestAsync_UnbindAndPoll exercises the full async lifecycle:
// provision -> bind -> unbind (async) -> poll -> deprovision.
func TestAsync_UnbindAndPoll(t *testing.T) {
	t.Parallel()

	_, client := StartAsyncTestBroker(t, asyncDelay)
	ctx := context.Background()
	instanceID := "async-unbind-" + t.Name()
	bindingID := "binding-unbind-1"

	// 1. Provision (async).
	_, _, err := client.Provision(ctx, instanceID, osbapi.ProvisionRequest{
		ServiceID:        asyncServiceID,
		PlanID:           asyncPlanID,
		OrganizationGUID: "org-1",
		SpaceGUID:        "space-1",
	}, true)
	require.NoError(t, err)

	_, err = client.PollInstanceUntilComplete(ctx, instanceID, osbapi.PollConfig{
		Interval: 10 * time.Millisecond,
		Timeout:  2 * time.Second,
	})
	require.NoError(t, err)

	// 2. Bind (async).
	bindResp, _, err := client.Bind(ctx, instanceID, bindingID, osbapi.BindRequest{
		ServiceID: asyncServiceID,
		PlanID:    asyncPlanID,
	}, true)
	require.NoError(t, err)

	_, err = client.PollBindingUntilComplete(ctx, instanceID, bindingID, osbapi.PollConfig{
		Interval: 10 * time.Millisecond,
		Timeout:  2 * time.Second,
	})
	require.NoError(t, err)

	// Verify binding exists.
	_, err = client.GetBinding(ctx, instanceID, bindingID, osbapi.FetchBindingRequest{})
	require.NoError(t, err)

	// 3. Unbind (async).
	unbindResp, isAsync, err := client.Unbind(ctx, instanceID, bindingID, osbapi.UnbindRequest{
		ServiceID: asyncServiceID,
		PlanID:    asyncPlanID,
	}, true)
	require.NoError(t, err)
	assert.True(t, isAsync, "expected async unbind (202)")
	assert.NotEmpty(t, unbindResp.Operation, "expected operation token")

	// Poll unbind until succeeded.
	result, err := client.PollBindingUntilComplete(ctx, instanceID, bindingID, osbapi.PollConfig{
		Interval: 10 * time.Millisecond,
		Timeout:  2 * time.Second,
	})
	require.NoError(t, err)
	assert.Equal(t, osbapi.StateSucceeded, result.State)

	// Binding should no longer be fetchable.
	_, err = client.GetBinding(ctx, instanceID, bindingID, osbapi.FetchBindingRequest{})
	require.Error(t, err)
	assert.True(t, osbapi.IsNotFound(err), "expected not found error, got: %v", err)

	// 4. Deprovision (async).
	deprovResp, isAsync, err := client.Deprovision(ctx, instanceID, osbapi.DeprovisionRequest{
		ServiceID: asyncServiceID,
		PlanID:    asyncPlanID,
	}, true)
	require.NoError(t, err)
	assert.True(t, isAsync)

	// Wait for deprovision.
	time.Sleep(asyncDelay + 20*time.Millisecond)

	pollResp, err := client.PollLastOperation(ctx, instanceID, osbapi.LastOperationRequest{
		Operation: deprovResp.Operation,
	})
	require.NoError(t, err)
	assert.Equal(t, osbapi.StateSucceeded, pollResp.State)

	// Instance gone.
	_, err = client.GetInstance(ctx, instanceID, osbapi.FetchInstanceRequest{})
	require.Error(t, err)
	assert.True(t, osbapi.IsNotFound(err), "expected not found error, got: %v", err)

	// Suppress unused variable warning for bindResp.
	_ = bindResp
}
