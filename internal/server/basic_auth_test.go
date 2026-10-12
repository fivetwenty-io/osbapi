package server_test

import (
	"net/http"
	"testing"

	osbapi "github.com/fivetwenty-io/osbapi/v2/pkg/osbapi"

	"github.com/fivetwenty-io/osbapi/v2/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBasicAuthCredentials checks every way a basic-auth header can be right
// or wrong against a handler configured with WithBasicAuth.
func TestBasicAuthCredentials(t *testing.T) {
	t.Parallel()

	broker := &mockBroker{
		catalog: &osbapi.Catalog{},
	}

	ts := newTestServer(t, broker, server.WithBasicAuth("admin", "secret"))
	t.Cleanup(ts.Close)

	tests := []struct {
		name     string
		setAuth  bool
		username string
		password string
		want     int
	}{
		{"correct credentials", true, "admin", "secret", http.StatusOK},
		{"no credentials", false, "", "", http.StatusUnauthorized},
		{"wrong username", true, "nobody", "secret", http.StatusUnauthorized},
		{"wrong password", true, "admin", "wrong", http.StatusUnauthorized},
		{"password prefix", true, "admin", "secre", http.StatusUnauthorized},
		{"password with suffix", true, "admin", "secret2", http.StatusUnauthorized},
		{"empty credentials", true, "", "", http.StatusUnauthorized},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(http.MethodGet, ts.URL+"/v2/catalog", nil)
			require.NoError(t, err)
			req.Header.Set(osbapi.HeaderAPIVersion, osbapi.APIVersion)
			if tc.setAuth {
				req.SetBasicAuth(tc.username, tc.password)
			}

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tc.want, resp.StatusCode)
			if tc.want == http.StatusUnauthorized {
				assert.Equal(t, `Basic realm="Service Broker"`, resp.Header.Get("WWW-Authenticate"))
			}
		})
	}
}
