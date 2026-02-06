package cmd

import (
	"bytes"
	"encoding/json"
	"testing"
)

type execResult struct {
	Stdout []byte
	Stderr []byte
	Err    error
}

func execCLI(t *testing.T, args ...string) execResult {
	t.Helper()
	c := newRootCmd()

	var out bytes.Buffer
	var errBuf bytes.Buffer
	c.SetOut(&out)
	c.SetErr(&errBuf)
	c.SetArgs(args)

	err := c.Execute()
	return execResult{Stdout: out.Bytes(), Stderr: errBuf.Bytes(), Err: err}
}

func mustJSON(t *testing.T, b []byte) any {
	t.Helper()
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatalf("invalid json: %v; data=%q", err, string(b))
	}
	return v
}
