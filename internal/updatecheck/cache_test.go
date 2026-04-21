package updatecheck

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReadWriteCache(t *testing.T) {
	t.Run("write then read returns the same entry", func(t *testing.T) {
		dir := t.TempDir()
		want := cacheEntry{
			CheckedAt:      time.Now().UTC().Truncate(time.Second),
			Latest:         "v0.2.0",
			CurrentVersion: "v0.1.0",
		}
		if err := writeCache(dir, want); err != nil {
			t.Fatalf("writeCache: %v", err)
		}
		got, err := readCache(dir)
		if err != nil {
			t.Fatalf("readCache: %v", err)
		}
		if !got.CheckedAt.Equal(want.CheckedAt) || got.Latest != want.Latest || got.CurrentVersion != want.CurrentVersion {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})

	t.Run("read returns error when file missing", func(t *testing.T) {
		dir := t.TempDir()
		_, err := readCache(dir)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("read returns error when file corrupt", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "update.json"), []byte("not json"), 0o600); err != nil {
			t.Fatal(err)
		}
		_, err := readCache(dir)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("writeCache creates the cache dir if missing", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "nested", "spot")
		entry := cacheEntry{CheckedAt: time.Now(), Latest: "v1.0.0", CurrentVersion: "v0.9.0"}
		if err := writeCache(dir, entry); err != nil {
			t.Fatalf("writeCache: %v", err)
		}
		if _, err := os.Stat(filepath.Join(dir, "update.json")); err != nil {
			t.Errorf("cache file missing: %v", err)
		}
	})
}
