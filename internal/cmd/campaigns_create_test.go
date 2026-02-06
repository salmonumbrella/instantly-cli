package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCampaignsCreate_AutoSenders_NonDryRun(t *testing.T) {
	var sawCampaigns bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/accounts":
			// Mix eligible and non-eligible accounts.
			_, _ = w.Write([]byte(`{
  "items": [
    null,
    {"email":"nope@example.com","status":2,"setup_pending":false,"warmup_status":1},
    {"email":"yes@example.com","status":1,"setup_pending":false,"warmup_status":1},
    {"email":"later@example.com","status":1,"setup_pending":false,"warmup_status":1}
  ]
}`))
			return
		case r.Method == http.MethodPost && r.URL.Path == "/campaigns":
			sawCampaigns = true
			var body any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			m := body.(map[string]any)
			emailList := m["email_list"].([]any)
			if len(emailList) != 1 || emailList[0].(string) != "yes@example.com" {
				t.Fatalf("email_list=%v", emailList)
			}
			_, _ = w.Write([]byte(`{"id":"cid"}`))
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
		"campaigns", "create",
		"--name", "n",
		"--subject", "sub\nj",
		"--body", "body",
		"--senders", "auto",
		"--senders-max", "1",
	)
	if res.Err != nil {
		t.Fatalf("err=%v stdout=%q", res.Err, string(res.Stdout))
	}
	if !sawCampaigns {
		t.Fatalf("expected /campaigns call")
	}
	if strings.Contains(string(res.Stdout), "\n  \"subject\": \"sub\\nj\"") {
		t.Fatalf("subject should be sanitized: %q", string(res.Stdout))
	}
}

func TestCampaignsCreate_AutoSenders_Errors(t *testing.T) {
	// /accounts request fails
	srv0 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/accounts" {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(`{"message":"boom"}`))
			return
		}
		w.WriteHeader(404)
	}))
	defer srv0.Close()

	res := execCLI(t,
		"--base-url", srv0.URL,
		"--api-key", "k",
		"campaigns", "create",
		"--name", "n",
		"--subject", "s",
		"--body", "b",
		"--senders", "auto",
	)
	if res.Err == nil {
		t.Fatalf("expected error")
	}

	// /accounts wrong shape
	srv1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/accounts" {
			_, _ = w.Write([]byte(`[]`))
			return
		}
		w.WriteHeader(404)
	}))
	defer srv1.Close()

	res = execCLI(t,
		"--base-url", srv1.URL,
		"--api-key", "k",
		"campaigns", "create",
		"--name", "n",
		"--subject", "s",
		"--body", "b",
		"--senders", "auto",
	)
	if res.Err == nil {
		t.Fatalf("expected error")
	}

	// No eligible senders
	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/accounts" {
			_, _ = w.Write([]byte(`{"items":[{"email":"x","status":2,"setup_pending":true,"warmup_status":0}]}`))
			return
		}
		w.WriteHeader(404)
	}))
	defer srv2.Close()

	res = execCLI(t,
		"--base-url", srv2.URL,
		"--api-key", "k",
		"campaigns", "create",
		"--name", "n",
		"--subject", "s",
		"--body", "b",
		"--senders", "auto",
	)
	if res.Err == nil {
		t.Fatalf("expected error")
	}
}
