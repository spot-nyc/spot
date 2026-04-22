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
	"github.com/zalando/go-keyring"

	"github.com/spot-nyc/spot"
	"github.com/spot-nyc/spot/auth"
)

// integrationHarness configures the CLI for a full end-to-end test against
// an httptest-mocked Spot API. It sets SPOT_TOKEN and SPOT_BASE_URL via
// t.Setenv so the default token source + default client both resolve to
// the fake server.
//
// Credential-store isolation: the harness swaps the OS keyring to an
// in-memory mock and points XDG_CONFIG_HOME at a per-test tempdir. Tests
// that exercise DefaultStore (via `auth logout` or similar) cannot touch
// the developer's real credentials. Without this, a test run would wipe
// keychain entries on every invocation.
//
// The returned command is a fresh root tree with stdout and stderr
// captured in the provided buffers.
func integrationHarness(t *testing.T, serverURL, token string, stdout, stderr *bytes.Buffer) *cobra.Command {
	t.Helper()

	// Isolate credential storage for this test.
	keyring.MockInit()
	t.Cleanup(func() { keyring.MockInit() })
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

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
					"searchTargets": [
						{"restaurant": {"id": "r1", "name": "Gramercy Tavern"}}
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

func TestCLI_SearchesGet_JSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/searches/srch_abc", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"search":{"id":"srch_abc","party":2,"startDate":"2026-05-01","endDate":"2026-05-01","startTime":"18:00:00","endTime":"21:00:00"}}`)
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	cmd := integrationHarness(t, srv.URL, "test-token", &stdout, &stderr)
	cmd.SetArgs([]string{"searches", "get", "srch_abc", "--json"})

	require.NoError(t, cmd.Execute())

	var got spot.Search
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &got))
	assert.Equal(t, "srch_abc", got.ID)
}

func TestCLI_SearchesDelete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Soft-delete is POST /searches/:id with a deletedAt — the Spot API has no DELETE.
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/searches/srch_abc", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		deletedAt, ok := got["deletedAt"].(string)
		require.True(t, ok, "delete must POST a deletedAt timestamp")
		assert.NotEmpty(t, deletedAt)

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"search":{"id":"srch_abc"}}`)
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	cmd := integrationHarness(t, srv.URL, "test-token", &stdout, &stderr)
	cmd.SetArgs([]string{"searches", "delete", "srch_abc", "--json"})

	require.NoError(t, cmd.Execute())

	var got map[string]any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &got))
	assert.Equal(t, true, got["deleted"])
	assert.Equal(t, "srch_abc", got["id"])
}

func TestCLI_SearchesCreate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/searches", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		assert.EqualValues(t, 2, got["party"])
		assert.Equal(t, "18:00:00", got["startTime"], "time flag should be expanded to HH:MM:SS")
		assert.Equal(t, "21:00:00", got["endTime"])
		restaurantIDs, ok := got["restaurantIds"].([]any)
		require.True(t, ok, "the Spot API requires `restaurantIds` on create (was `restaurants` — regression guard)")
		assert.Len(t, restaurantIDs, 2)

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"search":{"id":"srch_new","party":2}}`)
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	cmd := integrationHarness(t, srv.URL, "test-token", &stdout, &stderr)
	cmd.SetArgs([]string{
		"searches", "create",
		"--party", "2",
		"--date", "2026-05-01",
		"--start-time", "18:00",
		"--end-time", "21:00",
		"--restaurant", "rst_abc",
		"--restaurant", "rst_def",
		"--json",
	})

	require.NoError(t, cmd.Execute())

	var got spot.Search
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &got))
	assert.Equal(t, "srch_new", got.ID)
}

func TestCLI_ReservationsList_JSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/searches/bookings", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"reservations":[{"id":"rsv_1","table":{"platform":"resy","date":"2026-05-01","time":"19:00:00","party":2,"seating":"Dining Room","restaurant":{"id":"r1","name":"Gramercy Tavern"}}}]}`)
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	cmd := integrationHarness(t, srv.URL, "test-token", &stdout, &stderr)
	cmd.SetArgs([]string{"reservations", "list", "--json"})

	require.NoError(t, cmd.Execute())

	var got []spot.Reservation
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &got))
	require.Len(t, got, 1)
	assert.Equal(t, "rsv_1", got[0].ID)
	assert.Equal(t, "Gramercy Tavern", got[0].Table.Restaurant.Name)
	assert.Equal(t, "Dining Room", got[0].Table.Seating)
	assert.Equal(t, "resy", got[0].Table.Platform)
}

func TestCLI_ReservationsCancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/reservations/rsv_abc/cancel", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	cmd := integrationHarness(t, srv.URL, "test-token", &stdout, &stderr)
	cmd.SetArgs([]string{"reservations", "cancel", "rsv_abc", "--json"})

	require.NoError(t, cmd.Execute())

	var got map[string]any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &got))
	assert.Equal(t, true, got["cancelled"])
	assert.Equal(t, "rsv_abc", got["id"])
}

func TestCLI_RestaurantsSearch_JSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/restaurants/search", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, "gramercy", got["query"])

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"restaurants":[{"id":"rst_abc","name":"Gramercy Tavern","neighborhood":"Flatiron","cuisine":"American","zone":"NYC","resyActive":true,"openTableActive":false,"sevenRoomsActive":false,"doorDashActive":false}]}`)
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	cmd := integrationHarness(t, srv.URL, "test-token", &stdout, &stderr)
	cmd.SetArgs([]string{"restaurants", "search", "gramercy", "--json"})

	require.NoError(t, cmd.Execute())

	var got []spot.Restaurant
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &got))
	require.Len(t, got, 1)
	assert.Equal(t, "rst_abc", got[0].ID)
	assert.Equal(t, "Flatiron", got[0].Neighborhood)
	assert.Equal(t, "American", got[0].Cuisine)
	assert.True(t, got[0].ResyActive)
	assert.Equal(t, []string{"Resy"}, got[0].Platforms())
}

func TestCLI_ReservationsSearch_JSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/reservations/search", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		ids, ok := got["restaurantIds"].([]any)
		require.True(t, ok)
		assert.Len(t, ids, 2)
		assert.Equal(t, "18:00:00", got["startTime"])
		assert.Equal(t, "21:00:00", got["endTime"])
		assert.Equal(t, "2026-05-15", got["date"])
		assert.EqualValues(t, 2, got["party"])

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"availability":[{"restaurant":{"id":"rst_a","name":"Gramercy Tavern","resyActive":true},"slots":[{"id":"slot_1","platform":"resy","date":"2026-05-15","time":"19:00:00","party":2,"seating":"default","restaurantId":"rst_a"}]}]}`)
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	cmd := integrationHarness(t, srv.URL, "test-token", &stdout, &stderr)
	cmd.SetArgs([]string{
		"reservations", "search",
		"--restaurant", "rst_a,rst_b",
		"--date", "2026-05-15",
		"--start-time", "18:00",
		"--end-time", "21:00",
		"--party", "2",
		"--json",
	})

	require.NoError(t, cmd.Execute())

	var got []spot.ReservationSlot
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &got))
	require.Len(t, got, 1)
	assert.Equal(t, "slot_1", got[0].ID)
	assert.Equal(t, "resy", got[0].Platform)
	require.NotNil(t, got[0].Restaurant)
	assert.Equal(t, "Gramercy Tavern", got[0].Restaurant.Name)
}

func TestCLI_SearchesUpdate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/searches/srch_abc", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))

		assert.EqualValues(t, 4, got["party"])
		_, hasStartDate := got["startDate"]
		assert.False(t, hasStartDate, "startDate should be omitted when --date is not set")
		_, hasRestaurants := got["restaurantIds"]
		assert.False(t, hasRestaurants, "restaurantIds should be omitted when --restaurant is not set")

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"search":{"id":"srch_abc","party":4}}`)
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	cmd := integrationHarness(t, srv.URL, "test-token", &stdout, &stderr)
	cmd.SetArgs([]string{"searches", "update", "srch_abc", "--party", "4", "--json"})

	require.NoError(t, cmd.Execute())

	var updated spot.Search
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &updated))
	assert.Equal(t, "srch_abc", updated.ID)
	assert.Equal(t, 4, updated.Party)
}

func TestCLI_SearchesUpdate_NothingToUpdate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("HTTP should not be reached when no flags are provided")
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	cmd := integrationHarness(t, srv.URL, "test-token", &stdout, &stderr)
	cmd.SetArgs([]string{"searches", "update", "srch_abc"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nothing to update")
}

func TestCLI_ReservationsBook_JSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/reservations/book", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, "slot_abc", got["slotId"])

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"reservation":{"id":"rsv_abc","table":{"id":"tbl_abc","platform":"resy","date":"2026-05-15","time":"19:00:00","party":2,"seating":"Dining Room","restaurant":{"id":"rst_a","name":"Gramercy Tavern"}}}}`)
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	cmd := integrationHarness(t, srv.URL, "test-token", &stdout, &stderr)
	cmd.SetArgs([]string{"reservations", "book", "slot_abc", "--json"})

	require.NoError(t, cmd.Execute())

	var got spot.Reservation
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &got))
	assert.Equal(t, "rsv_abc", got.ID)
}

func TestCLI_ReservationsBook_PlatformNotConnected_Exit10(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusPreconditionFailed)
		_, _ = io.WriteString(w, `{"error":"platform not connected","platform":"resy"}`)
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	cmd := integrationHarness(t, srv.URL, "test-token", &stdout, &stderr)
	cmd.SetArgs([]string{"reservations", "book", "slot_abc"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Equal(t, 10, ExitCodeFor(err))
}

func TestCLI_ReservationsBook_SlotExpired_Exit11(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusGone)
		_, _ = io.WriteString(w, `{"error":"slot is no longer available"}`)
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	cmd := integrationHarness(t, srv.URL, "test-token", &stdout, &stderr)
	cmd.SetArgs([]string{"reservations", "book", "slot_expired"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Equal(t, 11, ExitCodeFor(err))
}

func TestCLI_RestaurantsGet_JSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/restaurants/rst_abc", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"restaurant":{"id":"rst_abc","name":"Gramercy Tavern","neighborhood":"Flatiron","cuisine":"American","phone":"212-477-0777","website":"https://www.gramercytavern.com","minimumPartySize":1,"maximumPartySize":8,"bookingDifficulty":8,"resyActive":true,"openTableActive":false,"sevenRoomsActive":false,"doorDashActive":false}}`)
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	cmd := integrationHarness(t, srv.URL, "test-token", &stdout, &stderr)
	cmd.SetArgs([]string{"restaurants", "get", "rst_abc", "--json"})

	require.NoError(t, cmd.Execute())

	var got spot.Restaurant
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &got))
	assert.Equal(t, "rst_abc", got.ID)
	assert.Equal(t, "Gramercy Tavern", got.Name)
	assert.Equal(t, "Flatiron", got.Neighborhood)
	assert.Equal(t, []string{"Resy"}, got.Platforms())
}

func TestCLI_ReservationsHistory_JSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/reservations", r.URL.Path)
		assert.Equal(t, "true", r.URL.Query().Get("external"))

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"reservations": [
				{"id":"rsv_1","table":{"platform":"resy","date":"2026-04-20","time":"19:00:00","party":2,"seating":"default","restaurant":{"id":"r1","name":"Gramercy Tavern"}}},
				{"id":"rsv_2","table":{"platform":"opentable","date":"2026-03-15","time":"20:30:00","party":4,"seating":"Bar","restaurant":{"id":"r2","name":"Shuko"}}}
			]
		}`)
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	cmd := integrationHarness(t, srv.URL, "test-token", &stdout, &stderr)
	cmd.SetArgs([]string{"reservations", "history", "--json"})

	require.NoError(t, cmd.Execute())

	var got []spot.Reservation
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &got))
	require.Len(t, got, 2)
	assert.Equal(t, "rsv_1", got[0].ID)
	assert.Equal(t, "Gramercy Tavern", got[0].Table.Restaurant.Name)
	assert.Equal(t, "opentable", got[1].Table.Platform)
}

func TestCLI_AuthLogout_CallsServerRevocation(t *testing.T) {
	revocationCalled := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/users/me/logout" {
			revocationCalled = true
			assert.Equal(t, http.MethodPost, r.Method)
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			var got map[string]any
			require.NoError(t, json.Unmarshal(body, &got))
			_, hasScope := got["scope"]
			assert.False(t, hasScope, "default logout omits scope")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	cmd := integrationHarness(t, srv.URL, "test-token", &stdout, &stderr)
	cmd.SetArgs([]string{"auth", "logout"})

	require.NoError(t, cmd.Execute())
	assert.True(t, revocationCalled, "logout must call the revocation endpoint")
}

func TestCLI_AuthLogout_AllFlag_SendsGlobalScope(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/users/me/logout" {
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			var got map[string]any
			require.NoError(t, json.Unmarshal(body, &got))
			assert.Equal(t, "global", got["scope"])
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	cmd := integrationHarness(t, srv.URL, "test-token", &stdout, &stderr)
	cmd.SetArgs([]string{"auth", "logout", "--all"})

	require.NoError(t, cmd.Execute())
}

func TestCLI_AuthLogout_ServerError_ClearsLocalAnyway(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"error":"boom"}`)
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	cmd := integrationHarness(t, srv.URL, "test-token", &stdout, &stderr)
	cmd.SetArgs([]string{"auth", "logout"})

	// Local cleanup runs even though server-side revocation failed.
	// (stdout is JSON because the harness buffer is non-TTY.)
	require.NoError(t, cmd.Execute())
	assert.Contains(t, stderr.String(), "warning: server-side revocation did not complete")
	assert.Contains(t, stdout.String(), `"signedOut": true`)
}
