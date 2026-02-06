package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/salmonumbrella/instantly-cli/internal/outfmt"
)

func TestAPIMethodCmd_EdgeCases(t *testing.T) {
	// empty path after trim
	res := execCLI(t, "--dry-run", "--output", "json", "api", "get", " ")
	if res.Err == nil {
		t.Fatalf("expected error")
	}

	// empty --query pair should be ignored
	res = execCLI(t, "--dry-run", "--output", "json", "api", "get", "/x", "--query", "")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}
}

func TestAPIMethodCmd_UnsupportedMethod_DefaultBranch(t *testing.T) {
	resetFlagsToDefaults()
	flags.DryRun = true

	c := newAPIMethodCmd("nope")
	var out bytes.Buffer
	c.SetOut(&out)
	c.SetErr(&out)
	c.SetContext(outfmt.WithMode(context.Background(), outfmt.JSON))
	c.SetArgs([]string{"/x"})
	if err := c.Execute(); err == nil {
		t.Fatalf("expected error")
	}
}
