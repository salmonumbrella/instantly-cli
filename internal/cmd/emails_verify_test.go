package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestEmailsVerify_PollingAndTimeout(t *testing.T) {
	var getCalls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/email-verification":
			_, _ = w.Write([]byte(`{"verification_status":"pending"}`))
			return
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/email-verification/"):
			n := atomic.AddInt32(&getCalls, 1)
			if n == 1 {
				_, _ = w.Write([]byte(`{"verification_status":"verified"}`))
			} else {
				_, _ = w.Write([]byte(`{"verification_status":"pending"}`))
			}
			return
		default:
			w.WriteHeader(404)
			_, _ = w.Write([]byte(`{"message":"not found"}`))
		}
	}))
	defer srv.Close()

	email := "test@example.com"
	encoded := url.PathEscape(email)

	// Poll to completion.
	res := execCLI(t,
		"--base-url", srv.URL,
		"--api-key", "k",
		"--output", "json",
		"emails", "verify",
		"--email", email,
		"--max-wait", "50ms",
		"--poll-interval", "1ms",
	)
	if res.Err != nil {
		t.Fatalf("err=%v stdout=%q", res.Err, string(res.Stdout))
	}
	out := mustJSON(t, res.Stdout).(map[string]any)
	if out["verification_status"] != "verified" {
		t.Fatalf("out=%#v", out)
	}
	if _, ok := out["_polling_info"].(map[string]any); !ok {
		t.Fatalf("missing _polling_info: %#v", out)
	}
	if atomic.LoadInt32(&getCalls) == 0 {
		t.Fatalf("expected polling GET calls")
	}

	// Skip polling even if pending.
	getCalls = 0
	res = execCLI(t,
		"--base-url", srv.URL,
		"--api-key", "k",
		"--output", "json",
		"emails", "verify",
		"--email", email,
		"--skip-polling",
	)
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}
	out = mustJSON(t, res.Stdout).(map[string]any)
	if out["verification_status"] != "pending" {
		t.Fatalf("out=%#v", out)
	}

	// Defaults for max-wait/poll-interval when <= 0.
	srvDefaults := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost && r.URL.Path == "/email-verification" {
			_, _ = w.Write([]byte(`{"verification_status":"verified"}`))
			return
		}
		w.WriteHeader(404)
	}))
	defer srvDefaults.Close()

	res = execCLI(t,
		"--base-url", srvDefaults.URL,
		"--api-key", "k",
		"--output", "json",
		"emails", "verify",
		"--email", email,
		"--max-wait", "0s",
		"--poll-interval", "0s",
	)
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}

	// Timeout path (pending forever).
	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost && r.URL.Path == "/email-verification" {
			_, _ = w.Write([]byte(`{"verification_status":"pending"}`))
			return
		}
		if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/email-verification/") {
			_, _ = w.Write([]byte(`{"verification_status":"pending"}`))
			return
		}
		w.WriteHeader(404)
	}))
	defer srv2.Close()

	start := time.Now()
	res = execCLI(t,
		"--base-url", srv2.URL,
		"--api-key", "k",
		"--output", "json",
		"emails", "verify",
		"--email", email,
		"--max-wait", "5ms",
		"--poll-interval", "1ms",
	)
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}
	out = mustJSON(t, res.Stdout).(map[string]any)
	if out["verification_status"] != "pending" {
		t.Fatalf("out=%#v", out)
	}
	if info, ok := out["_polling_info"].(map[string]any); !ok || info["timeout_reached"] != true {
		t.Fatalf("out=%#v", out)
	}
	if time.Since(start) > 200*time.Millisecond {
		t.Fatalf("test took too long")
	}

	// Ensure the encoded email path is what we'd expect.
	if encoded == "" {
		t.Fatalf("bad path escape: %q", fmt.Sprint(encoded))
	}
}

func TestEmailsVerify_PollingContinuesOnErrors(t *testing.T) {
	var getCalls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/email-verification":
			_, _ = w.Write([]byte(`{"verification_status":"pending"}`))
			return
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/email-verification/"):
			n := atomic.AddInt32(&getCalls, 1)
			if n == 1 {
				w.WriteHeader(500)
				_, _ = w.Write([]byte(`{"message":"transient"}`))
				return
			}
			_, _ = w.Write([]byte(`{"verification_status":"verified"}`))
			return
		default:
			w.WriteHeader(404)
			_, _ = w.Write([]byte(`{"message":"not found"}`))
		}
	}))
	defer srv.Close()

	res := execCLI(t,
		"--base-url", srv.URL,
		"--api-key", "k",
		"--output", "json",
		"emails", "verify",
		"--email", "errpoll@example.com",
		"--max-wait", "50ms",
		"--poll-interval", "1ms",
	)
	if res.Err != nil {
		t.Fatalf("err=%v stdout=%q", res.Err, string(res.Stdout))
	}
	out := mustJSON(t, res.Stdout).(map[string]any)
	if out["verification_status"] != "verified" {
		t.Fatalf("out=%#v", out)
	}
	if atomic.LoadInt32(&getCalls) < 2 {
		t.Fatalf("expected at least 2 polling calls, got %d", atomic.LoadInt32(&getCalls))
	}
}
