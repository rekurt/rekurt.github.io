package buildinfo

import "testing"

func TestVersionIsSemver(t *testing.T) {
	if Version != "0.1.0" {
		t.Fatalf("Version = %q, want 0.1.0", Version)
	}
}
