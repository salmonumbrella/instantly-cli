package outfmt

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestParseMode(t *testing.T) {
	if m, err := ParseMode(""); err != nil || m != JSON {
		t.Fatalf("ParseMode(\"\") = %v, %v", m, err)
	}
	if m, err := ParseMode("agent"); err != nil || m != Agent {
		t.Fatalf("ParseMode(agent) = %v, %v", m, err)
	}
	if _, err := ParseMode("nope"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestModeContext(t *testing.T) {
	if got := ModeFrom(context.Background()); got != JSON {
		t.Fatalf("ModeFrom(background) = %v", got)
	}
	ctx := WithMode(context.Background(), Agent)
	if got := ModeFrom(ctx); got != Agent {
		t.Fatalf("ModeFrom(ctx) = %v", got)
	}
}

func TestPrintJSONAndJSONL(t *testing.T) {
	var b bytes.Buffer
	if err := PrintJSON(&b, map[string]any{"a": 1}); err != nil {
		t.Fatalf("PrintJSON err: %v", err)
	}
	s := b.String()
	if !strings.Contains(s, "\n  ") {
		t.Fatalf("expected indented json, got %q", s)
	}

	b.Reset()
	if err := PrintJSONL(&b, map[string]any{"a": 1}); err != nil {
		t.Fatalf("PrintJSONL err: %v", err)
	}
	s = b.String()
	if strings.Contains(s, "\n  ") {
		t.Fatalf("expected compact jsonl, got %q", s)
	}
}
