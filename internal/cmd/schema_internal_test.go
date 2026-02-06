package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestSubcommandsSchema_SkipsHiddenHelpCompletion(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	root.AddCommand(&cobra.Command{Use: "help"})
	root.AddCommand(&cobra.Command{Use: "completion"})
	root.AddCommand(&cobra.Command{Use: "hidden", Hidden: true, Run: func(*cobra.Command, []string) {}})
	root.AddCommand(&cobra.Command{Use: "ok", Run: func(*cobra.Command, []string) {}})

	out := subcommandsSchema(root)
	if len(out) != 1 || out[0].Use != "ok" {
		t.Fatalf("out=%#v", out)
	}
}

func TestCommandSchema_HTTPConfirmAndPayloadSignals(t *testing.T) {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete thing (DELETE /things/{id}, requires --confirm)",
	}
	cmd.Flags().Bool("confirm", false, "Confirm")
	cmd.Flags().String("data-json", "", "Body JSON")
	cmd.Flags().String("body-file", "", "Body file")

	out := commandSchema(cmd)
	if out.HTTPMethod != "DELETE" || out.Endpoint != "/things/{id}" || !out.IsWrite {
		t.Fatalf("http parse failed: %#v", out)
	}
	if !out.HasConfirm || !out.NeedsConfirm {
		t.Fatalf("confirm signals failed: %#v", out)
	}
	if len(out.PayloadFlags) != 2 || out.PayloadFlags[0] != "data-json" || out.PayloadFlags[1] != "body-file" {
		t.Fatalf("payload flags failed: %#v", out.PayloadFlags)
	}
}

func TestCommandSchema_NoHints(t *testing.T) {
	cmd := &cobra.Command{
		Use:   "noop",
		Short: "No method hint here",
	}
	cmd.Flags().Bool("confirm", false, "Confirm")

	out := commandSchema(cmd)
	if out.HTTPMethod != "" || out.Endpoint != "" || out.IsWrite {
		t.Fatalf("expected no http hints: %#v", out)
	}
	if !out.HasConfirm || out.NeedsConfirm {
		t.Fatalf("expected confirm flag but not required: %#v", out)
	}
	if out.PayloadFlags != nil {
		t.Fatalf("expected no payload flags: %#v", out.PayloadFlags)
	}
}

func TestParseMethodAndEndpoint(t *testing.T) {
	m, p, ok := parseMethodAndEndpoint("List leads (POST /leads/list)")
	if !ok || m != "POST" || p != "/leads/list" {
		t.Fatalf("got m=%q p=%q ok=%v", m, p, ok)
	}
	_, _, ok = parseMethodAndEndpoint("nope")
	if ok {
		t.Fatalf("expected ok=false")
	}
}

func TestPayloadFlagsFrom_Nil(t *testing.T) {
	if got := payloadFlagsFrom(nil); got != nil {
		t.Fatalf("expected nil, got %#v", got)
	}
}
