package osbapi_test

import (
	"testing"

	"github.com/fivetwenty-io/osbapi/v2/pkg/osbapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaderConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "X-Broker-API-Version", osbapi.HeaderAPIVersion)
	assert.Equal(t, "X-Broker-API-Originating-Identity", osbapi.HeaderOriginatingIdentity)
	assert.Equal(t, "X-Broker-API-Request-Identity", osbapi.HeaderRequestIdentity)
	assert.Equal(t, "2.17", osbapi.APIVersion)
}

func TestOriginatingIdentity_RoundTrip(t *testing.T) {
	t.Parallel()

	original := osbapi.OriginatingIdentity{
		Platform: "cloudfoundry",
		Value: map[string]any{
			"user_id": "683ea748-3092-4ff4-b656-39cacc4d5360",
		},
	}

	encoded, err := osbapi.EncodeOriginatingIdentity(original)
	require.NoError(t, err)

	// Verify the encoded string starts with the platform
	assert.Contains(t, encoded, "cloudfoundry ")

	decoded, err := osbapi.DecodeOriginatingIdentity(encoded)
	require.NoError(t, err)

	assert.Equal(t, original.Platform, decoded.Platform)
	assert.Equal(t, original.Value["user_id"], decoded.Value["user_id"])
}

func TestOriginatingIdentity_CloudFoundryPlatform(t *testing.T) {
	t.Parallel()

	identity := osbapi.OriginatingIdentity{
		Platform: "cloudfoundry",
		Value: map[string]any{
			"user_id": "683ea748-3092-4ff4-b656-39cacc4d5360",
		},
	}

	encoded, err := osbapi.EncodeOriginatingIdentity(identity)
	require.NoError(t, err)

	decoded, err := osbapi.DecodeOriginatingIdentity(encoded)
	require.NoError(t, err)

	assert.Equal(t, "cloudfoundry", decoded.Platform)
	assert.Equal(t, "683ea748-3092-4ff4-b656-39cacc4d5360", decoded.Value["user_id"])
}

func TestOriginatingIdentity_KubernetesPlatform(t *testing.T) {
	t.Parallel()

	identity := osbapi.OriginatingIdentity{
		Platform: "kubernetes",
		Value: map[string]any{
			"username": "admin",
			"uid":      "abc-123",
			"groups":   []any{"system:masters"},
		},
	}

	encoded, err := osbapi.EncodeOriginatingIdentity(identity)
	require.NoError(t, err)

	decoded, err := osbapi.DecodeOriginatingIdentity(encoded)
	require.NoError(t, err)

	assert.Equal(t, "kubernetes", decoded.Platform)
	assert.Equal(t, "admin", decoded.Value["username"])
	assert.Equal(t, "abc-123", decoded.Value["uid"])

	groups, ok := decoded.Value["groups"].([]any)
	require.True(t, ok)
	require.Len(t, groups, 1)
	assert.Equal(t, "system:masters", groups[0])
}

func TestOriginatingIdentity_MultipleFields(t *testing.T) {
	t.Parallel()

	identity := osbapi.OriginatingIdentity{
		Platform: "cloudfoundry",
		Value: map[string]any{
			"user_id": "user-123",
			"email":   "user@example.com",
			"org_id":  "org-456",
		},
	}

	encoded, err := osbapi.EncodeOriginatingIdentity(identity)
	require.NoError(t, err)

	decoded, err := osbapi.DecodeOriginatingIdentity(encoded)
	require.NoError(t, err)

	assert.Equal(t, "cloudfoundry", decoded.Platform)
	assert.Equal(t, "user-123", decoded.Value["user_id"])
	assert.Equal(t, "user@example.com", decoded.Value["email"])
	assert.Equal(t, "org-456", decoded.Value["org_id"])
}

func TestDecodeOriginatingIdentity_MissingSpace(t *testing.T) {
	t.Parallel()

	_, err := osbapi.DecodeOriginatingIdentity("cloudfoundry-no-space")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing space separator")
}

func TestDecodeOriginatingIdentity_InvalidBase64(t *testing.T) {
	t.Parallel()

	_, err := osbapi.DecodeOriginatingIdentity("cloudfoundry !!!not-base64!!!")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decoding originating identity value")
}

func TestDecodeOriginatingIdentity_InvalidJSON(t *testing.T) {
	t.Parallel()

	// Base64 encode invalid JSON
	// "not json" in base64 = "bm90IGpzb24="
	_, err := osbapi.DecodeOriginatingIdentity("cloudfoundry bm90IGpzb24=")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshaling originating identity value")
}

func TestDecodeOriginatingIdentity_EmptyValue(t *testing.T) {
	t.Parallel()

	// Base64 encode "{}" = "e30="
	decoded, err := osbapi.DecodeOriginatingIdentity("cloudfoundry e30=")
	require.NoError(t, err)

	assert.Equal(t, "cloudfoundry", decoded.Platform)
	assert.Empty(t, decoded.Value)
}

func TestEncodeOriginatingIdentity_EmptyValue(t *testing.T) {
	t.Parallel()

	identity := osbapi.OriginatingIdentity{
		Platform: "test-platform",
		Value:    map[string]any{},
	}

	encoded, err := osbapi.EncodeOriginatingIdentity(identity)
	require.NoError(t, err)

	decoded, err := osbapi.DecodeOriginatingIdentity(encoded)
	require.NoError(t, err)

	assert.Equal(t, "test-platform", decoded.Platform)
	assert.Empty(t, decoded.Value)
}
