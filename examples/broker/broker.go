package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/fivetwenty-io/osbapi/v2/pkg/osbapi"
)

// InMemoryBroker is an example service broker that stores state in memory.
type InMemoryBroker struct {
	mu        sync.RWMutex
	instances map[string]*instanceRecord
	bindings  map[string]*bindingRecord
}

type instanceRecord struct {
	ServiceID    string
	PlanID       string
	Parameters   map[string]any
	DashboardURL string
}

type bindingRecord struct {
	InstanceID  string
	ServiceID   string
	PlanID      string
	Credentials map[string]any
}

// NewInMemoryBroker creates a new in-memory service broker.
func NewInMemoryBroker() *InMemoryBroker {
	return &InMemoryBroker{
		instances: make(map[string]*instanceRecord),
		bindings:  make(map[string]*bindingRecord),
	}
}

// Compile-time check.
var _ osbapi.ServiceBroker = (*InMemoryBroker)(nil)

const (
	exampleServiceID = "a]b1c2d3-e4f5-6789-abcd-ef0123456789"
	freePlanID       = "f1e2d3c4-b5a6-7890-abcd-ef0123456789"
	paidPlanID       = "p1a2i3d4-c5b6-7890-abcd-ef0123456789"
)

// GetCatalog returns the service catalog with one service and two plans.
func (b *InMemoryBroker) GetCatalog(_ context.Context) (*osbapi.Catalog, error) {
	return &osbapi.Catalog{
		Services: []osbapi.Service{
			{
				ID:                   exampleServiceID,
				Name:                 "example-service",
				Description:          "An example service broker for demonstration purposes.",
				Bindable:             true,
				InstancesRetrievable: true,
				BindingsRetrievable:  true,
				PlanUpdateable:       osbapi.BoolPtr(true),
				Tags:                 []string{"example", "demo"},
				Plans: []osbapi.Plan{
					{
						ID:          freePlanID,
						Name:        "free",
						Description: "A free plan with limited resources.",
						Free:        osbapi.BoolPtr(true),
					},
					{
						ID:          paidPlanID,
						Name:        "paid",
						Description: "A paid plan with dedicated resources.",
						Free:        osbapi.BoolPtr(false),
					},
				},
			},
		},
	}, nil
}

// Provision creates a new service instance. If the instance already exists with
// the same plan, the response is idempotent (200 OK). If the plan differs, it
// returns ErrInstanceAlreadyExists.
func (b *InMemoryBroker) Provision(
	_ context.Context,
	instanceID string,
	req osbapi.ProvisionRequest,
	_ bool,
) (osbapi.ProvisionResponse, bool, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if existing, ok := b.instances[instanceID]; ok {
		if existing.PlanID == req.PlanID {
			return osbapi.ProvisionResponse{
				DashboardURL: existing.DashboardURL,
			}, false, nil
		}

		return osbapi.ProvisionResponse{}, false, osbapi.ErrInstanceAlreadyExists
	}

	dashboardURL := fmt.Sprintf("https://example.com/dashboard/%s", instanceID)

	b.instances[instanceID] = &instanceRecord{
		ServiceID:    req.ServiceID,
		PlanID:       req.PlanID,
		Parameters:   req.Parameters,
		DashboardURL: dashboardURL,
	}

	return osbapi.ProvisionResponse{
		DashboardURL: dashboardURL,
	}, false, nil
}

// Deprovision deletes a service instance. Returns ErrGoneError if the instance
// does not exist.
func (b *InMemoryBroker) Deprovision(
	_ context.Context,
	instanceID string,
	_ osbapi.DeprovisionRequest,
	_ bool,
) (osbapi.DeprovisionResponse, bool, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.instances[instanceID]; !ok {
		return osbapi.DeprovisionResponse{}, false, osbapi.ErrGoneError
	}

	delete(b.instances, instanceID)

	return osbapi.DeprovisionResponse{}, false, nil
}

// GetInstance fetches a service instance. Returns ErrInstanceNotFound if the
// instance does not exist.
func (b *InMemoryBroker) GetInstance(
	_ context.Context,
	instanceID string,
	_ osbapi.FetchInstanceRequest,
) (osbapi.FetchInstanceResponse, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	rec, ok := b.instances[instanceID]
	if !ok {
		return osbapi.FetchInstanceResponse{}, osbapi.ErrInstanceNotFound
	}

	return osbapi.FetchInstanceResponse{
		ServiceID:    rec.ServiceID,
		PlanID:       rec.PlanID,
		DashboardURL: rec.DashboardURL,
		Parameters:   rec.Parameters,
	}, nil
}

// Update modifies an existing service instance's plan. Returns
// ErrInstanceNotFound if the instance does not exist.
func (b *InMemoryBroker) Update(
	_ context.Context,
	instanceID string,
	req osbapi.UpdateRequest,
	_ bool,
) (osbapi.UpdateResponse, bool, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	rec, ok := b.instances[instanceID]
	if !ok {
		return osbapi.UpdateResponse{}, false, osbapi.ErrInstanceNotFound
	}

	if req.PlanID != "" {
		rec.PlanID = req.PlanID
	}

	if req.Parameters != nil {
		rec.Parameters = req.Parameters
	}

	return osbapi.UpdateResponse{
		DashboardURL: rec.DashboardURL,
	}, false, nil
}

// LastOperation returns the status of the last operation on a service instance.
// Since all operations are synchronous, this always returns StateSucceeded.
func (b *InMemoryBroker) LastOperation(
	_ context.Context,
	_ string,
	_ osbapi.LastOperationRequest,
) (osbapi.LastOperationResponse, error) {
	return osbapi.LastOperationResponse{
		State:       osbapi.StateSucceeded,
		Description: "operation completed",
	}, nil
}

// Bind creates a service binding with generated credentials. If the binding
// already exists with the same parameters, the response is idempotent.
func (b *InMemoryBroker) Bind(
	_ context.Context,
	instanceID, bindingID string,
	req osbapi.BindRequest,
	_ bool,
) (osbapi.BindResponse, bool, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	key := instanceID + "/" + bindingID

	if existing, ok := b.bindings[key]; ok {
		return osbapi.BindResponse{
			Credentials: existing.Credentials,
		}, false, nil
	}

	credentials := map[string]any{
		"uri":      fmt.Sprintf("postgres://example.com:5432/%s", instanceID),
		"username": fmt.Sprintf("user-%s", bindingID),
		"password": fmt.Sprintf("pass-%s", bindingID),
		"database": fmt.Sprintf("db-%s", instanceID),
	}

	b.bindings[key] = &bindingRecord{
		InstanceID:  instanceID,
		ServiceID:   req.ServiceID,
		PlanID:      req.PlanID,
		Credentials: credentials,
	}

	return osbapi.BindResponse{
		Credentials: credentials,
	}, false, nil
}

// Unbind deletes a service binding. Returns ErrGoneError if the binding does
// not exist.
func (b *InMemoryBroker) Unbind(
	_ context.Context,
	instanceID, bindingID string,
	_ osbapi.UnbindRequest,
	_ bool,
) (osbapi.UnbindResponse, bool, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	key := instanceID + "/" + bindingID

	if _, ok := b.bindings[key]; !ok {
		return osbapi.UnbindResponse{}, false, osbapi.ErrGoneError
	}

	delete(b.bindings, key)

	return osbapi.UnbindResponse{}, false, nil
}

// GetBinding fetches a service binding. Returns ErrBindingNotFound if the
// binding does not exist.
func (b *InMemoryBroker) GetBinding(
	_ context.Context,
	instanceID, bindingID string,
	_ osbapi.FetchBindingRequest,
) (osbapi.FetchBindingResponse, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	key := instanceID + "/" + bindingID

	rec, ok := b.bindings[key]
	if !ok {
		return osbapi.FetchBindingResponse{}, osbapi.ErrBindingNotFound
	}

	return osbapi.FetchBindingResponse{
		Credentials: rec.Credentials,
	}, nil
}

// LastBindingOperation returns the status of the last operation on a service
// binding. Since all operations are synchronous, this always returns
// StateSucceeded.
func (b *InMemoryBroker) LastBindingOperation(
	_ context.Context,
	_, _ string,
	_ osbapi.LastOperationRequest,
) (osbapi.LastOperationResponse, error) {
	return osbapi.LastOperationResponse{
		State:       osbapi.StateSucceeded,
		Description: "operation completed",
	}, nil
}
