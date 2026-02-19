package server_test

import (
	"context"

	osbapi "github.com/fivetwenty-io/osbapi/v2/pkg/osbapi"
)

// mockBroker implements osbapi.ServiceBroker for testing. Each method stores
// the arguments it was called with and returns pre-configured responses so
// that tests can verify both request parsing and response rendering.
type mockBroker struct {
	// Catalog
	catalog    *osbapi.Catalog
	catalogErr error

	// Provision
	provisionResp    osbapi.ProvisionResponse
	provisionAsync   bool
	provisionErr     error
	lastInstanceID   string
	lastProvisionReq osbapi.ProvisionRequest

	// Deprovision
	deprovisionResp    osbapi.DeprovisionResponse
	deprovisionAsync   bool
	deprovisionErr     error
	lastDeprovisionReq osbapi.DeprovisionRequest

	// FetchInstance
	fetchInstanceResp osbapi.FetchInstanceResponse
	fetchInstanceErr  error

	// Update
	updateResp    osbapi.UpdateResponse
	updateAsync   bool
	updateErr     error
	lastUpdateReq osbapi.UpdateRequest

	// LastOperation
	lastOpResp    osbapi.LastOperationResponse
	lastOpErr     error
	lastLastOpReq osbapi.LastOperationRequest

	// Bind
	bindResp      osbapi.BindResponse
	bindAsync     bool
	bindErr       error
	lastBindingID string
	lastBindReq   osbapi.BindRequest

	// Unbind
	unbindResp    osbapi.UnbindResponse
	unbindAsync   bool
	unbindErr     error
	lastUnbindReq osbapi.UnbindRequest

	// GetBinding
	getBindingResp   osbapi.FetchBindingResponse
	getBindingErr    error
	lastFetchBindReq osbapi.FetchBindingRequest

	// LastBindingOperation
	lastBindingOpResp    osbapi.LastOperationResponse
	lastBindingOpErr     error
	lastLastBindingOpReq osbapi.LastOperationRequest
}

// ---------------------------------------------------------------------------
// ServiceBroker implementation
// ---------------------------------------------------------------------------

func (m *mockBroker) GetCatalog(_ context.Context) (*osbapi.Catalog, error) {
	return m.catalog, m.catalogErr
}

func (m *mockBroker) Provision(_ context.Context, instanceID string, req osbapi.ProvisionRequest, _ bool) (osbapi.ProvisionResponse, bool, error) {
	m.lastInstanceID = instanceID
	m.lastProvisionReq = req
	return m.provisionResp, m.provisionAsync, m.provisionErr
}

func (m *mockBroker) Deprovision(_ context.Context, instanceID string, req osbapi.DeprovisionRequest, _ bool) (osbapi.DeprovisionResponse, bool, error) {
	m.lastInstanceID = instanceID
	m.lastDeprovisionReq = req
	return m.deprovisionResp, m.deprovisionAsync, m.deprovisionErr
}

func (m *mockBroker) GetInstance(_ context.Context, instanceID string, _ osbapi.FetchInstanceRequest) (osbapi.FetchInstanceResponse, error) {
	m.lastInstanceID = instanceID
	return m.fetchInstanceResp, m.fetchInstanceErr
}

func (m *mockBroker) Update(_ context.Context, instanceID string, req osbapi.UpdateRequest, _ bool) (osbapi.UpdateResponse, bool, error) {
	m.lastInstanceID = instanceID
	m.lastUpdateReq = req
	return m.updateResp, m.updateAsync, m.updateErr
}

func (m *mockBroker) LastOperation(_ context.Context, instanceID string, req osbapi.LastOperationRequest) (osbapi.LastOperationResponse, error) {
	m.lastInstanceID = instanceID
	m.lastLastOpReq = req
	return m.lastOpResp, m.lastOpErr
}

func (m *mockBroker) Bind(_ context.Context, instanceID, bindingID string, req osbapi.BindRequest, _ bool) (osbapi.BindResponse, bool, error) {
	m.lastInstanceID = instanceID
	m.lastBindingID = bindingID
	m.lastBindReq = req
	return m.bindResp, m.bindAsync, m.bindErr
}

func (m *mockBroker) Unbind(_ context.Context, instanceID, bindingID string, req osbapi.UnbindRequest, _ bool) (osbapi.UnbindResponse, bool, error) {
	m.lastInstanceID = instanceID
	m.lastBindingID = bindingID
	m.lastUnbindReq = req
	return m.unbindResp, m.unbindAsync, m.unbindErr
}

func (m *mockBroker) GetBinding(_ context.Context, instanceID, bindingID string, req osbapi.FetchBindingRequest) (osbapi.FetchBindingResponse, error) {
	m.lastInstanceID = instanceID
	m.lastBindingID = bindingID
	m.lastFetchBindReq = req
	return m.getBindingResp, m.getBindingErr
}

func (m *mockBroker) LastBindingOperation(_ context.Context, instanceID, bindingID string, req osbapi.LastOperationRequest) (osbapi.LastOperationResponse, error) {
	m.lastInstanceID = instanceID
	m.lastBindingID = bindingID
	m.lastLastBindingOpReq = req
	return m.lastBindingOpResp, m.lastBindingOpErr
}
