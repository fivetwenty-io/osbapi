package osbapi_test

import (
	"context"
	"testing"

	"github.com/fivetwenty-io/osbapi/v2/pkg/osbapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockBroker implements osbapi.ServiceBroker for compilation verification.
type mockBroker struct{}

// Compile-time interface satisfaction check.
var _ osbapi.ServiceBroker = (*mockBroker)(nil)

func (m *mockBroker) GetCatalog(_ context.Context) (*osbapi.Catalog, error) {
	return &osbapi.Catalog{}, nil
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
	return osbapi.LastOperationResponse{}, nil
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
	return osbapi.LastOperationResponse{}, nil
}

func TestBrokerInterface_Satisfiable(t *testing.T) {
	t.Parallel()

	var broker osbapi.ServiceBroker = &mockBroker{}
	require.NotNil(t, broker, "mockBroker must satisfy ServiceBroker")
}

func TestBrokerInterface_GetCatalog(t *testing.T) {
	t.Parallel()

	broker := &mockBroker{}
	catalog, err := broker.GetCatalog(context.Background())

	require.NoError(t, err)
	assert.NotNil(t, catalog)
}

func TestBrokerInterface_Provision(t *testing.T) {
	t.Parallel()

	broker := &mockBroker{}
	resp, isAsync, err := broker.Provision(
		context.Background(),
		"instance-1",
		osbapi.ProvisionRequest{ServiceID: "svc-1", PlanID: "plan-1"},
		true,
	)

	require.NoError(t, err)
	assert.False(t, isAsync)
	assert.Equal(t, osbapi.ProvisionResponse{}, resp)
}

func TestBrokerInterface_Deprovision(t *testing.T) {
	t.Parallel()

	broker := &mockBroker{}
	resp, isAsync, err := broker.Deprovision(
		context.Background(),
		"instance-1",
		osbapi.DeprovisionRequest{ServiceID: "svc-1", PlanID: "plan-1"},
		false,
	)

	require.NoError(t, err)
	assert.False(t, isAsync)
	assert.Equal(t, osbapi.DeprovisionResponse{}, resp)
}

func TestBrokerInterface_GetInstance(t *testing.T) {
	t.Parallel()

	broker := &mockBroker{}
	resp, err := broker.GetInstance(
		context.Background(),
		"instance-1",
		osbapi.FetchInstanceRequest{},
	)

	require.NoError(t, err)
	assert.Equal(t, osbapi.FetchInstanceResponse{}, resp)
}

func TestBrokerInterface_Update(t *testing.T) {
	t.Parallel()

	broker := &mockBroker{}
	resp, isAsync, err := broker.Update(
		context.Background(),
		"instance-1",
		osbapi.UpdateRequest{ServiceID: "svc-1", PlanID: "plan-2"},
		true,
	)

	require.NoError(t, err)
	assert.False(t, isAsync)
	assert.Equal(t, osbapi.UpdateResponse{}, resp)
}

func TestBrokerInterface_LastOperation(t *testing.T) {
	t.Parallel()

	broker := &mockBroker{}
	resp, err := broker.LastOperation(
		context.Background(),
		"instance-1",
		osbapi.LastOperationRequest{Operation: "op-1"},
	)

	require.NoError(t, err)
	assert.Equal(t, osbapi.LastOperationResponse{}, resp)
}

func TestBrokerInterface_Bind(t *testing.T) {
	t.Parallel()

	broker := &mockBroker{}
	resp, isAsync, err := broker.Bind(
		context.Background(),
		"instance-1",
		"binding-1",
		osbapi.BindRequest{ServiceID: "svc-1", PlanID: "plan-1"},
		false,
	)

	require.NoError(t, err)
	assert.False(t, isAsync)
	assert.Equal(t, osbapi.BindResponse{}, resp)
}

func TestBrokerInterface_Unbind(t *testing.T) {
	t.Parallel()

	broker := &mockBroker{}
	resp, isAsync, err := broker.Unbind(
		context.Background(),
		"instance-1",
		"binding-1",
		osbapi.UnbindRequest{ServiceID: "svc-1", PlanID: "plan-1"},
		true,
	)

	require.NoError(t, err)
	assert.False(t, isAsync)
	assert.Equal(t, osbapi.UnbindResponse{}, resp)
}

func TestBrokerInterface_GetBinding(t *testing.T) {
	t.Parallel()

	broker := &mockBroker{}
	resp, err := broker.GetBinding(
		context.Background(),
		"instance-1",
		"binding-1",
		osbapi.FetchBindingRequest{},
	)

	require.NoError(t, err)
	assert.Equal(t, osbapi.FetchBindingResponse{}, resp)
}

func TestBrokerInterface_LastBindingOperation(t *testing.T) {
	t.Parallel()

	broker := &mockBroker{}
	resp, err := broker.LastBindingOperation(
		context.Background(),
		"instance-1",
		"binding-1",
		osbapi.LastOperationRequest{Operation: "op-1"},
	)

	require.NoError(t, err)
	assert.Equal(t, osbapi.LastOperationResponse{}, resp)
}
