package main

import (
	"context"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fivetwenty-io/osbapi/v2/pkg/broker"
	"github.com/fivetwenty-io/osbapi/v2/pkg/osbapi"
	"github.com/fivetwenty-io/osbapi/v2/pkg/osbclient"
)

// ---------------------------------------------------------------------------
// Test broker implementation
// ---------------------------------------------------------------------------

// testBroker is a minimal in-memory ServiceBroker that tracks provisioned
// instances and bindings so the full platform workflow can run end-to-end.
type testBroker struct {
	mu        sync.Mutex
	instances map[string]instanceRecord
	bindings  map[string]bindingRecord
}

type instanceRecord struct {
	serviceID string
	planID    string
}

type bindingRecord struct {
	serviceID   string
	planID      string
	credentials map[string]any
}

func newTestBroker() *testBroker {
	return &testBroker{
		instances: make(map[string]instanceRecord),
		bindings:  make(map[string]bindingRecord),
	}
}

func (b *testBroker) GetCatalog(_ context.Context) (*osbapi.Catalog, error) {
	return &osbapi.Catalog{
		Services: []osbapi.Service{
			{
				ID:                   "test-svc-id",
				Name:                 "test-service",
				Description:          "A test service for the platform example",
				Bindable:             true,
				InstancesRetrievable: true,
				BindingsRetrievable:  true,
				Plans: []osbapi.Plan{
					{
						ID:          "test-plan-id",
						Name:        "default",
						Description: "Default test plan",
					},
				},
			},
		},
	}, nil
}

func (b *testBroker) Provision(_ context.Context, instanceID string, req osbapi.ProvisionRequest, _ bool) (osbapi.ProvisionResponse, bool, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.instances[instanceID] = instanceRecord{
		serviceID: req.ServiceID,
		planID:    req.PlanID,
	}

	return osbapi.ProvisionResponse{
		DashboardURL: "https://dashboard.example.com/" + instanceID,
	}, false, nil
}

func (b *testBroker) GetInstance(_ context.Context, instanceID string, _ osbapi.FetchInstanceRequest) (osbapi.FetchInstanceResponse, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	inst, ok := b.instances[instanceID]
	if !ok {
		return osbapi.FetchInstanceResponse{}, osbapi.ErrInstanceNotFound
	}

	return osbapi.FetchInstanceResponse{
		ServiceID:    inst.serviceID,
		PlanID:       inst.planID,
		DashboardURL: "https://dashboard.example.com/" + instanceID,
	}, nil
}

func (b *testBroker) Update(_ context.Context, _ string, _ osbapi.UpdateRequest, _ bool) (osbapi.UpdateResponse, bool, error) {
	return osbapi.UpdateResponse{}, false, nil
}

func (b *testBroker) Deprovision(_ context.Context, instanceID string, _ osbapi.DeprovisionRequest, _ bool) (osbapi.DeprovisionResponse, bool, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.instances, instanceID)

	return osbapi.DeprovisionResponse{}, false, nil
}

func (b *testBroker) LastOperation(_ context.Context, _ string, _ osbapi.LastOperationRequest) (osbapi.LastOperationResponse, error) {
	return osbapi.LastOperationResponse{State: osbapi.StateSucceeded}, nil
}

func (b *testBroker) Bind(_ context.Context, instanceID, bindingID string, req osbapi.BindRequest, _ bool) (osbapi.BindResponse, bool, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	creds := map[string]any{
		"username": "test-user",
		"password": "test-pass",
		"host":     "db.example.com",
		"port":     float64(5432),
	}

	key := instanceID + "/" + bindingID
	b.bindings[key] = bindingRecord{
		serviceID:   req.ServiceID,
		planID:      req.PlanID,
		credentials: creds,
	}

	return osbapi.BindResponse{
		Credentials: creds,
	}, false, nil
}

func (b *testBroker) GetBinding(_ context.Context, instanceID, bindingID string, _ osbapi.FetchBindingRequest) (osbapi.FetchBindingResponse, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	key := instanceID + "/" + bindingID
	binding, ok := b.bindings[key]
	if !ok {
		return osbapi.FetchBindingResponse{}, osbapi.ErrBindingNotFound
	}

	return osbapi.FetchBindingResponse{
		Credentials: binding.credentials,
	}, nil
}

func (b *testBroker) Unbind(_ context.Context, instanceID, bindingID string, _ osbapi.UnbindRequest, _ bool) (osbapi.UnbindResponse, bool, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	key := instanceID + "/" + bindingID
	delete(b.bindings, key)

	return osbapi.UnbindResponse{}, false, nil
}

func (b *testBroker) LastBindingOperation(_ context.Context, _, _ string, _ osbapi.LastOperationRequest) (osbapi.LastOperationResponse, error) {
	return osbapi.LastOperationResponse{State: osbapi.StateSucceeded}, nil
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestRunWorkflow(t *testing.T) {
	t.Parallel()

	tb := newTestBroker()
	handler := broker.NewHandler(tb, broker.WithBasicAuth("broker", "secret"))
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	client, err := osbclient.NewWithBasicAuth(srv.URL, "broker", "secret")
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = runWorkflow(ctx, client)
	assert.NoError(t, err)
}

func TestRunWorkflow_VerifiesState(t *testing.T) {
	t.Parallel()

	tb := newTestBroker()
	handler := broker.NewHandler(tb, broker.WithBasicAuth("broker", "secret"))
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	client, err := osbclient.NewWithBasicAuth(srv.URL, "broker", "secret")
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Run the full workflow
	err = runWorkflow(ctx, client)
	require.NoError(t, err)

	// After the workflow, the instance should be deprovisioned
	tb.mu.Lock()
	assert.Empty(t, tb.instances, "instances should be empty after deprovision")
	assert.Empty(t, tb.bindings, "bindings should be empty after unbind")
	tb.mu.Unlock()
}
