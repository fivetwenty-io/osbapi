package osbapi_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/fivetwenty-io/osbapi/v2/pkg/osbapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOSBError_ErrorWithCode(t *testing.T) {
	t.Parallel()

	osbErr := &osbapi.OSBError{
		ErrorCode:   "AsyncRequired",
		Description: "This broker requires async operations",
		StatusCode:  422,
	}

	assert.Equal(t, "AsyncRequired: This broker requires async operations", osbErr.Error())
}

func TestOSBError_ErrorWithoutCode(t *testing.T) {
	t.Parallel()

	osbErr := &osbapi.OSBError{
		Description: "Something went wrong",
		StatusCode:  500,
	}

	assert.Equal(t, "Something went wrong", osbErr.Error())
}

func TestOSBError_ImplementsErrorInterface(t *testing.T) {
	t.Parallel()

	var err error = &osbapi.OSBError{
		ErrorCode:   "TestError",
		Description: "test",
	}

	assert.Error(t, err)
}

func TestOSBError_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	usable := true
	repeatable := false

	original := &osbapi.OSBError{
		ErrorCode:        "ConcurrencyError",
		Description:      "operation in progress",
		InstanceUsable:   &usable,
		UpdateRepeatable: &repeatable,
		StatusCode:       422,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	// Verify JSON structure
	var raw map[string]any
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)

	assert.Equal(t, "ConcurrencyError", raw["error"])
	assert.Equal(t, "operation in progress", raw["description"])
	assert.Equal(t, true, raw["instance_usable"])
	assert.Equal(t, false, raw["update_repeatable"])

	// StatusCode should not appear in JSON (has json:"-" tag)
	assert.NotContains(t, raw, "StatusCode")
	assert.NotContains(t, raw, "status_code")

	// Unmarshal back
	var decoded osbapi.OSBError
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.ErrorCode, decoded.ErrorCode)
	assert.Equal(t, original.Description, decoded.Description)
	require.NotNil(t, decoded.InstanceUsable)
	assert.True(t, *decoded.InstanceUsable)
	require.NotNil(t, decoded.UpdateRepeatable)
	assert.False(t, *decoded.UpdateRepeatable)

	// StatusCode is not marshaled, so it will be zero after round-trip
	assert.Zero(t, decoded.StatusCode)
}

func TestOSBError_JSONOmitsEmptyFields(t *testing.T) {
	t.Parallel()

	osbErr := &osbapi.OSBError{}

	data, err := json.Marshal(osbErr)
	require.NoError(t, err)

	var raw map[string]any
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)

	assert.NotContains(t, raw, "error")
	assert.NotContains(t, raw, "description")
	assert.NotContains(t, raw, "instance_usable")
	assert.NotContains(t, raw, "update_repeatable")
}

func TestIsAsyncRequired_SentinelError(t *testing.T) {
	t.Parallel()

	assert.True(t, osbapi.IsAsyncRequired(osbapi.ErrAsyncRequired))
}

func TestIsAsyncRequired_WrappedSentinelError(t *testing.T) {
	t.Parallel()

	wrapped := fmt.Errorf("provision failed: %w", osbapi.ErrAsyncRequired)
	assert.True(t, osbapi.IsAsyncRequired(wrapped))
}

func TestIsAsyncRequired_OSBError(t *testing.T) {
	t.Parallel()

	osbErr := &osbapi.OSBError{
		ErrorCode:   osbapi.ErrorCodeAsyncRequired,
		Description: "async required",
		StatusCode:  422,
	}

	assert.True(t, osbapi.IsAsyncRequired(osbErr))
}

func TestIsAsyncRequired_UnrelatedError(t *testing.T) {
	t.Parallel()

	assert.False(t, osbapi.IsAsyncRequired(errors.New("unrelated error")))
}

func TestIsAsyncRequired_DifferentOSBError(t *testing.T) {
	t.Parallel()

	osbErr := &osbapi.OSBError{
		ErrorCode:   osbapi.ErrorCodeConcurrencyError,
		Description: "concurrent",
		StatusCode:  422,
	}

	assert.False(t, osbapi.IsAsyncRequired(osbErr))
}

func TestIsConcurrencyError_SentinelError(t *testing.T) {
	t.Parallel()

	assert.True(t, osbapi.IsConcurrencyError(osbapi.ErrConcurrencyError))
}

func TestIsConcurrencyError_WrappedSentinelError(t *testing.T) {
	t.Parallel()

	wrapped := fmt.Errorf("update failed: %w", osbapi.ErrConcurrencyError)
	assert.True(t, osbapi.IsConcurrencyError(wrapped))
}

func TestIsConcurrencyError_OSBError(t *testing.T) {
	t.Parallel()

	osbErr := &osbapi.OSBError{
		ErrorCode:   osbapi.ErrorCodeConcurrencyError,
		Description: "concurrent operation",
		StatusCode:  422,
	}

	assert.True(t, osbapi.IsConcurrencyError(osbErr))
}

func TestIsConcurrencyError_UnrelatedError(t *testing.T) {
	t.Parallel()

	assert.False(t, osbapi.IsConcurrencyError(errors.New("something else")))
}

func TestIsNotFound_InstanceNotFound(t *testing.T) {
	t.Parallel()

	assert.True(t, osbapi.IsNotFound(osbapi.ErrInstanceNotFound))
}

func TestIsNotFound_BindingNotFound(t *testing.T) {
	t.Parallel()

	assert.True(t, osbapi.IsNotFound(osbapi.ErrBindingNotFound))
}

func TestIsNotFound_WrappedError(t *testing.T) {
	t.Parallel()

	wrapped := fmt.Errorf("fetch failed: %w", osbapi.ErrInstanceNotFound)
	assert.True(t, osbapi.IsNotFound(wrapped))
}

func TestIsNotFound_OSBError404(t *testing.T) {
	t.Parallel()

	osbErr := &osbapi.OSBError{
		Description: "not found",
		StatusCode:  404,
	}

	assert.True(t, osbapi.IsNotFound(osbErr))
}

func TestIsNotFound_OSBError410(t *testing.T) {
	t.Parallel()

	osbErr := &osbapi.OSBError{
		Description: "gone",
		StatusCode:  410,
	}

	assert.True(t, osbapi.IsNotFound(osbErr))
}

func TestIsNotFound_UnrelatedError(t *testing.T) {
	t.Parallel()

	assert.False(t, osbapi.IsNotFound(errors.New("other error")))
}

func TestIsGone_SentinelError(t *testing.T) {
	t.Parallel()

	assert.True(t, osbapi.IsGone(osbapi.ErrGoneError))
}

func TestIsGone_WrappedSentinelError(t *testing.T) {
	t.Parallel()

	wrapped := fmt.Errorf("deprovision check: %w", osbapi.ErrGoneError)
	assert.True(t, osbapi.IsGone(wrapped))
}

func TestIsGone_OSBError410(t *testing.T) {
	t.Parallel()

	osbErr := &osbapi.OSBError{
		Description: "resource deleted",
		StatusCode:  410,
	}

	assert.True(t, osbapi.IsGone(osbErr))
}

func TestIsGone_OSBError404(t *testing.T) {
	t.Parallel()

	// 404 is not "gone" - it's just "not found"
	osbErr := &osbapi.OSBError{
		Description: "not found",
		StatusCode:  404,
	}

	assert.False(t, osbapi.IsGone(osbErr))
}

func TestIsGone_UnrelatedError(t *testing.T) {
	t.Parallel()

	assert.False(t, osbapi.IsGone(errors.New("some error")))
}

func TestBoolPtr(t *testing.T) {
	t.Parallel()

	truePtr := osbapi.BoolPtr(true)
	require.NotNil(t, truePtr)
	assert.True(t, *truePtr)

	falsePtr := osbapi.BoolPtr(false)
	require.NotNil(t, falsePtr)
	assert.False(t, *falsePtr)
}

func TestBoolPtr_ReturnsDistinctPointers(t *testing.T) {
	t.Parallel()

	p1 := osbapi.BoolPtr(true)
	p2 := osbapi.BoolPtr(true)

	// Each call should return a new pointer
	assert.NotSame(t, p1, p2)
	assert.Equal(t, *p1, *p2)
}

func TestSentinelErrors_AreDistinct(t *testing.T) {
	t.Parallel()

	sentinels := []error{
		osbapi.ErrAsyncRequired,
		osbapi.ErrConcurrencyError,
		osbapi.ErrMaintenanceInfoConflict,
		osbapi.ErrRequiresApp,
		osbapi.ErrInstanceNotFound,
		osbapi.ErrBindingNotFound,
		osbapi.ErrInstanceAlreadyExists,
		osbapi.ErrBindingAlreadyExists,
		osbapi.ErrGoneError,
		osbapi.ErrBadRequest,
		osbapi.ErrUnauthorized,
		osbapi.ErrPlanQuotaExceeded,
		osbapi.ErrInvalidParameters,
	}

	for i, a := range sentinels {
		for j, b := range sentinels {
			if i != j {
				assert.NotErrorIs(t, a, b,
					"sentinel errors %d and %d should be distinct", i, j)
			}
		}
	}
}

func TestErrorCodeConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "AsyncRequired", osbapi.ErrorCodeAsyncRequired)
	assert.Equal(t, "ConcurrencyError", osbapi.ErrorCodeConcurrencyError)
	assert.Equal(t, "MaintenanceInfoConflict", osbapi.ErrorCodeMaintenanceInfoConflict)
	assert.Equal(t, "RequiresApp", osbapi.ErrorCodeRequiresApp)
}
