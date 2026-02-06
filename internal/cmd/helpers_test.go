package cmd

import (
	"bytes"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func TestApplyQueryPairs(t *testing.T) {
	q := url.Values{}
	if err := applyQueryPairs(q, []string{"", "a=1", "b=2"}); err != nil {
		t.Fatalf("err=%v", err)
	}
	if q.Get("a") != "1" || q.Get("b") != "2" {
		t.Fatalf("q=%v", q)
	}

	if err := applyQueryPairs(q, []string{"nope"}); err == nil {
		t.Fatalf("expected error")
	}

	if err := applyQueryPairs(q, []string{"=x"}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestReadJSONInput(t *testing.T) {
	if _, err := readJSONInput(`{"a":1}`, "x.json"); err == nil {
		t.Fatalf("expected error when both --data and --data-file are set")
	}

	b, err := readJSONInput(`{"a":1}`, "")
	if err != nil || string(b) != `{"a":1}` {
		t.Fatalf("b=%q err=%v", string(b), err)
	}

	if b, err := readJSONInput("", ""); err != nil || b != nil {
		t.Fatalf("b=%v err=%v", b, err)
	}

	dir := t.TempDir()
	p := filepath.Join(dir, "body.json")
	if err := os.WriteFile(p, []byte(`{"x":true}`), 0600); err != nil {
		t.Fatalf("write: %v", err)
	}
	b, err = readJSONInput("", p)
	if err != nil || string(bytes.TrimSpace(b)) != `{"x":true}` {
		t.Fatalf("b=%q err=%v", string(b), err)
	}

	// stdin
	oldStdin := os.Stdin
	oldStdinReader := stdinReader
	t.Cleanup(func() { os.Stdin = oldStdin })
	t.Cleanup(func() { stdinReader = oldStdinReader })
	r, w, _ := os.Pipe()
	os.Stdin = r
	stdinReader = r
	_, _ = w.Write([]byte(`{"stdin":true}`))
	_ = w.Close()
	b, err = readJSONInput("", "-")
	if err != nil || string(bytes.TrimSpace(b)) != `{"stdin":true}` {
		t.Fatalf("b=%q err=%v", string(b), err)
	}

	// missing file
	if _, err := readJSONInput("", filepath.Join(dir, "missing.json")); err == nil {
		t.Fatalf("expected error")
	}

	// stdin read error
	stdinReader = errReader{}
	if _, err := readJSONInput("", "-"); err == nil {
		t.Fatalf("expected error")
	}
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, os.ErrInvalid }

func TestReadJSONObjectInputAndMergeMaps(t *testing.T) {
	m, err := readJSONObjectInput("", "")
	if err != nil || len(m) != 0 {
		t.Fatalf("m=%#v err=%v", m, err)
	}

	if _, err := readJSONObjectInput("{", ""); err == nil {
		t.Fatalf("expected error")
	}

	if _, err := readJSONObjectInput("{}", "x.json"); err == nil {
		t.Fatalf("expected error")
	}

	if _, err := readJSONObjectInput("[]", ""); err == nil {
		t.Fatalf("expected error")
	}

	m, err = readJSONObjectInput(`{"a":1,"b":{"c":2}}`, "")
	if err != nil {
		t.Fatalf("m=%#v err=%v", m, err)
	}

	dst := map[string]any{"b": map[string]any{"d": 3}}
	mergeMaps(dst, m)
	bm := dst["b"].(map[string]any)
	// json.Unmarshal numeric types can vary; accept int/float64.
	asInt := func(v any) int {
		switch vv := v.(type) {
		case int:
			return vv
		case float64:
			return int(vv)
		default:
			return 0
		}
	}
	if asInt(bm["c"]) != 2 || asInt(bm["d"]) != 3 {
		t.Fatalf("dst=%#v", dst)
	}
}
