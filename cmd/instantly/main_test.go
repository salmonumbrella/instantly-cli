package main

import (
	"os"
	"testing"
)

func TestRun_Success(t *testing.T) {
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	os.Args = []string{"instantly", "version", "--output", "json"}

	if code := run(); code != 0 {
		t.Fatalf("code = %d, want 0", code)
	}
}

func TestRun_Error(t *testing.T) {
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	os.Args = []string{"instantly", "version", "--output", "nope"}

	if code := run(); code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
}

func TestMain_UsesExit(t *testing.T) {
	oldExit := exit
	t.Cleanup(func() { exit = oldExit })

	var got int
	exit = func(code int) { got = code }

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	os.Args = []string{"instantly", "version", "--output", "nope"}

	main()
	if got != 1 {
		t.Fatalf("exit code = %d, want 1", got)
	}
}
