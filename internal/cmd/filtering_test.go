package cmd

import (
	"os"
	"strings"
	"testing"
)

func TestDefaultOutput(t *testing.T) {
	old := os.Getenv("INSTANTLY_OUTPUT")
	t.Cleanup(func() { _ = os.Setenv("INSTANTLY_OUTPUT", old) })

	_ = os.Setenv("INSTANTLY_OUTPUT", "agent")
	if got := defaultOutput(); got != "agent" {
		t.Fatalf("got=%q", got)
	}

	_ = os.Setenv("INSTANTLY_OUTPUT", "")
	if got := defaultOutput(); got != "agent" {
		t.Fatalf("got=%q", got)
	}
}

func TestEffectiveJQExpression(t *testing.T) {
	resetFlagsToDefaults()

	flags.JQ = "."
	flags.Fields = "a"
	if _, err := effectiveJQExpression(); err == nil {
		t.Fatalf("expected error")
	}

	flags.JQ = ".a"
	flags.Fields = ""
	if got, err := effectiveJQExpression(); err != nil || got != ".a" {
		t.Fatalf("got=%q err=%v", got, err)
	}

	flags.JQ = ""
	flags.Fields = ""
	if got, err := effectiveJQExpression(); err != nil || got != "" {
		t.Fatalf("got=%q err=%v", got, err)
	}

	flags.JQ = ""
	flags.Fields = "a,b,b"
	got, err := effectiveJQExpression()
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if !strings.Contains(got, "items") {
		t.Fatalf("got=%q", got)
	}

	flags.JQ = ""
	flags.Fields = "a.-b"
	if _, err := effectiveJQExpression(); err == nil {
		t.Fatalf("expected error")
	}
}

func TestParseFields(t *testing.T) {
	if _, err := parseFields(""); err == nil {
		t.Fatalf("expected error")
	}
	if _, err := parseFields("a,-b"); err == nil {
		t.Fatalf("expected error")
	}
	if _, err := parseFields(", ,"); err == nil {
		t.Fatalf("expected error")
	}
	out, err := parseFields("a, a.b ,a")
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if len(out) != 2 {
		t.Fatalf("out=%v", out)
	}
	out, err = parseFields("a,,b")
	if err != nil || len(out) != 2 {
		t.Fatalf("out=%v err=%v", out, err)
	}
}

func TestBuildObjectExpr(t *testing.T) {
	got := buildObjectExpr([]string{"a", "b.c"})
	if !strings.Contains(got, "a: .a") || !strings.Contains(got, "c: .b.c") {
		t.Fatalf("got=%q", got)
	}
}
