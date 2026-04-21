package updatecheck

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCheck(t *testing.T) {
	t.Run("returns latest tag and newer=true when remote is ahead", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/repos/spot-nyc/spot/releases/latest" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"tag_name":"v0.2.0"}`)
		}))
		defer srv.Close()

		latest, available, err := checkWithBaseURL(context.Background(), "v0.1.0", srv.URL)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if latest != "v0.2.0" {
			t.Errorf("latest = %q, want v0.2.0", latest)
		}
		if !available {
			t.Error("available = false, want true")
		}
	})

	t.Run("returns available=false when remote matches current", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = io.WriteString(w, `{"tag_name":"v0.1.0"}`)
		}))
		defer srv.Close()

		_, available, err := checkWithBaseURL(context.Background(), "v0.1.0", srv.URL)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if available {
			t.Error("available = true, want false")
		}
	})

	t.Run("returns error when API responds non-2xx", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer srv.Close()

		_, _, err := checkWithBaseURL(context.Background(), "v0.1.0", srv.URL)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("honors context timeout", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			time.Sleep(200 * time.Millisecond)
			_, _ = io.WriteString(w, `{"tag_name":"v0.2.0"}`)
		}))
		defer srv.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		_, _, err := checkWithBaseURL(ctx, "v0.1.0", srv.URL)
		if err == nil {
			t.Error("expected timeout error, got nil")
		}
	})
}
