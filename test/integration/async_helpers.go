//go:build integration

package integration

import (
	"context"
	"fmt"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fivetwenty-io/osbapi/v2/pkg/broker"
	"github.com/fivetwenty-io/osbapi/v2/pkg/osbapi"
	"github.com/fivetwenty-io/osbapi/v2/pkg/osbclient"
)

// asyncBroker wraps a sync broker and makes operations async.
// When an async operation is requested:
// 1. Store the operation as "in progress"
// 2. Start a goroutine that marks it "succeeded" after a short delay
// 3. Return isAsync=true with an operation token
type asyncBroker struct {
	mu              sync.Mutex
	instances       map[string]*asyncInstanceRecord
	bindings        map[string]*asyncBindingRecord
	operations      map[string]*asyncOperation
	lastInstanceOp  map[string]string // instanceID -> last operation token
	lastBindingOp   map[string]string // instanceID/bindingID -> last operation token
	delay           time.Duration     // How long before operations complete
	opCounter       atomic.Int64
}

type asyncInstanceRecord struct {
	ServiceID  string
	PlanID     string
	Parameters map[string]any
}

type asyncBindingRecord struct {
	InstanceID  string
	ServiceID   string
	PlanID      string
	Credentials map[string]any
}

type asyncOperation struct {
	State       osbapi.OperationState
	Description string
}

var _ osbapi.ServiceBroker = (*asyncBroker)(nil)

func newAsyncBroker(delay time.Duration) *asyncBroker {
	return &asyncBroker{
		instances:      make(map[string]*asyncInstanceRecord),
		bindings:       make(map[string]*asyncBindingRecord),
		operations:     make(map[string]*asyncOperation),
		lastInstanceOp: make(map[string]string),
		lastBindingOp:  make(map[string]string),
		delay:          delay,
	}
}

func (b *asyncBroker) nextOpToken(id string) string {
	n := b.opCounter.Add(1)
	return fmt.Sprintf("op-%s-%d", id, n)
}

// GetCatalog returns the service catalog.
func (b *asyncBroker) GetCatalog(_ context.Context) (*osbapi.Catalog, error) {
	return &osbapi.Catalog{
		Services: []osbapi.Service{
			{
				ID:                   "async-service-1",
				Name:                 "async-test-service",
				Description:          "An async test service",
				Bindable:             true,
				InstancesRetrievable: true,
				BindingsRetrievable:  true,
				PlanUpdateable:       osbapi.BoolPtr(true),
				Plans: []osbapi.Plan{
					{
						ID:          "async-plan-1",
						Name:        "async-small",
						Description: "A small async plan",
						Free:        osbapi.BoolPtr(true),
					},
					{
						ID:          "async-plan-2",
						Name:        "async-large",
						Description: "A large async plan",
						Free:        osbapi.BoolPtr(false),
					},
				},
			},
		},
	}, nil
}

// Provision creates a new service instance asynchronously.
func (b *asyncBroker) Provision(
	_ context.Context,
	instanceID string,
	req osbapi.ProvisionRequest,
	async bool,
) (osbapi.ProvisionResponse, bool, error) {
	if !async {
		return osbapi.ProvisionResponse{}, false, osbapi.ErrAsyncRequired
	}

	opToken := b.nextOpToken(instanceID)

	b.mu.Lock()
	b.operations[opToken] = &asyncOperation{
		State:       osbapi.StateInProgress,
		Description: "provisioning",
	}
	b.lastInstanceOp[instanceID] = opToken
	b.mu.Unlock()

	go func() {
		time.Sleep(b.delay)

		b.mu.Lock()
		defer b.mu.Unlock()

		b.instances[instanceID] = &asyncInstanceRecord{
			ServiceID:  req.ServiceID,
			PlanID:     req.PlanID,
			Parameters: req.Parameters,
		}
		b.operations[opToken] = &asyncOperation{
			State:       osbapi.StateSucceeded,
			Description: "provisioned",
		}
	}()

	return osbapi.ProvisionResponse{
		Operation: opToken,
	}, true, nil
}

// Deprovision deletes a service instance asynchronously.
func (b *asyncBroker) Deprovision(
	_ context.Context,
	instanceID string,
	_ osbapi.DeprovisionRequest,
	async bool,
) (osbapi.DeprovisionResponse, bool, error) {
	if !async {
		return osbapi.DeprovisionResponse{}, false, osbapi.ErrAsyncRequired
	}

	opToken := b.nextOpToken(instanceID)

	b.mu.Lock()
	b.operations[opToken] = &asyncOperation{
		State:       osbapi.StateInProgress,
		Description: "deprovisioning",
	}
	b.lastInstanceOp[instanceID] = opToken
	b.mu.Unlock()

	go func() {
		time.Sleep(b.delay)

		b.mu.Lock()
		defer b.mu.Unlock()

		delete(b.instances, instanceID)
		b.operations[opToken] = &asyncOperation{
			State:       osbapi.StateSucceeded,
			Description: "deprovisioned",
		}
	}()

	return osbapi.DeprovisionResponse{
		Operation: opToken,
	}, true, nil
}

// GetInstance fetches a service instance.
func (b *asyncBroker) GetInstance(
	_ context.Context,
	instanceID string,
	_ osbapi.FetchInstanceRequest,
) (osbapi.FetchInstanceResponse, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	inst, ok := b.instances[instanceID]
	if !ok {
		return osbapi.FetchInstanceResponse{}, osbapi.ErrInstanceNotFound
	}

	return osbapi.FetchInstanceResponse{
		ServiceID:  inst.ServiceID,
		PlanID:     inst.PlanID,
		Parameters: inst.Parameters,
	}, nil
}

// Update modifies an existing service instance asynchronously.
func (b *asyncBroker) Update(
	_ context.Context,
	instanceID string,
	req osbapi.UpdateRequest,
	async bool,
) (osbapi.UpdateResponse, bool, error) {
	if !async {
		return osbapi.UpdateResponse{}, false, osbapi.ErrAsyncRequired
	}

	opToken := b.nextOpToken(instanceID)

	b.mu.Lock()
	b.operations[opToken] = &asyncOperation{
		State:       osbapi.StateInProgress,
		Description: "updating",
	}
	b.lastInstanceOp[instanceID] = opToken
	b.mu.Unlock()

	go func() {
		time.Sleep(b.delay)

		b.mu.Lock()
		defer b.mu.Unlock()

		if inst, ok := b.instances[instanceID]; ok {
			if req.PlanID != "" {
				inst.PlanID = req.PlanID
			}
			if req.Parameters != nil {
				inst.Parameters = req.Parameters
			}
		}
		b.operations[opToken] = &asyncOperation{
			State:       osbapi.StateSucceeded,
			Description: "updated",
		}
	}()

	return osbapi.UpdateResponse{
		Operation: opToken,
	}, true, nil
}

// LastOperation polls the status of an async instance operation.
// If no operation token is given, falls back to the most recent operation for
// the instance.
func (b *asyncBroker) LastOperation(
	_ context.Context,
	instanceID string,
	req osbapi.LastOperationRequest,
) (osbapi.LastOperationResponse, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	opKey := req.Operation
	if opKey == "" {
		var ok bool
		opKey, ok = b.lastInstanceOp[instanceID]
		if !ok {
			return osbapi.LastOperationResponse{}, osbapi.ErrInstanceNotFound
		}
	}

	op, ok := b.operations[opKey]
	if !ok {
		return osbapi.LastOperationResponse{}, osbapi.ErrInstanceNotFound
	}

	return osbapi.LastOperationResponse{
		State:       op.State,
		Description: op.Description,
	}, nil
}

// Bind creates a new service binding asynchronously.
func (b *asyncBroker) Bind(
	_ context.Context,
	instanceID, bindingID string,
	req osbapi.BindRequest,
	async bool,
) (osbapi.BindResponse, bool, error) {
	if !async {
		return osbapi.BindResponse{}, false, osbapi.ErrAsyncRequired
	}

	opToken := b.nextOpToken(bindingID)

	bindKey := instanceID + "/" + bindingID

	b.mu.Lock()
	b.operations[opToken] = &asyncOperation{
		State:       osbapi.StateInProgress,
		Description: "binding",
	}
	b.lastBindingOp[bindKey] = opToken
	b.mu.Unlock()

	go func() {
		time.Sleep(b.delay)

		b.mu.Lock()
		defer b.mu.Unlock()

		b.bindings[bindKey] = &asyncBindingRecord{
			InstanceID: instanceID,
			ServiceID:  req.ServiceID,
			PlanID:     req.PlanID,
			Credentials: map[string]any{
				"username": "async-user",
				"password": "async-pass",
			},
		}
		b.operations[opToken] = &asyncOperation{
			State:       osbapi.StateSucceeded,
			Description: "bound",
		}
	}()

	return osbapi.BindResponse{
		Operation: opToken,
	}, true, nil
}

// Unbind deletes a service binding asynchronously.
func (b *asyncBroker) Unbind(
	_ context.Context,
	instanceID, bindingID string,
	_ osbapi.UnbindRequest,
	async bool,
) (osbapi.UnbindResponse, bool, error) {
	if !async {
		return osbapi.UnbindResponse{}, false, osbapi.ErrAsyncRequired
	}

	opToken := b.nextOpToken(bindingID)

	unbindKey := instanceID + "/" + bindingID

	b.mu.Lock()
	b.operations[opToken] = &asyncOperation{
		State:       osbapi.StateInProgress,
		Description: "unbinding",
	}
	b.lastBindingOp[unbindKey] = opToken
	b.mu.Unlock()

	go func() {
		time.Sleep(b.delay)

		b.mu.Lock()
		defer b.mu.Unlock()

		delete(b.bindings, unbindKey)
		b.operations[opToken] = &asyncOperation{
			State:       osbapi.StateSucceeded,
			Description: "unbound",
		}
	}()

	return osbapi.UnbindResponse{
		Operation: opToken,
	}, true, nil
}

// GetBinding fetches a service binding.
func (b *asyncBroker) GetBinding(
	_ context.Context,
	instanceID, bindingID string,
	_ osbapi.FetchBindingRequest,
) (osbapi.FetchBindingResponse, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	key := instanceID + "/" + bindingID
	bind, ok := b.bindings[key]
	if !ok {
		return osbapi.FetchBindingResponse{}, osbapi.ErrBindingNotFound
	}

	return osbapi.FetchBindingResponse{
		Credentials: bind.Credentials,
	}, nil
}

// LastBindingOperation polls the status of an async binding operation.
// If no operation token is given, falls back to the most recent operation for
// the binding.
func (b *asyncBroker) LastBindingOperation(
	_ context.Context,
	instanceID, bindingID string,
	req osbapi.LastOperationRequest,
) (osbapi.LastOperationResponse, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	opKey := req.Operation
	if opKey == "" {
		key := instanceID + "/" + bindingID
		var ok bool
		opKey, ok = b.lastBindingOp[key]
		if !ok {
			return osbapi.LastOperationResponse{}, osbapi.ErrBindingNotFound
		}
	}

	op, ok := b.operations[opKey]
	if !ok {
		return osbapi.LastOperationResponse{}, osbapi.ErrBindingNotFound
	}

	return osbapi.LastOperationResponse{
		State:       op.State,
		Description: op.Description,
	}, nil
}

// StartAsyncTestBroker starts an async test broker with the given operation delay.
func StartAsyncTestBroker(t *testing.T, delay time.Duration) (*httptest.Server, osbapi.Client) {
	t.Helper()

	b := newAsyncBroker(delay)
	handler := broker.NewHandler(b,
		broker.WithBasicAuth(testUsername, testPassword),
	)

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	client, err := osbclient.NewWithBasicAuth(srv.URL, testUsername, testPassword)
	if err != nil {
		t.Fatalf("creating test client: %v", err)
	}

	return srv, client
}
