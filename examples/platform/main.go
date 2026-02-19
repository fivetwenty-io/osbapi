package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fivetwenty-io/osbapi/v2/pkg/osbapi"
	"github.com/fivetwenty-io/osbapi/v2/pkg/osbclient"
)

func main() {
	brokerURL := os.Getenv("BROKER_URL")
	if brokerURL == "" {
		brokerURL = "http://localhost:8080"
	}

	username := os.Getenv("BROKER_USERNAME")
	if username == "" {
		username = "broker"
	}

	password := os.Getenv("BROKER_PASSWORD")
	if password == "" {
		password = "secret"
	}

	client, err := osbclient.NewWithBasicAuth(brokerURL, username, password)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := runWorkflow(ctx, client); err != nil {
		log.Fatalf("Workflow failed: %v", err)
	}
}

func runWorkflow(ctx context.Context, client osbapi.Client) error {
	// 1. Fetch catalog
	fmt.Println("Fetching catalog...")
	catalog, err := client.GetCatalog(ctx)
	if err != nil {
		return fmt.Errorf("fetching catalog: %w", err)
	}
	fmt.Printf("Found %d services\n", len(catalog.Services))

	if len(catalog.Services) == 0 {
		return fmt.Errorf("no services in catalog")
	}

	svc := catalog.Services[0]
	fmt.Printf("Service: %s (%s)\n", svc.Name, svc.ID)

	if len(svc.Plans) == 0 {
		return fmt.Errorf("no plans for service %s", svc.Name)
	}

	plan := svc.Plans[0]
	fmt.Printf("Plan: %s (%s)\n", plan.Name, plan.ID)

	// 2. Provision an instance
	instanceID := "example-instance-001"
	fmt.Printf("\nProvisioning instance %s...\n", instanceID)
	provResp, isAsync, err := client.Provision(ctx, instanceID, osbapi.ProvisionRequest{
		ServiceID:        svc.ID,
		PlanID:           plan.ID,
		OrganizationGUID: "example-org",
		SpaceGUID:        "example-space",
	}, false)
	if err != nil {
		return fmt.Errorf("provisioning: %w", err)
	}
	fmt.Printf("Provisioned (async=%v, dashboard=%s)\n", isAsync, provResp.DashboardURL)

	// 3. Fetch the instance
	fmt.Printf("\nFetching instance %s...\n", instanceID)
	instResp, err := client.GetInstance(ctx, instanceID, osbapi.FetchInstanceRequest{})
	if err != nil {
		return fmt.Errorf("fetching instance: %w", err)
	}
	fmt.Printf("Instance service=%s plan=%s\n", instResp.ServiceID, instResp.PlanID)

	// 4. Create a binding
	bindingID := "example-binding-001"
	fmt.Printf("\nBinding %s to instance %s...\n", bindingID, instanceID)
	bindResp, isAsync, err := client.Bind(ctx, instanceID, bindingID, osbapi.BindRequest{
		ServiceID: svc.ID,
		PlanID:    plan.ID,
	}, false)
	if err != nil {
		return fmt.Errorf("binding: %w", err)
	}
	fmt.Printf("Bound (async=%v, credentials=%v)\n", isAsync, bindResp.Credentials)

	// 5. Fetch the binding
	fmt.Printf("\nFetching binding %s...\n", bindingID)
	fetchBindResp, err := client.GetBinding(ctx, instanceID, bindingID, osbapi.FetchBindingRequest{})
	if err != nil {
		return fmt.Errorf("fetching binding: %w", err)
	}
	fmt.Printf("Binding credentials: %v\n", fetchBindResp.Credentials)

	// 6. Unbind
	fmt.Printf("\nUnbinding %s...\n", bindingID)
	_, _, err = client.Unbind(ctx, instanceID, bindingID, osbapi.UnbindRequest{
		ServiceID: svc.ID,
		PlanID:    plan.ID,
	}, false)
	if err != nil {
		return fmt.Errorf("unbinding: %w", err)
	}
	fmt.Println("Unbound successfully")

	// 7. Deprovision
	fmt.Printf("\nDeprovisioning instance %s...\n", instanceID)
	_, _, err = client.Deprovision(ctx, instanceID, osbapi.DeprovisionRequest{
		ServiceID: svc.ID,
		PlanID:    plan.ID,
	}, false)
	if err != nil {
		return fmt.Errorf("deprovisioning: %w", err)
	}
	fmt.Println("Deprovisioned successfully")

	fmt.Println("\nWorkflow completed successfully!")
	return nil
}
