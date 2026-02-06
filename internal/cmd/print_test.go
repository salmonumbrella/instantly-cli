package cmd

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/instantly-cli/internal/api"
	"github.com/salmonumbrella/instantly-cli/internal/outfmt"
)

func TestMetaFrom(t *testing.T) {
	if got := metaFrom(nil, map[string]any{"items": []any{}}); got != nil {
		t.Fatalf("got=%#v", got)
	}

	meta := &api.Meta{}
	meta.Request.URL = "http://x"
	resp := map[string]any{"items": []any{}, "next_starting_after": "c1"}
	got := metaFrom(meta, resp)
	if got["request_url"] != "http://x" {
		t.Fatalf("got=%#v", got)
	}
	p := got["pagination"].(map[string]any)
	if p["next_starting_after"] != "c1" {
		t.Fatalf("p=%#v", p)
	}

	// rate limit only
	meta = &api.Meta{}
	meta.RateLimit = &api.RateLimitInfo{}
	got = metaFrom(meta, map[string]any{"ok": true})
	if _, ok := got["rate_limit"]; !ok {
		t.Fatalf("got=%#v", got)
	}

	// meta empty + no pagination => nil
	meta = &api.Meta{}
	got = metaFrom(meta, map[string]any{"ok": true})
	if got != nil {
		t.Fatalf("got=%#v", got)
	}
}

func TestPrintResult_ModesAndFiltering(t *testing.T) {
	resetFlagsToDefaults()

	makeCmd := func(mode outfmt.Mode) (*cobra.Command, *bytes.Buffer) {
		c := &cobra.Command{}
		var out bytes.Buffer
		var errBuf bytes.Buffer
		c.SetOut(&out)
		c.SetErr(&errBuf)
		c.SetContext(outfmt.WithMode(context.Background(), mode))
		return c, &out
	}

	// text
	{
		c, out := makeCmd(outfmt.Text)
		if err := printResult(c, "k", map[string]any{"a": 1}, nil); err != nil {
			t.Fatalf("err=%v", err)
		}
		if strings.TrimSpace(out.String()) != "ok" {
			t.Fatalf("out=%q", out.String())
		}
	}

	// json
	{
		c, out := makeCmd(outfmt.JSON)
		if err := printResult(c, "k", map[string]any{"a": 1}, nil); err != nil {
			t.Fatalf("err=%v", err)
		}
		if !strings.Contains(out.String(), "\"a\": 1") {
			t.Fatalf("out=%q", out.String())
		}
	}

	// agent
	{
		c, out := makeCmd(outfmt.Agent)
		if err := printResult(c, "k", map[string]any{"a": 1}, map[string]any{"m": 1}); err != nil {
			t.Fatalf("err=%v", err)
		}
		if !strings.Contains(out.String(), "\"kind\": \"k\"") {
			t.Fatalf("out=%q", out.String())
		}
	}

	// json + jq
	{
		resetFlagsToDefaults()
		flags.JQ = ".a"
		c, out := makeCmd(outfmt.JSON)
		if err := printResult(c, "k", map[string]any{"a": 1}, nil); err != nil {
			t.Fatalf("err=%v", err)
		}
		if strings.TrimSpace(out.String()) != "1" {
			t.Fatalf("out=%q", out.String())
		}
	}

	// jsonl
	{
		resetFlagsToDefaults()
		c, out := makeCmd(outfmt.JSONL)
		if err := printResult(c, "k", map[string]any{"a": 1}, nil); err != nil {
			t.Fatalf("err=%v", err)
		}
		if strings.Contains(out.String(), "\n  ") {
			t.Fatalf("expected jsonl, got %q", out.String())
		}
	}

	// jsonl + jq (covers jsonl filtering branch)
	{
		resetFlagsToDefaults()
		flags.JQ = ".a"
		c, out := makeCmd(outfmt.JSONL)
		if err := printResult(c, "k", map[string]any{"a": 1}, nil); err != nil {
			t.Fatalf("err=%v", err)
		}
		if strings.TrimSpace(out.String()) != "1" {
			t.Fatalf("out=%q", out.String())
		}
	}

	// jsonl + invalid jq should error
	{
		resetFlagsToDefaults()
		flags.JQ = "???"
		c, out := makeCmd(outfmt.JSONL)
		err := printResult(c, "k", map[string]any{"a": 1}, nil)
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(out.String(), "\"error\"") {
			t.Fatalf("out=%q", out.String())
		}
	}

	// agent + fields (exercises fields query path)
	{
		resetFlagsToDefaults()
		flags.Fields = "request.method"
		c, out := makeCmd(outfmt.Agent)
		resp := map[string]any{"items": []any{map[string]any{"request": map[string]any{"method": "GET"}}}}
		if err := printResult(c, "k", resp, nil); err != nil {
			t.Fatalf("err=%v", err)
		}
		if !strings.Contains(out.String(), "\"items\"") {
			t.Fatalf("out=%q", out.String())
		}
	}

	// jq/fields conflict should surface as error
	{
		resetFlagsToDefaults()
		flags.JQ = "."
		flags.Fields = "a"
		c, out := makeCmd(outfmt.JSON)
		err := printResult(c, "k", map[string]any{"a": 1}, nil)
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(out.String(), "\"error\"") {
			t.Fatalf("out=%q", out.String())
		}
	}

	// agent + invalid jq should hit error path inside agent branch
	{
		resetFlagsToDefaults()
		flags.JQ = "???"
		c, out := makeCmd(outfmt.Agent)
		err := printResult(c, "k", map[string]any{"a": 1}, nil)
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(out.String(), "\"error\"") {
			t.Fatalf("out=%q", out.String())
		}
	}

	// invalid jq should print structured error (json mode)
	{
		resetFlagsToDefaults()
		flags.JQ = "???"
		c, out := makeCmd(outfmt.JSON)
		err := printResult(c, "k", map[string]any{"a": 1}, nil)
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(out.String(), "\"error\"") {
			t.Fatalf("out=%q", out.String())
		}
	}

	// normalizeForJQ failure (unmarshalable type) should emit structured error.
	{
		resetFlagsToDefaults()
		flags.JQ = "."
		c, out := makeCmd(outfmt.JSON)
		err := printResult(c, "k", map[string]any{"f": func() {}}, nil)
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(out.String(), "\"error\"") {
			t.Fatalf("out=%q", out.String())
		}
	}

	// normalizeForJQ unmarshal failure path via injection
	{
		old := jsonUnmarshal
		t.Cleanup(func() { jsonUnmarshal = old })
		jsonUnmarshal = func(_ []byte, _ any) error { return errors.New("boom") }

		resetFlagsToDefaults()
		flags.JQ = "."
		c, out := makeCmd(outfmt.JSON)
		err := printResult(c, "k", map[string]any{"a": 1}, nil)
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(out.String(), "\"error\"") {
			t.Fatalf("out=%q", out.String())
		}
	}
}

func TestPrintError_Modes(t *testing.T) {
	makeCmd := func(mode outfmt.Mode) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
		c := &cobra.Command{}
		var out bytes.Buffer
		var errBuf bytes.Buffer
		c.SetOut(&out)
		c.SetErr(&errBuf)
		c.SetContext(outfmt.WithMode(context.Background(), mode))
		return c, &out, &errBuf
	}

	// text goes to stderr
	{
		c, out, errBuf := makeCmd(outfmt.Text)
		e := errors.New("nope")
		_ = printError(c, "k", e, nil)
		if out.Len() != 0 {
			t.Fatalf("stdout=%q", out.String())
		}
		if !strings.Contains(errBuf.String(), "nope") {
			t.Fatalf("stderr=%q", errBuf.String())
		}
	}

	// jsonl goes to stdout (one line)
	{
		c, out, _ := makeCmd(outfmt.JSONL)
		e := errors.New("nope")
		_ = printError(c, "k", e, map[string]any{"m": 1})
		if !strings.Contains(out.String(), "\"error\"") {
			t.Fatalf("stdout=%q", out.String())
		}
	}
}
