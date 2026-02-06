package api

import (
	"context"
	"net/url"
	"os"
	"testing"
	"time"
)

func TestLive_ReadOnlySmoke(t *testing.T) {
	if os.Getenv("INSTANTLY_LIVE_TESTS") != "1" {
		t.Skip("set INSTANTLY_LIVE_TESTS=1 to run live tests")
	}
	apiKey := os.Getenv("INSTANTLY_API_KEY")
	if apiKey == "" {
		t.Skip("set INSTANTLY_API_KEY to run live tests")
	}

	c := NewClient(DefaultBaseURL, apiKey, 60*time.Second)
	q := url.Values{"limit": []string{"1"}}
	out, _, err := c.GetJSON(context.Background(), "/accounts", q)
	if err != nil {
		t.Fatalf("GET /accounts failed: %v", err)
	}
	m, ok := out.(map[string]any)
	if !ok {
		t.Fatalf("unexpected response type: %T", out)
	}
	if _, ok := m["items"]; !ok {
		t.Fatalf("expected response to have items, got %#v", m)
	}
}
