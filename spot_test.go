package spot

import "testing"

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Fatal("spot.Version is empty; expected a non-empty default")
	}
}
