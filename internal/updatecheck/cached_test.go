package updatecheck

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestCheckCached(t *testing.T) {
	t.Run("fresh cache hit does not call network", func(t *testing.T) {
		var calls atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			calls.Add(1)
			_, _ = io.WriteString(w, `{"tag_name":"v0.2.0"}`)
		}))
		defer srv.Close()

		dir := t.TempDir()
		// Seed a fresh cache (checked just now).
		fresh := cacheEntry{CheckedAt: time.Now(), Latest: "v0.3.0", CurrentVersion: "v0.1.0"}
		if err := writeCache(dir, fresh); err != nil {
			t.Fatal(err)
		}

		latest, available, err := checkCachedWithBaseURL(context.Background(), "v0.1.0", dir, 24*time.Hour, srv.URL)
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}
		if latest != "v0.3.0" || !available {
			t.Errorf("got (%q, %v), want (v0.3.0, true)", latest, available)
		}
		if calls.Load() != 0 {
			t.Errorf("network called %d times, want 0", calls.Load())
		}
	})

	t.Run("stale cache triggers network and rewrites cache", func(t *testing.T) {
		var calls atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			calls.Add(1)
			_, _ = io.WriteString(w, `{"tag_name":"v0.2.0"}`)
		}))
		defer srv.Close()

		dir := t.TempDir()
		// Seed a stale cache (checked 2 days ago).
		stale := cacheEntry{CheckedAt: time.Now().Add(-48 * time.Hour), Latest: "v0.1.5", CurrentVersion: "v0.1.0"}
		if err := writeCache(dir, stale); err != nil {
			t.Fatal(err)
		}

		latest, available, err := checkCachedWithBaseURL(context.Background(), "v0.1.0", dir, 24*time.Hour, srv.URL)
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}
		if latest != "v0.2.0" || !available {
			t.Errorf("got (%q, %v), want (v0.2.0, true)", latest, available)
		}
		if calls.Load() != 1 {
			t.Errorf("network called %d times, want 1", calls.Load())
		}

		// The refreshed cache should reflect v0.2.0.
		got, err := readCache(dir)
		if err != nil {
			t.Fatalf("readCache: %v", err)
		}
		if got.Latest != "v0.2.0" {
			t.Errorf("cached latest = %q, want v0.2.0", got.Latest)
		}
	})

	t.Run("cache for a different currentVersion is ignored", func(t *testing.T) {
		var calls atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			calls.Add(1)
			_, _ = io.WriteString(w, `{"tag_name":"v0.2.0"}`)
		}))
		defer srv.Close()

		dir := t.TempDir()
		// Cache is fresh but for a different version of the binary.
		wrongVersion := cacheEntry{CheckedAt: time.Now(), Latest: "v0.2.0", CurrentVersion: "v0.0.9"}
		if err := writeCache(dir, wrongVersion); err != nil {
			t.Fatal(err)
		}

		_, _, err := checkCachedWithBaseURL(context.Background(), "v0.1.0", dir, 24*time.Hour, srv.URL)
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}
		if calls.Load() != 1 {
			t.Errorf("network called %d times, want 1", calls.Load())
		}
	})

	t.Run("missing cache triggers network", func(t *testing.T) {
		var calls atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			calls.Add(1)
			_, _ = io.WriteString(w, `{"tag_name":"v0.2.0"}`)
		}))
		defer srv.Close()

		dir := t.TempDir()
		_, _, err := checkCachedWithBaseURL(context.Background(), "v0.1.0", dir, 24*time.Hour, srv.URL)
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}
		if calls.Load() != 1 {
			t.Errorf("network called %d times, want 1", calls.Load())
		}
	})
}
