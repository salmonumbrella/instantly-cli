package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestParseAPIErrorMessage_Shapes(t *testing.T) {
	cases := []struct {
		name string
		body []byte
		want string
	}{
		{"message", []byte(`{"message":"nope"}`), "nope"},
		{"detail", []byte(`{"detail":"nope2"}`), "nope2"},
		{"error_string", []byte(`{"error":"bad"}`), "bad"},
		{"error_obj_message", []byte(`{"error":{"message":"bad2"}}`), "bad2"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := parseAPIErrorMessage(400, tc.body)
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestClient_Retry429_GET(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"message":"rate limited"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "test", 5*time.Second)
	c.Max429Retries = 1
	c.RetryDelay = 10 * time.Millisecond
	c.MaxRetryDelay = 50 * time.Millisecond

	out, _, err := c.GetJSON(context.Background(), "/x", nil)
	if err != nil {
		t.Fatalf("GetJSON returned error: %v", err)
	}
	m, ok := out.(map[string]any)
	if !ok || m["ok"] != true {
		t.Fatalf("unexpected out: %#v", out)
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
}

func TestClient_NoRetryWritesWithoutIdempotency(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom"}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "test", 5*time.Second)
	c.Max5xxRetries = 2
	c.RetryDelay = 10 * time.Millisecond
	c.MaxRetryDelay = 50 * time.Millisecond

	_, _, err := c.PostJSON(context.Background(), "/x", nil, map[string]any{"a": 1})
	if err == nil {
		t.Fatalf("expected error")
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
}

func TestClient_RetryWritesWithIdempotency(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"message":"boom"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "test", 5*time.Second)
	c.IdempotencyKey = "idemp-1"
	c.Max5xxRetries = 2
	c.RetryDelay = 10 * time.Millisecond
	c.MaxRetryDelay = 50 * time.Millisecond

	out, _, err := c.PostJSON(context.Background(), "/x", nil, map[string]any{"a": 1})
	if err != nil {
		t.Fatalf("PostJSON returned error: %v", err)
	}
	m, ok := out.(map[string]any)
	if !ok || m["ok"] != true {
		t.Fatalf("unexpected out: %#v", out)
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
}

func TestClient_DryRun_NoNetwork_NoAPIKeyRequired(t *testing.T) {
	c := NewClient("https://example.invalid/api/v2", "", 2*time.Second)
	c.DryRun = true
	// If DryRun is broken and we touch the network, this should panic rather than silently doing IO.
	c.HTTPClient = nil

	resp, meta, err := c.PostJSON(context.Background(), "/campaigns", nil, map[string]any{"name": "x"})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if meta == nil || !strings.Contains(meta.Request.URL, "/campaigns") {
		t.Fatalf("meta.Request.URL = %#v", meta)
	}
	m, ok := resp.(map[string]any)
	if !ok {
		t.Fatalf("resp type = %T, want map", resp)
	}
	if dry, _ := m["dry_run"].(bool); !dry {
		t.Fatalf("dry_run = %v, want true", m["dry_run"])
	}
}

func TestNewClient_Defaults(t *testing.T) {
	c := NewClient("", "k", 0)
	if c.BaseURL != DefaultBaseURL {
		t.Fatalf("BaseURL=%q", c.BaseURL)
	}
	if c.HTTPClient.Timeout <= 0 {
		t.Fatalf("timeout=%v", c.HTTPClient.Timeout)
	}
}

func TestAPIError_Error(t *testing.T) {
	var e *APIError
	if e.Error() != "" {
		t.Fatalf("nil error string=%q", e.Error())
	}
	e = &APIError{Status: 400, Message: "bad"}
	if !strings.Contains(e.Error(), "bad") {
		t.Fatalf("error=%q", e.Error())
	}
	e = &APIError{Status: 500}
	if !strings.Contains(e.Error(), "http 500") {
		t.Fatalf("error=%q", e.Error())
	}
}

func TestEnsureAPIKey(t *testing.T) {
	c := NewClient(DefaultBaseURL, "", 1*time.Second)
	if err := c.ensureAPIKey(); err == nil {
		t.Fatalf("expected error")
	}
	c.APIKey = "x"
	if err := c.ensureAPIKey(); err != nil {
		t.Fatalf("err=%v", err)
	}
}

func TestParseRateLimit(t *testing.T) {
	if got := parseRateLimit(http.Header{}); got != nil {
		t.Fatalf("expected nil, got %#v", got)
	}
	h := http.Header{}
	h.Set("x-ratelimit-remaining", "9")
	h.Set("x-ratelimit-limit", "10")
	h.Set("x-ratelimit-reset", "1700000000")
	got := parseRateLimit(h)
	if got == nil || got.Remaining == nil || *got.Remaining != 9 {
		t.Fatalf("got=%#v", got)
	}
	if got.Limit == nil || *got.Limit != 10 {
		t.Fatalf("got=%#v", got)
	}
	if got.ResetAt == nil {
		t.Fatalf("got=%#v", got)
	}
}

func TestParseRateLimit_InvalidValues(t *testing.T) {
	h := http.Header{}
	h.Set("x-ratelimit-remaining", "nope")
	h.Set("x-ratelimit-limit", "nope")
	h.Set("x-ratelimit-reset", "nope")
	if got := parseRateLimit(h); got != nil {
		t.Fatalf("got=%#v", got)
	}

	h = http.Header{}
	h.Set("x-ratelimit-remaining", "1")
	h.Set("x-ratelimit-reset", "nope")
	got := parseRateLimit(h)
	if got == nil || got.Remaining == nil || *got.Remaining != 1 {
		t.Fatalf("got=%#v", got)
	}
	if got.ResetAt != nil {
		t.Fatalf("expected nil reset_at, got=%#v", got)
	}
}

func TestParseAPIErrorMessage_Fallbacks(t *testing.T) {
	long := strings.Repeat("x", 400)
	msg := parseAPIErrorMessage(500, []byte(long))
	if !strings.HasSuffix(msg, "...") {
		t.Fatalf("msg=%q", msg)
	}

	if got := parseAPIErrorMessage(400, []byte(`[]`)); got != "http 400" {
		t.Fatalf("got=%q", got)
	}

	if got := parseAPIErrorMessage(400, []byte(`{"error":{"detail":"nope"}}`)); got != "nope" {
		t.Fatalf("got=%q", got)
	}

	if got := parseAPIErrorMessage(400, []byte(`{"error":{"foo":"bar"}}`)); got != "http 400" {
		t.Fatalf("got=%q", got)
	}
}

func TestRetryAfterDelay(t *testing.T) {
	if d := retryAfterDelay(http.Header{}, 7*time.Second); d != 7*time.Second {
		t.Fatalf("d=%v", d)
	}
	h := http.Header{}
	h.Set("Retry-After", "nope")
	if d := retryAfterDelay(h, 3*time.Second); d != 3*time.Second {
		t.Fatalf("d=%v", d)
	}
	h.Set("Retry-After", "2")
	if d := retryAfterDelay(h, 3*time.Second); d != 2*time.Second {
		t.Fatalf("d=%v", d)
	}
}

func TestSleepHelpers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := sleepWithJitter(ctx, 0); err != nil {
		t.Fatalf("err=%v", err)
	}

	oldIntn := randIntn
	oldInt63n := randInt63n
	t.Cleanup(func() {
		randIntn = oldIntn
		randInt63n = oldInt63n
	})

	// Cover both +/- jitter branches deterministically.
	randInt63n = func(_ int64) int64 { return 0 }
	randIntn = func(_ int) int { return 0 }
	if err := sleepWithJitter(ctx, 1*time.Millisecond); !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
	randIntn = func(_ int) int { return 1 }
	if err := sleepWithJitter(ctx, 1*time.Millisecond); !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}

	if err := sleepWithBackoff(ctx, 1*time.Millisecond, 2*time.Millisecond, 2); !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
	if err := sleepWithBackoff(ctx, 0, 0, 1); !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
}

type errReadCloser struct{}

func (e errReadCloser) Read([]byte) (int, error) { return 0, errors.New("read boom") }
func (e errReadCloser) Close() error             { return nil }

type rtFunc func(*http.Request) (*http.Response, error)

func (f rtFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestDo_ReadResponseError(t *testing.T) {
	c := NewClient("http://example.com", "k", 1*time.Second)
	c.HTTPClient.Transport = rtFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       errReadCloser{},
			Request:    r,
		}, nil
	})

	_, _, err := c.GetJSON(context.Background(), "/x", nil)
	if err == nil || !strings.Contains(err.Error(), "read response") {
		t.Fatalf("err=%v", err)
	}
}

func TestDo_CreateRequestError(t *testing.T) {
	c := NewClient("http://[::1", "k", 1*time.Second) // invalid base URL
	_, _, err := c.GetJSON(context.Background(), "/x", nil)
	if err == nil || !strings.Contains(err.Error(), "create request") {
		t.Fatalf("err=%v", err)
	}
}

func TestDo_EnsureAPIKeyError(t *testing.T) {
	c := NewClient("http://example.com", "", 1*time.Second)
	_, _, err := c.GetJSON(context.Background(), "/x", nil)
	if err == nil || !strings.Contains(err.Error(), "missing API key") {
		t.Fatalf("err=%v", err)
	}
}

func TestDo_NetworkError_NoRetry(t *testing.T) {
	c := NewClient("http://example.com", "k", 1*time.Second)
	c.HTTPClient.Transport = rtFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("network boom")
	})
	_, _, err := c.GetJSON(context.Background(), "/x", nil)
	if err == nil || !strings.Contains(err.Error(), "request failed") {
		t.Fatalf("err=%v", err)
	}
}

func TestDo_NetworkError_RetryThenSucceed(t *testing.T) {
	var calls int32
	c := NewClient("http://example.com", "k", 1*time.Second)
	c.Max5xxRetries = 1
	c.RetryDelay = 1 * time.Millisecond
	c.MaxRetryDelay = 1 * time.Millisecond

	oldIntn := randIntn
	oldInt63n := randInt63n
	t.Cleanup(func() {
		randIntn = oldIntn
		randInt63n = oldInt63n
	})
	randIntn = func(_ int) int { return 0 }
	randInt63n = func(_ int64) int64 { return 0 }

	c.HTTPClient.Transport = rtFunc(func(r *http.Request) (*http.Response, error) {
		if atomic.AddInt32(&calls, 1) == 1 {
			return nil, errors.New("network boom")
		}
		return &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       ioNopCloser{strings.NewReader(`{"ok":true}`)},
			Request:    r,
		}, nil
	})

	out, _, err := c.GetJSON(context.Background(), "/x", nil)
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	m := out.(map[string]any)
	if m["ok"] != true {
		t.Fatalf("out=%#v", out)
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("calls=%d", calls)
	}
}

type ioNopCloser struct{ io.Reader }

func (ioNopCloser) Close() error { return nil }

func TestDo_429_SleepErrorAndMaxDelayCap(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	c := NewClient("http://example.com", "k", 1*time.Second)
	c.Max429Retries = 1
	c.RetryDelay = 0
	c.MaxRetryDelay = 1 * time.Second
	c.HTTPClient.Transport = rtFunc(func(r *http.Request) (*http.Response, error) {
		h := http.Header{}
		h.Set("Retry-After", "10")
		return &http.Response{
			StatusCode: 429,
			Header:     h,
			Body:       ioNopCloser{strings.NewReader(`{"message":"rate limited"}`)},
			Request:    r,
		}, nil
	})

	_, _, err := c.GetJSON(ctx, "/x", nil)
	if err == nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
}

func TestDo_5xx_SleepError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	c := NewClient("http://example.com", "k", 1*time.Second)
	c.Max5xxRetries = 1
	c.RetryDelay = 1 * time.Second
	c.MaxRetryDelay = 1 * time.Second
	c.HTTPClient.Transport = rtFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 500,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       ioNopCloser{strings.NewReader(`{"message":"boom"}`)},
			Request:    r,
		}, nil
	})

	_, _, err := c.GetJSON(ctx, "/x", nil)
	if err == nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
}

func TestDo_DryRun_QueryAndInvalidBodyAndLeadingSlash(t *testing.T) {
	c := NewClient("http://example.com", "", 1*time.Second)
	c.DryRun = true

	respBody, _, err := c.do(context.Background(), http.MethodPost, "x", url.Values{"a": []string{"1"}}, []byte("not-json"))
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	var v any
	if err := json.Unmarshal(respBody, &v); err != nil {
		t.Fatalf("decode=%v", err)
	}
	m := v.(map[string]any)
	req := m["request"].(map[string]any)
	if req["method"] != "POST" {
		t.Fatalf("req=%#v", req)
	}
	if !strings.Contains(req["url"].(string), "/x") {
		t.Fatalf("req=%#v", req)
	}
	if req["body"] != "not-json" {
		t.Fatalf("req=%#v", req)
	}
	q := req["query"].(map[string]any)
	if q["a"].([]any)[0].(string) != "1" {
		t.Fatalf("q=%#v", q)
	}
}

func TestDo_DryRun_MarshalError(t *testing.T) {
	old := jsonMarshal
	t.Cleanup(func() { jsonMarshal = old })
	jsonMarshal = func(any) ([]byte, error) { return nil, errors.New("boom") }

	c := NewClient("http://example.com", "", 1*time.Second)
	c.DryRun = true
	_, _, err := c.do(context.Background(), http.MethodGet, "/x", nil, nil)
	if err == nil || !strings.Contains(err.Error(), "encode dry-run response") {
		t.Fatalf("err=%v", err)
	}
}

func TestDo_NetworkError_RetrySleepError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	c := NewClient("http://example.com", "k", 1*time.Second)
	c.Max5xxRetries = 1
	c.RetryDelay = 1 * time.Second
	c.MaxRetryDelay = 1 * time.Second
	c.HTTPClient.Transport = rtFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("network boom")
	})

	_, _, err := c.GetJSON(ctx, "/x", nil)
	if err == nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
}

func TestClient_EmptyBodySuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "k", 2*time.Second)
	out, _, err := c.GetJSON(context.Background(), "/x", nil)
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	m := out.(map[string]any)
	if m["success"] != true {
		t.Fatalf("out=%#v", out)
	}
}

func TestClient_DecodeError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not-json"))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "k", 2*time.Second)
	_, _, err := c.GetJSON(context.Background(), "/x", nil)
	if err == nil || !strings.Contains(err.Error(), "decode json") {
		t.Fatalf("err=%v", err)
	}
}

func TestClient_EncodeError(t *testing.T) {
	c := NewClient("http://example.com", "k", 1*time.Second)
	_, _, err := c.PostJSON(context.Background(), "/x", nil, map[string]any{"f": func() {}})
	if err == nil || !strings.Contains(err.Error(), "encode json") {
		t.Fatalf("err=%v", err)
	}
}

func TestClient_PatchAndDelete(t *testing.T) {
	var lastMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "k", 2*time.Second)

	_, _, err := c.PatchJSON(context.Background(), "/x", url.Values{"a": []string{"1"}}, map[string]any{"a": 1})
	if err != nil || lastMethod != http.MethodPatch {
		t.Fatalf("err=%v method=%s", err, lastMethod)
	}

	_, _, err = c.DeleteJSON(context.Background(), "/x", nil)
	if err != nil || lastMethod != http.MethodDelete {
		t.Fatalf("err=%v method=%s", err, lastMethod)
	}
}

func TestClient_IdempotencyKeyHeader(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("Idempotency-Key")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "k", 2*time.Second)
	c.IdempotencyKey = "idemp-xyz"
	_, _, err := c.PostJSON(context.Background(), "/x", nil, map[string]any{"a": 1})
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if got != "idemp-xyz" {
		t.Fatalf("got=%q", got)
	}
}

func TestClient_PostPayloadNilAndEmptyResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "k", 2*time.Second)
	out, _, err := c.PostJSON(context.Background(), "/x", nil, nil)
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if out.(map[string]any)["success"] != true {
		t.Fatalf("out=%#v", out)
	}
}

func TestClient_PatchEncodeAndDeleteDecodeErrors(t *testing.T) {
	c := NewClient("http://example.com", "k", 1*time.Second)
	_, _, err := c.PatchJSON(context.Background(), "/x", nil, map[string]any{"f": func() {}})
	if err == nil || !strings.Contains(err.Error(), "encode json") {
		t.Fatalf("err=%v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("nope"))
	}))
	defer srv.Close()

	c = NewClient(srv.URL, "k", 1*time.Second)
	_, _, err = c.DeleteJSON(context.Background(), "/x", nil)
	if err == nil || !strings.Contains(err.Error(), "decode json") {
		t.Fatalf("err=%v", err)
	}
}

func TestClient_PostPatchDecodeAndEmptyBranches(t *testing.T) {
	// decode errors
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("nope"))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "k", 1*time.Second)
	if _, _, err := c.PostJSON(context.Background(), "/x", nil, map[string]any{"a": 1}); err == nil || !strings.Contains(err.Error(), "decode json") {
		t.Fatalf("err=%v", err)
	}
	if _, _, err := c.PatchJSON(context.Background(), "/x", nil, map[string]any{"a": 1}); err == nil || !strings.Contains(err.Error(), "decode json") {
		t.Fatalf("err=%v", err)
	}

	// empty response for patch/delete
	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}))
	defer srv2.Close()

	c = NewClient(srv2.URL, "k", 1*time.Second)
	out, _, err := c.PatchJSON(context.Background(), "/x", nil, map[string]any{"a": 1})
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if out.(map[string]any)["success"] != true {
		t.Fatalf("out=%#v", out)
	}
	out, _, err = c.DeleteJSON(context.Background(), "/x", nil)
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if out.(map[string]any)["success"] != true {
		t.Fatalf("out=%#v", out)
	}
}

func TestClient_PatchAndDelete_DoError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		_, _ = w.Write([]byte(`{"message":"nope"}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "k", 1*time.Second)
	if _, _, err := c.PatchJSON(context.Background(), "/x", nil, map[string]any{"a": 1}); err == nil {
		t.Fatalf("expected error")
	}
	if _, _, err := c.DeleteJSON(context.Background(), "/x", nil); err == nil {
		t.Fatalf("expected error")
	}
}
