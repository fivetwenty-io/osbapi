//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCatalog_RetrievesCatalog(t *testing.T) {
	t.Parallel()

	_, client := StartTestBroker(t)
	ctx := context.Background()

	catalog, err := client.GetCatalog(ctx)
	require.NoError(t, err)
	require.NotNil(t, catalog)
	assert.NotEmpty(t, catalog.Services, "catalog should contain at least one service")
}

func TestCatalog_HasExpectedService(t *testing.T) {
	t.Parallel()

	_, client := StartTestBroker(t)
	ctx := context.Background()

	catalog, err := client.GetCatalog(ctx)
	require.NoError(t, err)
	require.Len(t, catalog.Services, 1)

	svc := catalog.Services[0]
	assert.Equal(t, testServiceID, svc.ID)
	assert.Equal(t, "example-service", svc.Name)
	assert.True(t, svc.Bindable, "service should be bindable")
	assert.True(t, svc.InstancesRetrievable, "service should be instances retrievable")
	assert.True(t, svc.BindingsRetrievable, "service should be bindings retrievable")
	assert.NotNil(t, svc.PlanUpdateable)
	assert.True(t, *svc.PlanUpdateable, "service should be plan updateable")
	assert.Len(t, svc.Plans, 2, "service should have 2 plans")
	assert.Contains(t, svc.Tags, "example")
	assert.Contains(t, svc.Tags, "demo")
}

func TestCatalog_PlanDetails(t *testing.T) {
	t.Parallel()

	_, client := StartTestBroker(t)
	ctx := context.Background()

	catalog, err := client.GetCatalog(ctx)
	require.NoError(t, err)
	require.Len(t, catalog.Services, 1)

	plans := catalog.Services[0].Plans
	require.Len(t, plans, 2)

	// Find plans by name for stable assertions.
	var freePlan, paidPlan *struct {
		ID          string
		Name        string
		Description string
		Free        *bool
	}

	for i := range plans {
		p := plans[i]
		switch p.Name {
		case "free":
			freePlan = &struct {
				ID          string
				Name        string
				Description string
				Free        *bool
			}{p.ID, p.Name, p.Description, p.Free}
		case "paid":
			paidPlan = &struct {
				ID          string
				Name        string
				Description string
				Free        *bool
			}{p.ID, p.Name, p.Description, p.Free}
		}
	}

	require.NotNil(t, freePlan, "free plan should exist")
	assert.Equal(t, freePlanID, freePlan.ID)
	assert.Equal(t, "free", freePlan.Name)
	assert.NotEmpty(t, freePlan.Description)
	require.NotNil(t, freePlan.Free)
	assert.True(t, *freePlan.Free, "free plan should be free")

	require.NotNil(t, paidPlan, "paid plan should exist")
	assert.Equal(t, paidPlanID, paidPlan.ID)
	assert.Equal(t, "paid", paidPlan.Name)
	assert.NotEmpty(t, paidPlan.Description)
	require.NotNil(t, paidPlan.Free)
	assert.False(t, *paidPlan.Free, "paid plan should not be free")
}
