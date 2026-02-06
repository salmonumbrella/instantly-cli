package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func withStdoutStderr(t *testing.T, fn func()) string {
	t.Helper()

	origOut := os.Stdout
	origErr := os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	done := make(chan struct{})
	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	go func() {
		_, _ = outBuf.ReadFrom(rOut)
		_, _ = errBuf.ReadFrom(rErr)
		close(done)
	}()

	fn()

	_ = wOut.Close()
	_ = wErr.Close()
	<-done

	os.Stdout = origOut
	os.Stderr = origErr
	return outBuf.String()
}

func TestExecute(t *testing.T) {
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	os.Args = []string{"instantly", "version", "--output", "json"}
	stdout := withStdoutStderr(t, func() {
		if err := Execute(); err != nil {
			t.Fatalf("err=%v", err)
		}
	})
	if stdout == "" {
		t.Fatalf("expected output")
	}

	os.Args = []string{"instantly", "version", "--output", "nope"}
	_ = withStdoutStderr(t, func() {
		if err := Execute(); err == nil {
			t.Fatalf("expected error")
		}
	})
}

func TestClientFromFlags(t *testing.T) {
	resetFlagsToDefaults()
	flags.APIKey = ""
	flags.DryRun = false
	if _, err := clientFromFlags(); err == nil {
		t.Fatalf("expected error")
	}

	resetFlagsToDefaults()
	flags.APIKey = ""
	flags.DryRun = true
	if _, err := clientFromFlags(); err != nil {
		t.Fatalf("err=%v", err)
	}
}

func TestPersistentPreRunBehavior(t *testing.T) {
	// --json should force output json.
	res := execCLI(t, "version", "--json")
	if res.Err != nil {
		t.Fatalf("err=%v stdout=%q", res.Err, string(res.Stdout))
	}
	if !strings.HasPrefix(strings.TrimSpace(string(res.Stdout)), "{") {
		t.Fatalf("expected json output, got %q", string(res.Stdout))
	}

	// --json conflicts with explicit --output != json.
	res = execCLI(t, "version", "--output", "agent", "--json")
	if res.Err == nil {
		t.Fatalf("expected error")
	}

	// --jq with explicit --output text should error.
	res = execCLI(t, "version", "--output", "text", "--jq", ".version")
	if res.Err == nil {
		t.Fatalf("expected error")
	}

	// If default output is text (via env), --jq should auto-switch to json.
	old := os.Getenv("INSTANTLY_OUTPUT")
	t.Cleanup(func() { _ = os.Setenv("INSTANTLY_OUTPUT", old) })
	_ = os.Setenv("INSTANTLY_OUTPUT", "text")
	res = execCLI(t, "version", "--jq", ".version")
	if res.Err != nil {
		t.Fatalf("err=%v stdout=%q", res.Err, string(res.Stdout))
	}
}

func TestQuietAndSilent(t *testing.T) {
	res := execCLI(t, "version", "--output", "text", "--quiet")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}
	if len(res.Stdout) != 0 {
		t.Fatalf("expected suppressed stdout, got %q", string(res.Stdout))
	}

	res = execCLI(t, "version", "--output", "json", "--silent")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}
	if len(res.Stdout) != 0 {
		t.Fatalf("expected no stdout, got %q", string(res.Stdout))
	}
}
