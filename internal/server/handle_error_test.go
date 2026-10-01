package server_test

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	osbapi "github.com/fivetwenty-io/osbapi/v2/pkg/osbapi"

	"github.com/fivetwenty-io/osbapi/v2/internal/server"
)

// logEntry is one call recorded by recordingLogger.
type logEntry struct {
	level  string
	msg    string
	fields map[string]any
}

// recordingLogger implements osbapi.Logger and keeps every call so tests can
// assert on what the handler logged. The handler logs from the server's
// goroutine, so access is guarded by a mutex.
type recordingLogger struct {
	mu      sync.Mutex
	entries []logEntry
}

func (l *recordingLogger) record(level, msg string, fields map[string]any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = append(l.entries, logEntry{level: level, msg: msg, fields: fields})
}

func (l *recordingLogger) Debug(msg string, fields map[string]any) { l.record("debug", msg, fields) }
func (l *recordingLogger) Info(msg string, fields map[string]any)  { l.record("info", msg, fields) }
func (l *recordingLogger) Warn(msg string, fields map[string]any)  { l.record("warn", msg, fields) }
func (l *recordingLogger) Error(msg string, fields map[string]any) { l.record("error", msg, fields) }

func (l *recordingLogger) snapshot() []logEntry {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]logEntry(nil), l.entries...)
}

// errValkeyMisconf stands in for a backend failure the library has no
// sentinel for.
var errValkeyMisconf = errors.New("MISCONF Valkey is configured to save RDB snapshots, but it's currently unable to persist to disk")

// TestHandleError_UntypedError_DescriptionAndLog verifies that an error the
// library has no mapping for is answered with a 500 whose description is the
// error's message, and that the error is logged with the request's context.
func TestHandleError_UntypedError_DescriptionAndLog(t *testing.T) {
	t.Parallel()

	unbindErr := fmt.Errorf("unbind binding-1: ping valkey: %w", errValkeyMisconf)
	broker := &mockBroker{unbindErr: unbindErr}
	logger := &recordingLogger{}

	ts := newTestServer(t, broker, server.WithLogger(logger))
	defer ts.Close()

	path := "/v2/service_instances/inst-1/service_bindings/binding-1"
	resp := doRequest(t, http.MethodDelete, ts.URL+path+"?service_id=svc-1&plan_id=plan-1", nil)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	osbErr := decodeBody[osbapi.OSBError](t, resp)
	assert.Equal(t, unbindErr.Error(), osbErr.Description)
	assert.Contains(t, osbErr.Description, "MISCONF")
	assert.Empty(t, osbErr.ErrorCode)

	entries := logger.snapshot()
	require.Len(t, entries, 1)
	entry := entries[0]
	assert.Equal(t, "error", entry.level)
	assert.Equal(t, "broker unbind failed", entry.msg)
	assert.Equal(t, "unbind", entry.fields["operation"])
	assert.Equal(t, http.MethodDelete, entry.fields["method"])
	assert.Equal(t, path, entry.fields["path"])
	assert.Equal(t, http.StatusInternalServerError, entry.fields["status"])
	assert.Equal(t, "inst-1", entry.fields["instance_id"])
	assert.Equal(t, "binding-1", entry.fields["binding_id"])
	assert.Equal(t, unbindErr.Error(), entry.fields["error"])
	assert.Equal(t, "*fmt.wrapError", entry.fields["error_type"])
}

// TestHandleError_InstanceRoute_OmitsBindingID verifies that an instance
// route logs the instance ID and leaves binding_id out.
func TestHandleError_InstanceRoute_OmitsBindingID(t *testing.T) {
	t.Parallel()

	broker := &mockBroker{provisionErr: errValkeyMisconf}
	logger := &recordingLogger{}

	ts := newTestServer(t, broker, server.WithLogger(logger))
	defer ts.Close()

	req := osbapi.ProvisionRequest{ServiceID: "svc-1", PlanID: "plan-1"}
	resp := doRequest(t, http.MethodPut, ts.URL+"/v2/service_instances/inst-1", req)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	osbErr := decodeBody[osbapi.OSBError](t, resp)
	assert.Equal(t, errValkeyMisconf.Error(), osbErr.Description)

	entries := logger.snapshot()
	require.Len(t, entries, 1)
	assert.Equal(t, "provision", entries[0].fields["operation"])
	assert.Equal(t, "inst-1", entries[0].fields["instance_id"])
	assert.NotContains(t, entries[0].fields, "binding_id")
}

// TestHandleError_TypedServerError_Logged verifies that an OSBError without a
// status code still maps to 500, keeps its own body, and is logged.
func TestHandleError_TypedServerError_Logged(t *testing.T) {
	t.Parallel()

	typed := &osbapi.OSBError{ErrorCode: "BackendDown", Description: "valkey is unreachable"}
	broker := &mockBroker{bindErr: fmt.Errorf("bind: %w", typed)}
	logger := &recordingLogger{}

	ts := newTestServer(t, broker, server.WithLogger(logger))
	defer ts.Close()

	req := osbapi.BindRequest{ServiceID: "svc-1", PlanID: "plan-1"}
	resp := doRequest(t, http.MethodPut, ts.URL+"/v2/service_instances/inst-1/service_bindings/binding-1", req)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	osbErr := decodeBody[osbapi.OSBError](t, resp)
	assert.Equal(t, "BackendDown", osbErr.ErrorCode)
	assert.Equal(t, "valkey is unreachable", osbErr.Description)

	entries := logger.snapshot()
	require.Len(t, entries, 1)
	assert.Equal(t, "bind", entries[0].fields["operation"])
	assert.Equal(t, "bind: BackendDown: valkey is unreachable", entries[0].fields["error"])
}

// TestHandleError_ClientErrors_NotLogged verifies that sentinel and typed
// errors that map to 4xx keep their status and are not logged as failures.
func TestHandleError_ClientErrors_NotLogged(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		err    error
		status int
	}{
		{"instance not found", osbapi.ErrInstanceNotFound, http.StatusNotFound},
		{"gone", fmt.Errorf("lookup: %w", osbapi.ErrGoneError), http.StatusGone},
		{"concurrency", osbapi.ErrConcurrencyError, http.StatusUnprocessableEntity},
		{"already exists", osbapi.ErrInstanceAlreadyExists, http.StatusConflict},
		{"bad request", osbapi.ErrBadRequest, http.StatusBadRequest},
		{"typed 4xx", &osbapi.OSBError{Description: "bad plan", StatusCode: http.StatusBadRequest}, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			broker := &mockBroker{fetchInstanceErr: tt.err}
			logger := &recordingLogger{}

			ts := newTestServer(t, broker, server.WithLogger(logger))
			defer ts.Close()

			resp := doRequest(t, http.MethodGet, ts.URL+"/v2/service_instances/inst-1", nil)
			defer resp.Body.Close()
			assert.Equal(t, tt.status, resp.StatusCode)
			assert.Empty(t, logger.snapshot())
		})
	}
}

// TestHandleError_DefaultLogger_UsesSlogDefault verifies that a handler built
// without WithLogger still logs server errors, through slog.Default(). It
// swaps the process-wide default logger, so it must not run in parallel.
func TestHandleError_DefaultLogger_UsesSlogDefault(t *testing.T) {
	var buf bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	broker := &mockBroker{unbindErr: errValkeyMisconf}

	ts := newTestServer(t, broker)
	defer ts.Close()

	resp := doRequest(t, http.MethodDelete, ts.URL+"/v2/service_instances/inst-1/service_bindings/binding-1", nil)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	out := buf.String()
	assert.Contains(t, out, "level=ERROR")
	assert.Contains(t, out, `msg="broker unbind failed"`)
	assert.Contains(t, out, "operation=unbind")
	assert.Contains(t, out, "instance_id=inst-1")
	assert.Contains(t, out, "binding_id=binding-1")
	assert.Contains(t, out, "MISCONF Valkey")
}
