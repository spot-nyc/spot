package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spot-nyc/spot"
	"github.com/spot-nyc/spot/auth"
)

// integrationHarness configures the CLI for a full end-to-end test against
// an httptest-mocked morty. It sets SPOT_TOKEN and SPOT_BASE_URL via t.Setenv
// so the default token source + default client both resolve to the fake
// server. The returned command is a fresh root tree with stdout and stderr
// captured in the provided buffers.
func integrationHarness(t *testing.T, serverURL, token string, stdout, stderr *bytes.Buffer) *cobra.Command {
	t.Helper()

	// Env must be set BEFORE newRootCmd wires up DefaultTokenSource/etc.
	t.Setenv(auth.EnvTokenVar, token)
	t.Setenv(SpotBaseURLEnv, serverURL)

	cmd := newRootCmd()
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	return cmd
}

func TestCLI_SearchesList_JSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/searches/active", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"searches": [
				{
					"id": "srch_abcdef1234",
					"userId": "u1",
					"party": 2,
					"startDate": "2026-05-01",
					"endDate": "2026-05-01",
					"startTime": "18:00:00",
					"endTime": "21:00:00",
					"upgrade": false,
					"searchTargets": [
						{"id": "t1", "rank": 0, "restaurant": {"id": "r1", "name": "Gramercy Tavern"}}
					]
				}
			]
		}`)
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	cmd := integrationHarness(t, srv.URL, "test-token", &stdout, &stderr)
	cmd.SetArgs([]string{"searches", "list", "--json"})

	require.NoError(t, cmd.Execute())

	var got []spot.Search
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &got))
	require.Len(t, got, 1)
	assert.Equal(t, "srch_abcdef1234", got[0].ID)
	assert.Equal(t, 2, got[0].Party)
	require.Len(t, got[0].SearchTargets, 1)
	assert.Equal(t, "Gramercy Tavern", got[0].SearchTargets[0].Restaurant.Name)
}

func TestCLI_SearchesList_AutoDetectsJSONOnNonTTY(t *testing.T) {
	// When stdout is a buffer (not a TTY), resolveFormat should pick JSON
	// even without --json. This is the pipe-redirect path.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"searches":[]}`)
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	cmd := integrationHarness(t, srv.URL, "test-token", &stdout, &stderr)
	cmd.SetArgs([]string{"searches", "list"})

	require.NoError(t, cmd.Execute())

	var got []spot.Search
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &got))
	assert.Empty(t, got)
}

func TestCLI_AuthWhoami_JSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/users/me", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"user":{"id":"u1","phone":"+15555550123","name":"Brian"}}`)
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	cmd := integrationHarness(t, srv.URL, "test-token", &stdout, &stderr)
	cmd.SetArgs([]string{"auth", "whoami", "--json"})

	require.NoError(t, cmd.Execute())

	var got spot.User
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &got))
	assert.Equal(t, "u1", got.ID)
	assert.Equal(t, "Brian", got.Name)
}

func TestCLI_AuthWhoami_UnauthenticatedReturnsTypedError(t *testing.T) {
	// With SPOT_TOKEN unset (empty), DefaultTokenSource should return
	// ErrNoCredentials, which surfaces as ErrUnauthenticated to the caller
	// via our token check in Client.do.
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("HTTP should not be reached when credentials are missing")
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	cmd := integrationHarness(t, srv.URL, "", &stdout, &stderr)
	cmd.SetArgs([]string{"auth", "whoami"})

	err := cmd.Execute()
	require.Error(t, err)

	// The error should map to exit code 3 (auth required).
	assert.Equal(t, 3, ExitCodeFor(err))
}

func TestCLI_SearchesList_MapsHTTP401ToExitCode3(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"error":"Invalid or expired token"}`)
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	cmd := integrationHarness(t, srv.URL, "stale-token", &stdout, &stderr)
	cmd.SetArgs([]string{"searches", "list"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Equal(t, 3, ExitCodeFor(err))

	// Error message includes the server-provided text.
	assert.Contains(t, err.Error(), "Invalid or expired token")
}

// Tiny sanity check that --json output is not contaminated by any other
// stdout writes (e.g. stray fmt.Println in command bodies).
func TestCLI_SearchesList_JSONOutputIsPureJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"searches":[]}`)
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	cmd := integrationHarness(t, srv.URL, "test-token", &stdout, &stderr)
	cmd.SetArgs([]string{"searches", "list", "--json"})

	require.NoError(t, cmd.Execute())

	// Pure JSON means the buffer starts with '[' or '{' and ends with '\n'.
	got := strings.TrimRight(stdout.String(), "\n")
	assert.True(t, strings.HasPrefix(got, "["), "expected JSON array, got: %q", got)
}
