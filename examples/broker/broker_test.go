package main

import (
	"context"
	"testing"

	"github.com/fivetwenty-io/osbapi/v2/pkg/osbapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryBroker_GetCatalog(t *testing.T) {
	t.Parallel()

	b := NewInMemoryBroker()
	catalog, err := b.GetCatalog(context.Background())

	require.NoError(t, err)
	require.NotNil(t, catalog)
	require.Len(t, catalog.Services, 1)

	svc := catalog.Services[0]
	assert.Equal(t, "example-service", svc.Name)
	assert.True(t, svc.Bindable)
	require.Len(t, svc.Plans, 2)

	freePlan := svc.Plans[0]
	assert.Equal(t, "free", freePlan.Name)
	require.NotNil(t, freePlan.Free)
	assert.True(t, *freePlan.Free)

	paidPlan := svc.Plans[1]
	assert.Equal(t, "paid", paidPlan.Name)
	require.NotNil(t, paidPlan.Free)
	assert.False(t, *paidPlan.Free)
}

func TestInMemoryBroker_ProvisionAndGetInstance(t *testing.T) {
	t.Parallel()

	b := NewInMemoryBroker()
	ctx := context.Background()
	instanceID := "test-instance-001"

	provResp, isAsync, err := b.Provision(ctx, instanceID, osbapi.ProvisionRequest{
		ServiceID:  exampleServiceID,
		PlanID:     freePlanID,
		Parameters: map[string]any{"env": "test"},
	}, false)

	require.NoError(t, err)
	assert.False(t, isAsync)
	assert.Contains(t, provResp.DashboardURL, instanceID)

	fetchResp, err := b.GetInstance(ctx, instanceID, osbapi.FetchInstanceRequest{})

	require.NoError(t, err)
	assert.Equal(t, exampleServiceID, fetchResp.ServiceID)
	assert.Equal(t, freePlanID, fetchResp.PlanID)
	assert.Equal(t, provResp.DashboardURL, fetchResp.DashboardURL)
	assert.Equal(t, "test", fetchResp.Parameters["env"])
}

func TestInMemoryBroker_ProvisionIdempotent(t *testing.T) {
	t.Parallel()

	b := NewInMemoryBroker()
	ctx := context.Background()
	instanceID := "idempotent-instance"

	req := osbapi.ProvisionRequest{
		ServiceID: exampleServiceID,
		PlanID:    freePlanID,
	}

	resp1, _, err := b.Provision(ctx, instanceID, req, false)
	require.NoError(t, err)

	resp2, _, err := b.Provision(ctx, instanceID, req, false)
	require.NoError(t, err)
	assert.Equal(t, resp1.DashboardURL, resp2.DashboardURL)

	// Provision with a different plan should fail.
	_, _, err = b.Provision(ctx, instanceID, osbapi.ProvisionRequest{
		ServiceID: exampleServiceID,
		PlanID:    paidPlanID,
	}, false)
	assert.ErrorIs(t, err, osbapi.ErrInstanceAlreadyExists)
}

func TestInMemoryBroker_DeprovisionNotFound(t *testing.T) {
	t.Parallel()

	b := NewInMemoryBroker()

	_, _, err := b.Deprovision(context.Background(), "nonexistent", osbapi.DeprovisionRequest{
		ServiceID: exampleServiceID,
		PlanID:    freePlanID,
	}, false)

	assert.ErrorIs(t, err, osbapi.ErrGoneError)
}

func TestInMemoryBroker_BindAndGetBinding(t *testing.T) {
	t.Parallel()

	b := NewInMemoryBroker()
	ctx := context.Background()
	instanceID := "bind-instance"
	bindingID := "bind-001"

	_, _, err := b.Provision(ctx, instanceID, osbapi.ProvisionRequest{
		ServiceID: exampleServiceID,
		PlanID:    freePlanID,
	}, false)
	require.NoError(t, err)

	bindResp, isAsync, err := b.Bind(ctx, instanceID, bindingID, osbapi.BindRequest{
		ServiceID: exampleServiceID,
		PlanID:    freePlanID,
	}, false)

	require.NoError(t, err)
	assert.False(t, isAsync)
	require.NotNil(t, bindResp.Credentials)
	assert.NotEmpty(t, bindResp.Credentials["uri"])
	assert.NotEmpty(t, bindResp.Credentials["username"])
	assert.NotEmpty(t, bindResp.Credentials["password"])
	assert.NotEmpty(t, bindResp.Credentials["database"])

	fetchResp, err := b.GetBinding(ctx, instanceID, bindingID, osbapi.FetchBindingRequest{})

	require.NoError(t, err)
	assert.Equal(t, bindResp.Credentials, fetchResp.Credentials)
}

func TestInMemoryBroker_UnbindNotFound(t *testing.T) {
	t.Parallel()

	b := NewInMemoryBroker()

	_, _, err := b.Unbind(context.Background(), "nonexistent", "no-binding", osbapi.UnbindRequest{
		ServiceID: exampleServiceID,
		PlanID:    freePlanID,
	}, false)

	assert.ErrorIs(t, err, osbapi.ErrGoneError)
}

func TestInMemoryBroker_UpdatePlan(t *testing.T) {
	t.Parallel()

	b := NewInMemoryBroker()
	ctx := context.Background()
	instanceID := "update-instance"

	_, _, err := b.Provision(ctx, instanceID, osbapi.ProvisionRequest{
		ServiceID: exampleServiceID,
		PlanID:    freePlanID,
	}, false)
	require.NoError(t, err)

	_, isAsync, err := b.Update(ctx, instanceID, osbapi.UpdateRequest{
		ServiceID: exampleServiceID,
		PlanID:    paidPlanID,
	}, false)

	require.NoError(t, err)
	assert.False(t, isAsync)

	fetchResp, err := b.GetInstance(ctx, instanceID, osbapi.FetchInstanceRequest{})

	require.NoError(t, err)
	assert.Equal(t, paidPlanID, fetchResp.PlanID)
}
