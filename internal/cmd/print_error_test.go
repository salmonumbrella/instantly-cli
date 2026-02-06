package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/instantly-cli/internal/api"
	"github.com/salmonumbrella/instantly-cli/internal/outfmt"
)

func TestPrintError_IncludesHTTPStatusForAPIErrors(t *testing.T) {
	c := &cobra.Command{}
	var buf bytes.Buffer
	c.SetOut(&buf)
	c.SetErr(&buf)
	c.SetContext(outfmt.WithMode(context.Background(), outfmt.JSON))

	err := printError(c, "k", &api.APIError{Status: 500, Message: "boom"}, nil)
	if err == nil {
		t.Fatalf("expected error")
	}

	var got map[string]any
	if e := json.Unmarshal(buf.Bytes(), &got); e != nil {
		t.Fatalf("parse output json: %v; out=%q", e, buf.String())
	}
	meta, _ := got["meta"].(map[string]any)
	if meta == nil {
		t.Fatalf("expected meta in output; got=%v", got)
	}
	if meta["http_status"] != float64(500) {
		t.Fatalf("expected http_status=500; got=%v", meta["http_status"])
	}
}
