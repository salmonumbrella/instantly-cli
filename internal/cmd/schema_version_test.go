package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

func TestSchemaCommand(t *testing.T) {
	res := execCLI(t, "schema", "--output", "json")
	if res.Err != nil {
		t.Fatalf("err=%v stdout=%q", res.Err, string(res.Stdout))
	}
	v := mustJSON(t, res.Stdout).(map[string]any)
	cmds := v["commands"].([]any)
	joined := strings.ToLower(string(res.Stdout))
	if !strings.Contains(joined, "instantly accounts") {
		t.Fatalf("schema missing accounts: %q", string(res.Stdout))
	}
	if strings.Contains(joined, "instantly completion") {
		t.Fatalf("schema should omit completion: %q", string(res.Stdout))
	}
	if len(cmds) == 0 {
		t.Fatalf("expected commands")
	}
}

func TestVersionCommand(t *testing.T) {
	res := execCLI(t, "version", "--output", "json")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}
	v := mustJSON(t, res.Stdout).(map[string]any)
	if _, ok := v["version"].(string); !ok {
		t.Fatalf("v=%#v", v)
	}

	res = execCLI(t, "version", "--output", "text")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}
	if strings.TrimSpace(string(res.Stdout)) == "" {
		t.Fatalf("expected text output")
	}
}

func TestFlagsFromFlagSet_NilAndEmpty(t *testing.T) {
	if got := flagsFromFlagSet(nil); got != nil {
		t.Fatalf("got=%#v", got)
	}
	fs := pflag.NewFlagSet("x", pflag.ContinueOnError)
	if got := flagsFromFlagSet(fs); got != nil {
		t.Fatalf("got=%#v", got)
	}

	_ = fs.StringP("name", "n", "d", "usage")
	got := flagsFromFlagSet(fs)
	if len(got) != 1 || got[0].Name != "name" || got[0].Shorthand != "n" {
		t.Fatalf("got=%#v", got)
	}
}
