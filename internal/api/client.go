package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var randIntn = rand.Intn
var randInt63n = rand.Int63n
var jsonMarshal = json.Marshal
var jsonUnmarshal = json.Unmarshal

const (
	DefaultBaseURL = "https://api.instantly.ai/api/v2"
)

// Client handles Instantly API communication.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	APIKey     string

	DryRun bool

	Max429Retries  int
	Max5xxRetries  int
	RetryDelay     time.Duration
	MaxRetryDelay  time.Duration
	IdempotencyKey string
}

// NewClient creates a new API client.
func NewClient(baseURL, apiKey string, timeout time.Duration) *Client {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = DefaultBaseURL
	}
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
		RetryDelay:    1 * time.Second,
		MaxRetryDelay: 30 * time.Second,
	}
}

type RateLimitInfo struct {
	Remaining *int       `json:"remaining,omitempty"`
	Limit     *int       `json:"limit,omitempty"`
	ResetAt   *time.Time `json:"reset_at,omitempty"`
}

type Meta struct {
	Request struct {
		Method string `json:"method"`
		URL    string `json:"url"`
	} `json:"request"`
	RateLimit *RateLimitInfo `json:"rate_limit,omitempty"`
}

type APIError struct {
	Status  int
	Message string
	Body    []byte
}

func (e *APIError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return fmt.Sprintf("instantly api error (http %d): %s", e.Status, e.Message)
	}
	return fmt.Sprintf("instantly api error (http %d)", e.Status)
}

func (c *Client) ensureAPIKey() error {
	if strings.TrimSpace(c.APIKey) == "" {
		return errors.New("missing API key: set INSTANTLY_API_KEY or pass --api-key")
	}
	return nil
}

func parseRateLimit(headers http.Header) *RateLimitInfo {
	parseIntPtr := func(key string) *int {
		v := strings.TrimSpace(headers.Get(key))
		if v == "" {
			return nil
		}
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil
		}
		return &n
	}

	remaining := parseIntPtr("x-ratelimit-remaining")
	limit := parseIntPtr("x-ratelimit-limit")

	var resetAt *time.Time
	if resetRaw := strings.TrimSpace(headers.Get("x-ratelimit-reset")); resetRaw != "" {
		if sec, err := strconv.ParseInt(resetRaw, 10, 64); err == nil {
			t := time.Unix(sec, 0).UTC()
			resetAt = &t
		}
	}

	if remaining == nil && limit == nil && resetAt == nil {
		return nil
	}

	return &RateLimitInfo{Remaining: remaining, Limit: limit, ResetAt: resetAt}
}

func parseAPIErrorMessage(status int, body []byte) string {
	// Match the Python MCP server behavior: handle a few common error shapes.
	var data any
	if err := json.Unmarshal(body, &data); err != nil {
		// Fall back to a small text snippet to keep errors compact for agents.
		msg := strings.TrimSpace(string(body))
		if len(msg) > 300 {
			msg = msg[:300] + "..."
		}
		return msg
	}

	m, ok := data.(map[string]any)
	if !ok {
		return fmt.Sprintf("http %d", status)
	}

	// Prefer a specific message/detail field over a generic error label.
	if msg, ok := m["message"].(string); ok && msg != "" {
		return msg
	}
	if msg, ok := m["detail"].(string); ok && msg != "" {
		return msg
	}

	if v, ok := m["error"]; ok {
		switch vv := v.(type) {
		case string:
			return vv
		case map[string]any:
			if msg, ok := vv["message"].(string); ok && msg != "" {
				return msg
			}
			if msg, ok := vv["detail"].(string); ok && msg != "" {
				return msg
			}
		}
	}

	return fmt.Sprintf("http %d", status)
}

func (c *Client) do(ctx context.Context, method, path string, query url.Values, body []byte) ([]byte, *Meta, error) {
	base := strings.TrimRight(c.BaseURL, "/")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	fullURL := base + path
	if len(query) > 0 {
		fullURL = fullURL + "?" + query.Encode()
	}

	if c.DryRun {
		// Never touch the network. Return a small JSON payload describing the request.
		type dryRunRequest struct {
			Method string              `json:"method"`
			URL    string              `json:"url"`
			Query  map[string][]string `json:"query,omitempty"`
			Body   any                 `json:"body,omitempty"`
		}
		type dryRunResponse struct {
			DryRun  bool          `json:"dry_run"`
			Request dryRunRequest `json:"request"`
		}

		out := dryRunResponse{
			DryRun: true,
			Request: dryRunRequest{
				Method: method,
				URL:    fullURL,
			},
		}
		if len(query) > 0 {
			out.Request.Query = map[string][]string(query)
		}
		if len(body) > 0 {
			var v any
			if err := jsonUnmarshal(body, &v); err == nil {
				out.Request.Body = v
			} else {
				out.Request.Body = string(body)
			}
		}

		meta := &Meta{}
		meta.Request.Method = method
		meta.Request.URL = fullURL

		b, err := jsonMarshal(out)
		if err != nil {
			return nil, meta, fmt.Errorf("encode dry-run response: %w", err)
		}
		return b, meta, nil
	}

	if err := c.ensureAPIKey(); err != nil {
		return nil, nil, err
	}

	canRetryWrite := method == http.MethodGet || strings.TrimSpace(c.IdempotencyKey) != ""

	retries429 := 0
	retries5xx := 0
	var lastMeta *Meta
	for {
		var bodyReader io.Reader
		if body != nil {
			bodyReader = bytes.NewReader(body)
		}

		req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
		if err != nil {
			return nil, lastMeta, fmt.Errorf("create request: %w", err)
		}

		// Some Instantly endpoints reject empty JSON bodies when a JSON content-type is present.
		// Only set content-type when we actually send a non-empty JSON payload.
		if len(body) > 0 {
			req.Header.Set("Content-Type", "application/json")
		}
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
		if strings.TrimSpace(c.IdempotencyKey) != "" {
			req.Header.Set("Idempotency-Key", c.IdempotencyKey)
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			// Network errors are treated like 5xx retries for GETs / idempotent writes.
			if (method == http.MethodGet || canRetryWrite) && retries5xx < c.Max5xxRetries && c.Max5xxRetries > 0 {
				retries5xx++
				if sleepErr := sleepWithBackoff(ctx, c.RetryDelay, c.MaxRetryDelay, retries5xx); sleepErr != nil {
					return nil, lastMeta, sleepErr
				}
				continue
			}
			return nil, lastMeta, fmt.Errorf("request failed: %w", err)
		}

		respBody, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			return nil, lastMeta, fmt.Errorf("read response: %w", readErr)
		}

		meta := &Meta{}
		meta.Request.Method = method
		meta.Request.URL = fullURL
		meta.RateLimit = parseRateLimit(resp.Header)
		lastMeta = meta

		if resp.StatusCode == http.StatusTooManyRequests && (method == http.MethodGet || canRetryWrite) && c.Max429Retries > 0 && retries429 < c.Max429Retries {
			retries429++
			delay := retryAfterDelay(resp.Header, c.RetryDelay)
			if delay > c.MaxRetryDelay && c.MaxRetryDelay > 0 {
				delay = c.MaxRetryDelay
			}
			if sleepErr := sleepWithJitter(ctx, delay); sleepErr != nil {
				return nil, meta, sleepErr
			}
			continue
		}

		if resp.StatusCode >= 500 && resp.StatusCode <= 599 && (method == http.MethodGet || canRetryWrite) && c.Max5xxRetries > 0 && retries5xx < c.Max5xxRetries {
			retries5xx++
			if sleepErr := sleepWithBackoff(ctx, c.RetryDelay, c.MaxRetryDelay, retries5xx); sleepErr != nil {
				return nil, meta, sleepErr
			}
			continue
		}

		if resp.StatusCode >= 400 {
			return nil, meta, &APIError{
				Status:  resp.StatusCode,
				Message: parseAPIErrorMessage(resp.StatusCode, respBody),
				Body:    respBody,
			}
		}

		return respBody, meta, nil
	}
}

func retryAfterDelay(headers http.Header, fallback time.Duration) time.Duration {
	v := strings.TrimSpace(headers.Get("Retry-After"))
	if v == "" {
		return fallback
	}
	// Retry-After can be seconds.
	if sec, err := strconv.Atoi(v); err == nil && sec > 0 {
		return time.Duration(sec) * time.Second
	}
	return fallback
}

func sleepWithBackoff(ctx context.Context, base, max time.Duration, attempt int) error {
	if base <= 0 {
		base = 1 * time.Second
	}
	delay := base
	// Exponential: base * 2^(attempt-1)
	for i := 1; i < attempt; i++ {
		delay *= 2
		if max > 0 && delay >= max {
			delay = max
			break
		}
	}
	return sleepWithJitter(ctx, delay)
}

func sleepWithJitter(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	// +/- 20% jitter.
	j := time.Duration(randInt63n(int64(d/5) + 1)) // d/5 = 20%
	if randIntn(2) == 0 {
		d -= j
	} else {
		d += j
	}

	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func (c *Client) GetJSON(ctx context.Context, path string, query url.Values) (any, *Meta, error) {
	body, meta, err := c.do(ctx, http.MethodGet, path, query, nil)
	if err != nil {
		return nil, meta, err
	}
	if len(body) == 0 {
		return map[string]any{"success": true}, meta, nil
	}
	var v any
	if err := json.Unmarshal(body, &v); err != nil {
		return nil, meta, fmt.Errorf("decode json: %w", err)
	}
	return v, meta, nil
}

func (c *Client) PostJSON(ctx context.Context, path string, query url.Values, payload any) (any, *Meta, error) {
	var body []byte
	if payload != nil {
		var err error
		body, err = json.Marshal(payload)
		if err != nil {
			return nil, nil, fmt.Errorf("encode json: %w", err)
		}
	}
	respBody, meta, err := c.do(ctx, http.MethodPost, path, query, body)
	if err != nil {
		return nil, meta, err
	}
	if len(respBody) == 0 {
		return map[string]any{"success": true}, meta, nil
	}
	var v any
	if err := json.Unmarshal(respBody, &v); err != nil {
		return nil, meta, fmt.Errorf("decode json: %w", err)
	}
	return v, meta, nil
}

func (c *Client) PatchJSON(ctx context.Context, path string, query url.Values, payload any) (any, *Meta, error) {
	var body []byte
	if payload != nil {
		var err error
		body, err = json.Marshal(payload)
		if err != nil {
			return nil, nil, fmt.Errorf("encode json: %w", err)
		}
	}
	respBody, meta, err := c.do(ctx, http.MethodPatch, path, query, body)
	if err != nil {
		return nil, meta, err
	}
	if len(respBody) == 0 {
		return map[string]any{"success": true}, meta, nil
	}
	var v any
	if err := json.Unmarshal(respBody, &v); err != nil {
		return nil, meta, fmt.Errorf("decode json: %w", err)
	}
	return v, meta, nil
}

func (c *Client) DeleteJSON(ctx context.Context, path string, query url.Values) (any, *Meta, error) {
	respBody, meta, err := c.do(ctx, http.MethodDelete, path, query, nil)
	if err != nil {
		return nil, meta, err
	}
	if len(respBody) == 0 {
		return map[string]any{"success": true}, meta, nil
	}
	var v any
	if err := json.Unmarshal(respBody, &v); err != nil {
		return nil, meta, fmt.Errorf("decode json: %w", err)
	}
	return v, meta, nil
}

// DeleteJSONWithBody performs a DELETE with an optional JSON request body.
// Some Instantly endpoints validate DELETE bodies and can return 5xx without them.
func (c *Client) DeleteJSONWithBody(ctx context.Context, path string, query url.Values, payload any) (any, *Meta, error) {
	var body []byte
	if payload != nil {
		var err error
		body, err = json.Marshal(payload)
		if err != nil {
			return nil, nil, fmt.Errorf("encode json: %w", err)
		}
	}
	respBody, meta, err := c.do(ctx, http.MethodDelete, path, query, body)
	if err != nil {
		return nil, meta, err
	}
	if len(respBody) == 0 {
		return map[string]any{"success": true}, meta, nil
	}
	var v any
	if err := json.Unmarshal(respBody, &v); err != nil {
		return nil, meta, fmt.Errorf("decode json: %w", err)
	}
	return v, meta, nil
}
