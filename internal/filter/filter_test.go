package filter

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/itchyny/gojq"
)

func TestNormalizeExpression_ZshEscapedBang(t *testing.T) {
	in := `.foo \!= "bar"`
	got := NormalizeExpression(in)
	want := `.foo != "bar"`
	if got != want {
		t.Fatalf("NormalizeExpression(%q) = %q, want %q", in, got, want)
	}
}

func TestApply_SingleResult(t *testing.T) {
	in := map[string]any{"name": "ok"}
	out, err := Apply(in, ".name")
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	if out != "ok" {
		t.Fatalf("out = %v, want ok", out)
	}
}

func TestApply_MultipleResults(t *testing.T) {
	in := []any{1.0, 2.0, 3.0}
	out, err := Apply(in, ".[]")
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	arr, ok := out.([]any)
	if !ok {
		t.Fatalf("out type = %T, want []any", out)
	}
	if len(arr) != 3 {
		t.Fatalf("len(out) = %d, want 3", len(arr))
	}
}

func TestApply_EmptyExpression(t *testing.T) {
	in := map[string]any{"a": 1}
	out, err := Apply(in, "")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	m, ok := out.(map[string]any)
	if !ok || m["a"] != 1 {
		t.Fatalf("out = %#v", out)
	}
}

func TestApply_InvalidExpression(t *testing.T) {
	_, err := Apply(map[string]any{"a": 1}, "???")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestApply_RuntimeError(t *testing.T) {
	_, err := Apply(map[string]any{"a": 1}, `error("boom")`)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestApply_NoResults(t *testing.T) {
	out, err := Apply(map[string]any{"a": 1}, "empty")
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	arr, ok := out.([]any)
	if !ok {
		t.Fatalf("out=%T, want []any", out)
	}
	if len(arr) != 0 {
		t.Fatalf("len(out)=%d, want 0", len(arr))
	}
}

func TestApply_NormalizeError(t *testing.T) {
	// json.Marshal cannot encode functions.
	_, err := Apply(map[string]any{"f": func() {}}, ".")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestApply_PanicRecovery(t *testing.T) {
	old := runQuery
	t.Cleanup(func() { runQuery = old })
	runQuery = func(_ *gojq.Query, _ any) gojq.Iter { panic("boom") }

	_, err := Apply(map[string]any{"a": 1}, ".a")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestApplyToJSON(t *testing.T) {
	out, err := ApplyToJSON([]byte(`{"a":1}`), ".a")
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	var v any
	if err := json.Unmarshal(out, &v); err != nil {
		t.Fatalf("decode out: %v", err)
	}
	if v != float64(1) {
		t.Fatalf("v=%v", v)
	}

	if _, err := ApplyToJSON([]byte(`{`), "."); err == nil {
		t.Fatalf("expected error")
	}
}

func TestApplyToJSON_EmptyExpression(t *testing.T) {
	in := []byte(`{"a":1}`)
	out, err := ApplyToJSON(in, "")
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if string(out) != string(in) {
		t.Fatalf("out=%q", string(out))
	}
}

func TestNormalize_UnmarshalError(t *testing.T) {
	old := jsonUnmarshal
	t.Cleanup(func() { jsonUnmarshal = old })
	jsonUnmarshal = func(_ []byte, _ any) error { return errors.New("boom") } // any error works

	_, err := Apply(map[string]any{"a": 1}, ".a")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestApplyToJSON_FilterError(t *testing.T) {
	_, err := ApplyToJSON([]byte(`{"a":1}`), "???")
	if err == nil {
		t.Fatalf("expected error")
	}
}
