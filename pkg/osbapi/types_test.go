package osbapi_test

import (
	"encoding/json"
	"testing"

	"github.com/fivetwenty-io/osbapi/v2/pkg/osbapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvisionRequest_JSONRoundTrip_Full(t *testing.T) {
	t.Parallel()

	original := osbapi.ProvisionRequest{
		ServiceID:        "service-1",
		PlanID:           "plan-1",
		Context:          map[string]any{"platform": "cloudfoundry"},
		OrganizationGUID: "org-guid-1",
		SpaceGUID:        "space-guid-1",
		Parameters:       map[string]any{"param1": "value1", "param2": float64(42)},
		MaintenanceInfo:  &osbapi.MaintenanceInfo{Version: "1.0.0", Description: "patch update"},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded osbapi.ProvisionRequest
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.ServiceID, decoded.ServiceID)
	assert.Equal(t, original.PlanID, decoded.PlanID)
	assert.Equal(t, original.Context, decoded.Context)
	assert.Equal(t, original.OrganizationGUID, decoded.OrganizationGUID)
	assert.Equal(t, original.SpaceGUID, decoded.SpaceGUID)
	assert.Equal(t, original.Parameters, decoded.Parameters)
	require.NotNil(t, decoded.MaintenanceInfo)
	assert.Equal(t, "1.0.0", decoded.MaintenanceInfo.Version)
	assert.Equal(t, "patch update", decoded.MaintenanceInfo.Description)
}

func TestProvisionRequest_JSONRoundTrip_Minimal(t *testing.T) {
	t.Parallel()

	original := osbapi.ProvisionRequest{
		ServiceID:        "service-1",
		PlanID:           "plan-1",
		OrganizationGUID: "org-guid-1",
		SpaceGUID:        "space-guid-1",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded osbapi.ProvisionRequest
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.ServiceID, decoded.ServiceID)
	assert.Equal(t, original.PlanID, decoded.PlanID)
	assert.Equal(t, original.OrganizationGUID, decoded.OrganizationGUID)
	assert.Equal(t, original.SpaceGUID, decoded.SpaceGUID)
	assert.Nil(t, decoded.Context)
	assert.Nil(t, decoded.Parameters)
	assert.Nil(t, decoded.MaintenanceInfo)

	// Verify omitempty fields are absent from JSON
	var raw map[string]json.RawMessage
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)
	assert.NotContains(t, raw, "context")
	assert.NotContains(t, raw, "parameters")
	assert.NotContains(t, raw, "maintenance_info")
}

func TestUpdateRequest_JSONRoundTrip_WithPreviousValues(t *testing.T) {
	t.Parallel()

	original := osbapi.UpdateRequest{
		Context:    map[string]any{"platform": "kubernetes"},
		ServiceID:  "service-1",
		PlanID:     "plan-2",
		Parameters: map[string]any{"key": "value"},
		PreviousValues: &osbapi.PreviousValues{
			ServiceID:      "service-1",
			PlanID:         "plan-1",
			OrganizationID: "org-1",
			SpaceID:        "space-1",
			MaintenanceInfo: &osbapi.MaintenanceInfo{
				Version: "0.9.0",
			},
		},
		MaintenanceInfo: &osbapi.MaintenanceInfo{
			Version: "1.0.0",
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded osbapi.UpdateRequest
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.ServiceID, decoded.ServiceID)
	assert.Equal(t, original.PlanID, decoded.PlanID)
	assert.Equal(t, original.Context, decoded.Context)
	assert.Equal(t, original.Parameters, decoded.Parameters)

	require.NotNil(t, decoded.PreviousValues)
	assert.Equal(t, "service-1", decoded.PreviousValues.ServiceID)
	assert.Equal(t, "plan-1", decoded.PreviousValues.PlanID)
	assert.Equal(t, "org-1", decoded.PreviousValues.OrganizationID)
	assert.Equal(t, "space-1", decoded.PreviousValues.SpaceID)
	require.NotNil(t, decoded.PreviousValues.MaintenanceInfo)
	assert.Equal(t, "0.9.0", decoded.PreviousValues.MaintenanceInfo.Version)

	require.NotNil(t, decoded.MaintenanceInfo)
	assert.Equal(t, "1.0.0", decoded.MaintenanceInfo.Version)
}

func TestLastOperationResponse_JSONRoundTrip_AllStates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		state osbapi.OperationState
		desc  string
	}{
		{
			name:  "in progress",
			state: osbapi.StateInProgress,
			desc:  "creating service instance",
		},
		{
			name:  "succeeded",
			state: osbapi.StateSucceeded,
			desc:  "service instance created",
		},
		{
			name:  "failed",
			state: osbapi.StateFailed,
			desc:  "failed to create service instance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			original := osbapi.LastOperationResponse{
				State:       tt.state,
				Description: tt.desc,
			}

			data, err := json.Marshal(original)
			require.NoError(t, err)

			var decoded osbapi.LastOperationResponse
			err = json.Unmarshal(data, &decoded)
			require.NoError(t, err)

			assert.Equal(t, original.State, decoded.State)
			assert.Equal(t, original.Description, decoded.Description)
		})
	}
}

func TestOperationState_Constants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, osbapi.OperationState("in progress"), osbapi.StateInProgress)
	assert.Equal(t, osbapi.OperationState("succeeded"), osbapi.StateSucceeded)
	assert.Equal(t, osbapi.OperationState("failed"), osbapi.StateFailed)
}
