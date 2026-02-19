package osbapi_test

import (
	"encoding/json"
	"testing"

	"github.com/fivetwenty-io/osbapi/v2/pkg/osbapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBindRequest_JSONRoundTrip_WithBindResource(t *testing.T) {
	t.Parallel()

	original := osbapi.BindRequest{
		Context:   map[string]any{"platform": "cloudfoundry"},
		ServiceID: "service-1",
		PlanID:    "plan-1",
		AppGUID:   "app-guid-1",
		BindResource: &osbapi.BindResource{
			AppGUID:     "app-guid-1",
			Route:       "example.com",
			BackupAgent: true,
		},
		Parameters: map[string]any{"role": "admin"},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded osbapi.BindRequest
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.ServiceID, decoded.ServiceID)
	assert.Equal(t, original.PlanID, decoded.PlanID)
	assert.Equal(t, original.AppGUID, decoded.AppGUID)
	assert.Equal(t, original.Context, decoded.Context)
	assert.Equal(t, original.Parameters, decoded.Parameters)

	require.NotNil(t, decoded.BindResource)
	assert.Equal(t, "app-guid-1", decoded.BindResource.AppGUID)
	assert.Equal(t, "example.com", decoded.BindResource.Route)
	assert.True(t, decoded.BindResource.BackupAgent)
}

func TestBindResponse_JSONRoundTrip_Full(t *testing.T) {
	t.Parallel()

	original := osbapi.BindResponse{
		Credentials: map[string]any{
			"uri":      "postgres://user:pass@host:5432/db",
			"username": "user",
			"password": "pass",
		},
		SyslogDrainURL:  "syslog://logs.example.com:514",
		RouteServiceURL: "https://route.example.com",
		VolumeMounts: []osbapi.VolumeMount{
			{
				Driver:       "nfs",
				ContainerDir: "/var/data",
				Mode:         "rw",
				DeviceType:   "shared",
				Device: osbapi.VolumeMountDevice{
					VolumeID:    "vol-1",
					MountConfig: map[string]any{"source": "nfs://server/share"},
				},
			},
		},
		Endpoints: []osbapi.Endpoint{
			{
				Host:     "db.example.com",
				Ports:    []string{"5432", "5433"},
				Protocol: "tcp",
			},
		},
		Operation: "binding-op-1",
		Metadata: &osbapi.BindingMetadata{
			ExpiresAt:   "2026-12-31T23:59:59Z",
			RenewBefore: "2026-12-01T00:00:00Z",
			Labels:      map[string]string{"env": "production"},
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded osbapi.BindResponse
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.Credentials, decoded.Credentials)
	assert.Equal(t, original.SyslogDrainURL, decoded.SyslogDrainURL)
	assert.Equal(t, original.RouteServiceURL, decoded.RouteServiceURL)
	assert.Equal(t, original.Operation, decoded.Operation)

	require.Len(t, decoded.VolumeMounts, 1)
	vm := decoded.VolumeMounts[0]
	assert.Equal(t, "nfs", vm.Driver)
	assert.Equal(t, "/var/data", vm.ContainerDir)
	assert.Equal(t, "rw", vm.Mode)
	assert.Equal(t, "shared", vm.DeviceType)
	assert.Equal(t, "vol-1", vm.Device.VolumeID)
	assert.Equal(t, map[string]any{"source": "nfs://server/share"}, vm.Device.MountConfig)

	require.Len(t, decoded.Endpoints, 1)
	ep := decoded.Endpoints[0]
	assert.Equal(t, "db.example.com", ep.Host)
	assert.Equal(t, []string{"5432", "5433"}, ep.Ports)
	assert.Equal(t, "tcp", ep.Protocol)

	require.NotNil(t, decoded.Metadata)
	assert.Equal(t, "2026-12-31T23:59:59Z", decoded.Metadata.ExpiresAt)
	assert.Equal(t, "2026-12-01T00:00:00Z", decoded.Metadata.RenewBefore)
	assert.Equal(t, map[string]string{"env": "production"}, decoded.Metadata.Labels)
}

func TestFetchBindingResponse_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	original := osbapi.FetchBindingResponse{
		Credentials: map[string]any{
			"host":     "db.example.com",
			"port":     float64(5432),
			"username": "admin",
		},
		SyslogDrainURL:  "syslog://logs.example.com",
		RouteServiceURL: "https://route.example.com",
		VolumeMounts: []osbapi.VolumeMount{
			{
				Driver:       "cephfs",
				ContainerDir: "/mnt/data",
				Mode:         "r",
				DeviceType:   "shared",
				Device: osbapi.VolumeMountDevice{
					VolumeID: "vol-2",
				},
			},
		},
		Endpoints: []osbapi.Endpoint{
			{
				Host:  "api.example.com",
				Ports: []string{"443"},
			},
		},
		Parameters: map[string]any{"read_only": true},
		Metadata: &osbapi.BindingMetadata{
			Labels: map[string]string{"tier": "standard"},
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded osbapi.FetchBindingResponse
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.Credentials, decoded.Credentials)
	assert.Equal(t, original.SyslogDrainURL, decoded.SyslogDrainURL)
	assert.Equal(t, original.RouteServiceURL, decoded.RouteServiceURL)
	assert.Equal(t, original.Parameters, decoded.Parameters)

	require.Len(t, decoded.VolumeMounts, 1)
	assert.Equal(t, "cephfs", decoded.VolumeMounts[0].Driver)
	assert.Equal(t, "/mnt/data", decoded.VolumeMounts[0].ContainerDir)
	assert.Equal(t, "vol-2", decoded.VolumeMounts[0].Device.VolumeID)
	assert.Nil(t, decoded.VolumeMounts[0].Device.MountConfig)

	require.Len(t, decoded.Endpoints, 1)
	assert.Equal(t, "api.example.com", decoded.Endpoints[0].Host)
	assert.Equal(t, []string{"443"}, decoded.Endpoints[0].Ports)
	assert.Empty(t, decoded.Endpoints[0].Protocol)

	require.NotNil(t, decoded.Metadata)
	assert.Equal(t, map[string]string{"tier": "standard"}, decoded.Metadata.Labels)
}
