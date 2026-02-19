package osbapi_test

import (
	"context"
	"testing"

	"github.com/fivetwenty-io/osbapi/v2/pkg/osbapi"
	"github.com/stretchr/testify/assert"
)

// Compile-time verification that mockClient implements Client.
var _ osbapi.Client = (*mockClient)(nil)

// mockClient is a minimal implementation of the Client interface for testing.
type mockClient struct{}

func (m *mockClient) GetCatalog(_ context.Context) (*osbapi.Catalog, error) {
	return &osbapi.Catalog{}, nil
}

func (m *mockClient) Provision(_ context.Context, _ string, _ osbapi.ProvisionRequest, _ bool) (osbapi.ProvisionResponse, bool, error) {
	return osbapi.ProvisionResponse{}, false, nil
}

func (m *mockClient) Deprovision(_ context.Context, _ string, _ osbapi.DeprovisionRequest, _ bool) (osbapi.DeprovisionResponse, bool, error) {
	return osbapi.DeprovisionResponse{}, false, nil
}

func (m *mockClient) GetInstance(_ context.Context, _ string, _ osbapi.FetchInstanceRequest) (osbapi.FetchInstanceResponse, error) {
	return osbapi.FetchInstanceResponse{}, nil
}

func (m *mockClient) Update(_ context.Context, _ string, _ osbapi.UpdateRequest, _ bool) (osbapi.UpdateResponse, bool, error) {
	return osbapi.UpdateResponse{}, false, nil
}

func (m *mockClient) PollLastOperation(_ context.Context, _ string, _ osbapi.LastOperationRequest) (osbapi.LastOperationResponse, error) {
	return osbapi.LastOperationResponse{}, nil
}

func (m *mockClient) Bind(_ context.Context, _, _ string, _ osbapi.BindRequest, _ bool) (osbapi.BindResponse, bool, error) {
	return osbapi.BindResponse{}, false, nil
}

func (m *mockClient) Unbind(_ context.Context, _, _ string, _ osbapi.UnbindRequest, _ bool) (osbapi.UnbindResponse, bool, error) {
	return osbapi.UnbindResponse{}, false, nil
}

func (m *mockClient) GetBinding(_ context.Context, _, _ string, _ osbapi.FetchBindingRequest) (osbapi.FetchBindingResponse, error) {
	return osbapi.FetchBindingResponse{}, nil
}

func (m *mockClient) PollBindingLastOperation(_ context.Context, _, _ string, _ osbapi.LastOperationRequest) (osbapi.LastOperationResponse, error) {
	return osbapi.LastOperationResponse{}, nil
}

func (m *mockClient) PollInstanceUntilComplete(_ context.Context, _ string, _ osbapi.PollConfig) (osbapi.LastOperationResponse, error) {
	return osbapi.LastOperationResponse{}, nil
}

func (m *mockClient) PollBindingUntilComplete(_ context.Context, _, _ string, _ osbapi.PollConfig) (osbapi.LastOperationResponse, error) {
	return osbapi.LastOperationResponse{}, nil
}

func TestClientInterfaceCompilation(t *testing.T) {
	t.Parallel()

	mock := &mockClient{}
	var client osbapi.Client = mock
	assert.NotNil(t, client, "mockClient must satisfy Client interface")
}

func TestClientGetCatalog(t *testing.T) {
	t.Parallel()

	mock := &mockClient{}
	catalog, err := mock.GetCatalog(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, catalog)
}

func TestClientProvision(t *testing.T) {
	t.Parallel()

	mock := &mockClient{}
	resp, isAsync, err := mock.Provision(context.Background(), "instance-1", osbapi.ProvisionRequest{
		ServiceID: "svc-1",
		PlanID:    "plan-1",
	}, true)

	assert.NoError(t, err)
	assert.False(t, isAsync)
	assert.Equal(t, osbapi.ProvisionResponse{}, resp)
}

func TestClientDeprovision(t *testing.T) {
	t.Parallel()

	mock := &mockClient{}
	resp, isAsync, err := mock.Deprovision(context.Background(), "instance-1", osbapi.DeprovisionRequest{
		ServiceID: "svc-1",
		PlanID:    "plan-1",
	}, true)

	assert.NoError(t, err)
	assert.False(t, isAsync)
	assert.Equal(t, osbapi.DeprovisionResponse{}, resp)
}

func TestClientGetInstance(t *testing.T) {
	t.Parallel()

	mock := &mockClient{}
	resp, err := mock.GetInstance(context.Background(), "instance-1", osbapi.FetchInstanceRequest{})

	assert.NoError(t, err)
	assert.Equal(t, osbapi.FetchInstanceResponse{}, resp)
}

func TestClientUpdate(t *testing.T) {
	t.Parallel()

	mock := &mockClient{}
	resp, isAsync, err := mock.Update(context.Background(), "instance-1", osbapi.UpdateRequest{
		ServiceID: "svc-1",
	}, true)

	assert.NoError(t, err)
	assert.False(t, isAsync)
	assert.Equal(t, osbapi.UpdateResponse{}, resp)
}

func TestClientPollLastOperation(t *testing.T) {
	t.Parallel()

	mock := &mockClient{}
	resp, err := mock.PollLastOperation(context.Background(), "instance-1", osbapi.LastOperationRequest{
		Operation: "provision",
	})

	assert.NoError(t, err)
	assert.Equal(t, osbapi.LastOperationResponse{}, resp)
}

func TestClientBind(t *testing.T) {
	t.Parallel()

	mock := &mockClient{}
	resp, isAsync, err := mock.Bind(context.Background(), "instance-1", "binding-1", osbapi.BindRequest{
		ServiceID: "svc-1",
		PlanID:    "plan-1",
	}, true)

	assert.NoError(t, err)
	assert.False(t, isAsync)
	assert.Equal(t, osbapi.BindResponse{}, resp)
}

func TestClientUnbind(t *testing.T) {
	t.Parallel()

	mock := &mockClient{}
	resp, isAsync, err := mock.Unbind(context.Background(), "instance-1", "binding-1", osbapi.UnbindRequest{
		ServiceID: "svc-1",
		PlanID:    "plan-1",
	}, true)

	assert.NoError(t, err)
	assert.False(t, isAsync)
	assert.Equal(t, osbapi.UnbindResponse{}, resp)
}

func TestClientGetBinding(t *testing.T) {
	t.Parallel()

	mock := &mockClient{}
	resp, err := mock.GetBinding(context.Background(), "instance-1", "binding-1", osbapi.FetchBindingRequest{})

	assert.NoError(t, err)
	assert.Equal(t, osbapi.FetchBindingResponse{}, resp)
}

func TestClientPollBindingLastOperation(t *testing.T) {
	t.Parallel()

	mock := &mockClient{}
	resp, err := mock.PollBindingLastOperation(context.Background(), "instance-1", "binding-1", osbapi.LastOperationRequest{
		Operation: "bind",
	})

	assert.NoError(t, err)
	assert.Equal(t, osbapi.LastOperationResponse{}, resp)
}

func TestClientPollInstanceUntilComplete(t *testing.T) {
	t.Parallel()

	mock := &mockClient{}
	resp, err := mock.PollInstanceUntilComplete(context.Background(), "instance-1", osbapi.PollConfig{})

	assert.NoError(t, err)
	assert.Equal(t, osbapi.LastOperationResponse{}, resp)
}

func TestClientPollBindingUntilComplete(t *testing.T) {
	t.Parallel()

	mock := &mockClient{}
	resp, err := mock.PollBindingUntilComplete(context.Background(), "instance-1", "binding-1", osbapi.PollConfig{})

	assert.NoError(t, err)
	assert.Equal(t, osbapi.LastOperationResponse{}, resp)
}
