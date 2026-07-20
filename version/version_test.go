package version_test

import (
	"testing"

	"github.com/conallob/coding-interview-pattern-drill/version"
)

func TestDefaultVersionIsDev(t *testing.T) {
	if version.Version != "dev" {
		t.Errorf("Version = %q, want %q (default before ldflags override it)", version.Version, "dev")
	}
}
