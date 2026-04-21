package updatecheck

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// cacheFileName is the name of the cache file inside the cache directory
// (e.g. inside $XDG_CACHE_HOME/spot/).
const cacheFileName = "update.json"

// cacheEntry is the on-disk shape of the update-check cache. CurrentVersion
// is recorded so the cache is invalidated when the installed binary is
// upgraded — otherwise a user who just upgraded could see a stale "update
// available" message for the old version.
type cacheEntry struct {
	CheckedAt      time.Time `json:"checkedAt"`
	Latest         string    `json:"latest"`
	CurrentVersion string    `json:"currentVersion"`
}

// readCache loads a cache entry from dir. Returns an error when the file is
// missing, unreadable, or not valid JSON; callers should treat any error as
// a cache miss.
func readCache(dir string) (cacheEntry, error) {
	raw, err := os.ReadFile(filepath.Join(dir, cacheFileName))
	if err != nil {
		return cacheEntry{}, err
	}
	var entry cacheEntry
	if err := json.Unmarshal(raw, &entry); err != nil {
		return cacheEntry{}, err
	}
	return entry, nil
}

// writeCache persists entry to dir/update.json. Writes to a temp file in the
// same directory first then renames, so a crashed write never leaves a
// partial file behind.
func writeCache(dir string, entry cacheEntry) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	raw, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, "update.json.*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(raw); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return os.Rename(tmpPath, filepath.Join(dir, cacheFileName))
}
