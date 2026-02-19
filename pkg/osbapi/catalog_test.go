package osbapi_test

import (
	"encoding/json"
	"testing"

	"github.com/fivetwenty-io/osbapi/v2/pkg/osbapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func boolPtr(b bool) *bool { return &b }

func TestCatalog_FullRoundTrip(t *testing.T) {
	t.Parallel()

	original := osbapi.Catalog{
		Services: []osbapi.Service{
			{
				ID:                   "service-1",
				Name:                 "my-service",
				Description:          "A test service",
				Tags:                 []string{"tag1", "tag2"},
				Requires:             []string{"syslog_drain"},
				Bindable:             true,
				InstancesRetrievable: true,
				BindingsRetrievable:  true,
				AllowContextUpdates:  true,
				Metadata: map[string]any{
					"displayName":     "My Service",
					"longDescription": "A longer description of the service",
				},
				DashboardClient: &osbapi.DashboardClient{
					ID:          "client-id",
					Secret:      "client-secret",
					RedirectURI: "https://example.com/oauth/callback",
				},
				PlanUpdateable: boolPtr(true),
				Plans: []osbapi.Plan{
					{
						ID:          "plan-1",
						Name:        "small",
						Description: "A small plan",
						Metadata: map[string]any{
							"bullets": []any{"5 GB storage", "10 connections"},
						},
						Free:           boolPtr(true),
						Bindable:       boolPtr(true),
						PlanUpdateable: boolPtr(false),
						Schemas: &osbapi.Schemas{
							ServiceInstance: &osbapi.ServiceInstanceSchema{
								Create: &osbapi.InputParametersSchema{
									Parameters: map[string]any{
										"type": "object",
										"properties": map[string]any{
											"billing-account": map[string]any{
												"description": "Billing account number",
												"type":        "string",
											},
										},
									},
								},
								Update: &osbapi.InputParametersSchema{
									Parameters: map[string]any{
										"type": "object",
									},
								},
							},
							ServiceBinding: &osbapi.ServiceBindingSchema{
								Create: &osbapi.InputParametersSchema{
									Parameters: map[string]any{
										"type": "object",
									},
								},
							},
						},
						MaximumPollingDuration: 3600,
						MaintenanceInfo: &osbapi.MaintenanceInfo{
							Version:     "1.0.0",
							Description: "OS update",
						},
					},
				},
				MaintenanceInfo: &osbapi.MaintenanceInfo{
					Version:     "2.1.0",
					Description: "Service-level maintenance",
				},
			},
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err, "marshal should succeed")

	var decoded osbapi.Catalog
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err, "unmarshal should succeed")

	assert.Equal(t, original, decoded)
}

func TestCatalog_OmitsEmptyOptionalFields(t *testing.T) {
	t.Parallel()

	catalog := osbapi.Catalog{
		Services: []osbapi.Service{
			{
				ID:          "svc-1",
				Name:        "minimal",
				Description: "Minimal service",
				Bindable:    false,
				Plans: []osbapi.Plan{
					{
						ID:          "plan-1",
						Name:        "default",
						Description: "Default plan",
					},
				},
			},
		},
	}

	data, err := json.Marshal(catalog)
	require.NoError(t, err)

	var raw map[string]any
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)

	services := raw["services"].([]any)
	svc := services[0].(map[string]any)

	// Required fields must be present
	assert.Contains(t, svc, "id")
	assert.Contains(t, svc, "name")
	assert.Contains(t, svc, "description")
	assert.Contains(t, svc, "plans")

	// Optional fields must be absent
	assert.NotContains(t, svc, "tags")
	assert.NotContains(t, svc, "requires")
	assert.NotContains(t, svc, "metadata")
	assert.NotContains(t, svc, "dashboard_client")
	assert.NotContains(t, svc, "plan_updateable")
	assert.NotContains(t, svc, "maintenance_info")

	plan := svc["plans"].([]any)[0].(map[string]any)

	// Plan optional fields must be absent
	assert.NotContains(t, plan, "metadata")
	assert.NotContains(t, plan, "free")
	assert.NotContains(t, plan, "bindable")
	assert.NotContains(t, plan, "plan_updateable")
	assert.NotContains(t, plan, "schemas")
	assert.NotContains(t, plan, "maximum_polling_duration")
	assert.NotContains(t, plan, "maintenance_info")
}

func TestCatalog_BoolPointerFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		free     *bool
		bindable *bool
		wantFree any
		wantBind any
	}{
		{
			name:     "nil pointers are omitted",
			free:     nil,
			bindable: nil,
			wantFree: nil,
			wantBind: nil,
		},
		{
			name:     "true values are present",
			free:     boolPtr(true),
			bindable: boolPtr(true),
			wantFree: true,
			wantBind: true,
		},
		{
			name:     "false values are present",
			free:     boolPtr(false),
			bindable: boolPtr(false),
			wantFree: false,
			wantBind: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			plan := osbapi.Plan{
				ID:          "plan-1",
				Name:        "test",
				Description: "test plan",
				Free:        tt.free,
				Bindable:    tt.bindable,
			}

			data, err := json.Marshal(plan)
			require.NoError(t, err)

			var raw map[string]any
			err = json.Unmarshal(data, &raw)
			require.NoError(t, err)

			if tt.wantFree == nil {
				assert.NotContains(t, raw, "free",
					"nil *bool should be omitted from JSON")
			} else {
				require.Contains(t, raw, "free")
				assert.Equal(t, tt.wantFree, raw["free"])
			}

			if tt.wantBind == nil {
				assert.NotContains(t, raw, "bindable",
					"nil *bool should be omitted from JSON")
			} else {
				require.Contains(t, raw, "bindable")
				assert.Equal(t, tt.wantBind, raw["bindable"])
			}

			// Round-trip: unmarshal back and verify pointer values match
			var decoded osbapi.Plan
			err = json.Unmarshal(data, &decoded)
			require.NoError(t, err)
			assert.Equal(t, plan, decoded)
		})
	}
}

func TestCatalog_PlanUpdateablePointer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		planUpdateable *bool
		expectPresent  bool
		expectValue    bool
	}{
		{
			name:           "nil is omitted",
			planUpdateable: nil,
			expectPresent:  false,
		},
		{
			name:           "true is serialized",
			planUpdateable: boolPtr(true),
			expectPresent:  true,
			expectValue:    true,
		},
		{
			name:           "false is serialized",
			planUpdateable: boolPtr(false),
			expectPresent:  true,
			expectValue:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := osbapi.Service{
				ID:             "svc-1",
				Name:           "test",
				Description:    "test service",
				Bindable:       true,
				PlanUpdateable: tt.planUpdateable,
				Plans: []osbapi.Plan{
					{ID: "p1", Name: "default", Description: "default"},
				},
			}

			data, err := json.Marshal(svc)
			require.NoError(t, err)

			var raw map[string]any
			err = json.Unmarshal(data, &raw)
			require.NoError(t, err)

			if !tt.expectPresent {
				assert.NotContains(t, raw, "plan_updateable")
			} else {
				require.Contains(t, raw, "plan_updateable")
				assert.Equal(t, tt.expectValue, raw["plan_updateable"])
			}

			// Round-trip
			var decoded osbapi.Service
			err = json.Unmarshal(data, &decoded)
			require.NoError(t, err)
			assert.Equal(t, svc, decoded)
		})
	}
}

func TestCatalog_UnmarshalFromJSON(t *testing.T) {
	t.Parallel()

	raw := `{
		"services": [{
			"id": "svc-abc",
			"name": "example-service",
			"description": "An example",
			"bindable": true,
			"plan_updateable": false,
			"plans": [{
				"id": "plan-xyz",
				"name": "standard",
				"description": "Standard plan",
				"free": true,
				"bindable": false,
				"maximum_polling_duration": 7200,
				"schemas": {
					"service_instance": {
						"create": {
							"parameters": {
								"type": "object",
								"properties": {
									"region": {"type": "string"}
								}
							}
						}
					}
				}
			}],
			"dashboard_client": {
				"id": "dash-id",
				"secret": "dash-secret",
				"redirect_uri": "https://dash.example.com/callback"
			},
			"maintenance_info": {
				"version": "3.0.0"
			}
		}]
	}`

	var catalog osbapi.Catalog
	err := json.Unmarshal([]byte(raw), &catalog)
	require.NoError(t, err)

	require.Len(t, catalog.Services, 1)
	svc := catalog.Services[0]

	assert.Equal(t, "svc-abc", svc.ID)
	assert.Equal(t, "example-service", svc.Name)
	assert.True(t, svc.Bindable)

	require.NotNil(t, svc.PlanUpdateable)
	assert.False(t, *svc.PlanUpdateable)

	require.NotNil(t, svc.DashboardClient)
	assert.Equal(t, "dash-id", svc.DashboardClient.ID)
	assert.Equal(t, "dash-secret", svc.DashboardClient.Secret)
	assert.Equal(t, "https://dash.example.com/callback", svc.DashboardClient.RedirectURI)

	require.NotNil(t, svc.MaintenanceInfo)
	assert.Equal(t, "3.0.0", svc.MaintenanceInfo.Version)
	assert.Empty(t, svc.MaintenanceInfo.Description)

	require.Len(t, svc.Plans, 1)
	plan := svc.Plans[0]

	assert.Equal(t, "plan-xyz", plan.ID)
	assert.Equal(t, "standard", plan.Name)

	require.NotNil(t, plan.Free)
	assert.True(t, *plan.Free)

	require.NotNil(t, plan.Bindable)
	assert.False(t, *plan.Bindable)

	assert.Equal(t, 7200, plan.MaximumPollingDuration)

	require.NotNil(t, plan.Schemas)
	require.NotNil(t, plan.Schemas.ServiceInstance)
	require.NotNil(t, plan.Schemas.ServiceInstance.Create)
	assert.Contains(t, plan.Schemas.ServiceInstance.Create.Parameters, "type")
	assert.Nil(t, plan.Schemas.ServiceInstance.Update)
	assert.Nil(t, plan.Schemas.ServiceBinding)
}
