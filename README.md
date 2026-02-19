# osbapi

Go client and server library for the [Open Service Broker API v2.17](https://github.com/openservicebrokerapi/servicebroker/blob/v2.17/spec.md).

## Features

- Full coverage of all 10 OSB API v2.17 endpoints

- Both client (platform) and server (broker) implementations

- Synchronous and asynchronous operation support with polling helpers

- HTTP Basic Authentication for client and server

- API version header validation with configurable minimum version

- Zero external dependencies for routing (uses Go 1.22+ `net/http` ServeMux)

- Structured error handling with `OSBError` types and sentinel errors

- Originating identity header encoding/decoding

## Installation

```bash
go get github.com/fivetwenty-io/osbapi/v2
```

Requires Go 1.22 or later.

## Quick Start: Broker

Implement the `ServiceBroker` interface and serve it with `broker.NewHandler`.

```go
package main

import (
    "context"
    "log"
    "net/http"

    "github.com/fivetwenty-io/osbapi/v2/pkg/broker"
    "github.com/fivetwenty-io/osbapi/v2/pkg/osbapi"
)

type MyBroker struct{}

func (b *MyBroker) GetCatalog(_ context.Context) (*osbapi.Catalog, error) {
    return &osbapi.Catalog{
        Services: []osbapi.Service{
            {
                ID:          "my-service-id",
                Name:        "my-service",
                Description: "An example service.",
                Bindable:    true,
                Plans: []osbapi.Plan{
                    {
                        ID:          "my-plan-id",
                        Name:        "default",
                        Description: "The default plan.",
                        Free:        osbapi.BoolPtr(true),
                    },
                },
            },
        },
    }, nil
}

func (b *MyBroker) Provision(
    _ context.Context,
    instanceID string,
    req osbapi.ProvisionRequest,
    async bool,
) (osbapi.ProvisionResponse, bool, error) {
    // Create the service instance using req.ServiceID, req.PlanID, etc.
    return osbapi.ProvisionResponse{
        DashboardURL: "https://example.com/dashboard/" + instanceID,
    }, false, nil
}

// ... implement remaining ServiceBroker methods ...

func main() {
    handler := broker.NewHandler(&MyBroker{},
        broker.WithBasicAuth("broker-user", "broker-pass"),
    )
    log.Fatal(http.ListenAndServe(":8080", handler))
}
```

## Quick Start: Client

Use `osbclient.NewWithBasicAuth` to create a client, then call methods for each OSB API operation.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/fivetwenty-io/osbapi/v2/pkg/osbapi"
    "github.com/fivetwenty-io/osbapi/v2/pkg/osbclient"
)

func main() {
    client, err := osbclient.NewWithBasicAuth(
        "http://localhost:8080", "broker-user", "broker-pass",
    )
    if err != nil {
        log.Fatal(err)
    }

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Fetch the catalog
    catalog, err := client.GetCatalog(ctx)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Found %d services\n", len(catalog.Services))

    svc := catalog.Services[0]
    plan := svc.Plans[0]

    // Provision a service instance
    provResp, isAsync, err := client.Provision(ctx, "my-instance-1", osbapi.ProvisionRequest{
        ServiceID:        svc.ID,
        PlanID:           plan.ID,
        OrganizationGUID: "my-org",
        SpaceGUID:        "my-space",
    }, false)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Provisioned (async=%v, dashboard=%s)\n", isAsync, provResp.DashboardURL)

    // Create a service binding
    bindResp, _, err := client.Bind(ctx, "my-instance-1", "my-binding-1", osbapi.BindRequest{
        ServiceID: svc.ID,
        PlanID:    plan.ID,
    }, false)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Credentials: %v\n", bindResp.Credentials)
}
```

## Async Operations

Brokers that perform long-running work return `isAsync=true`, which maps to HTTP 202 Accepted. The platform must then poll the last operation endpoint until the operation completes.

### Provisioning with async

```go
resp, isAsync, err := client.Provision(ctx, "my-instance-1", osbapi.ProvisionRequest{
    ServiceID:        serviceID,
    PlanID:           planID,
    OrganizationGUID: "my-org",
    SpaceGUID:        "my-space",
}, true) // true = accepts_incomplete
if err != nil {
    log.Fatal(err)
}

if isAsync {
    // Poll until the operation completes, fails, or times out
    result, err := client.PollInstanceUntilComplete(ctx, "my-instance-1", osbapi.PollConfig{
        Interval: 5 * time.Second,
        Timeout:  10 * time.Minute,
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Final state: %s\n", result.State)
}
```

### Manual polling

```go
for {
    result, err := client.PollLastOperation(ctx, "my-instance-1", osbapi.LastOperationRequest{
        Operation: resp.Operation,
    })
    if err != nil {
        log.Fatal(err)
    }

    switch result.State {
    case osbapi.StateSucceeded:
        fmt.Println("Done")
        return
    case osbapi.StateFailed:
        log.Fatalf("Failed: %s", result.Description)
    case osbapi.StateInProgress:
        fmt.Println("Still working...")
        time.Sleep(5 * time.Second)
    }
}
```

## Package Overview

| Package | Import Path | Purpose |
|---------|-------------|---------|
| osbapi | `github.com/fivetwenty-io/osbapi/v2/pkg/osbapi` | Shared types, interfaces, and errors |
| broker | `github.com/fivetwenty-io/osbapi/v2/pkg/broker` | Server-side handler (broker authors) |
| osbclient | `github.com/fivetwenty-io/osbapi/v2/pkg/osbclient` | Client-side HTTP client (platforms) |

## API Coverage

All 10 endpoints defined by the OSB API v2.17 specification are implemented.

| # | Endpoint | Method | Path | Client Method | Broker Method |
|---|----------|--------|------|---------------|---------------|
| 1 | Catalog | GET | `/v2/catalog` | `GetCatalog` | `GetCatalog` |
| 2 | Provision | PUT | `/v2/service_instances/:id` | `Provision` | `Provision` |
| 3 | Fetch Instance | GET | `/v2/service_instances/:id` | `GetInstance` | `GetInstance` |
| 4 | Update | PATCH | `/v2/service_instances/:id` | `Update` | `Update` |
| 5 | Deprovision | DELETE | `/v2/service_instances/:id` | `Deprovision` | `Deprovision` |
| 6 | Last Operation (Instance) | GET | `/v2/service_instances/:id/last_operation` | `PollLastOperation` | `LastOperation` |
| 7 | Bind | PUT | `/v2/service_instances/:id/service_bindings/:bid` | `Bind` | `Bind` |
| 8 | Fetch Binding | GET | `/v2/service_instances/:id/service_bindings/:bid` | `GetBinding` | `GetBinding` |
| 9 | Unbind | DELETE | `/v2/service_instances/:id/service_bindings/:bid` | `Unbind` | `Unbind` |
| 10 | Last Operation (Binding) | GET | `/v2/service_instances/:id/service_bindings/:bid/last_operation` | `PollBindingLastOperation` | `LastBindingOperation` |

The client also provides two polling helpers that are not part of the OSB API itself:

- `PollInstanceUntilComplete` -- polls instance last operation until success, failure, or timeout

- `PollBindingUntilComplete` -- polls binding last operation until success, failure, or timeout

## Error Handling

### OSBError

All OSB API error responses are returned as `*osbapi.OSBError`, which implements the `error` interface.

```go
var osbErr *osbapi.OSBError
if errors.As(err, &osbErr) {
    fmt.Printf("Error code: %s\n", osbErr.ErrorCode)
    fmt.Printf("Description: %s\n", osbErr.Description)
    fmt.Printf("HTTP status: %d\n", osbErr.StatusCode)
}
```

### Sentinel Errors

The `osbapi` package defines sentinel errors for common failure modes. Broker implementations return these to produce the correct HTTP status codes automatically.

| Sentinel Error | HTTP Status | Use Case |
|----------------|-------------|----------|
| `ErrAsyncRequired` | 422 | Broker requires `accepts_incomplete=true` |
| `ErrConcurrencyError` | 422 | Concurrent operation already in progress |
| `ErrInstanceNotFound` | 404 | Service instance does not exist |
| `ErrBindingNotFound` | 404 | Service binding does not exist |
| `ErrGoneError` | 410 | Resource has been deleted |
| `ErrBadRequest` | 400 | Malformed or invalid request |
| `ErrInstanceAlreadyExists` | 409 | Instance already exists with different config |
| `ErrBindingAlreadyExists` | 409 | Binding already exists with different config |
| `ErrMaintenanceInfoConflict` | 422 | Maintenance info version mismatch |
| `ErrRequiresApp` | 422 | Binding requires `app_guid` |
| `ErrUnauthorized` | 401 | Authentication failure |
| `ErrPlanQuotaExceeded` | 422 | Plan quota has been exceeded |
| `ErrInvalidParameters` | 400 | Parameters failed validation |

### Classification Helpers

Convenience functions for classifying errors without inspecting the type directly:

```go
if osbapi.IsAsyncRequired(err) {
    // Retry with accepts_incomplete=true
}

if osbapi.IsConcurrencyError(err) {
    // Wait and retry
}

if osbapi.IsNotFound(err) {
    // Resource does not exist (404 or 410)
}

if osbapi.IsGone(err) {
    // Resource has been permanently deleted (410)
}
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, testing, and pull request guidelines.

## License

Apache 2.0
