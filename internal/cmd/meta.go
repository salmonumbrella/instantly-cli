package cmd

import (
	"github.com/salmonumbrella/instantly-cli/internal/api"
)

func metaFrom(meta *api.Meta, resp any) map[string]any {
	out := map[string]any{}
	if meta != nil {
		if meta.Request.URL != "" {
			out["request_url"] = meta.Request.URL
		}
		if meta.RateLimit != nil {
			out["rate_limit"] = meta.RateLimit
		}
	}

	if p := paginationFrom(resp); p != nil {
		out["pagination"] = p
	}

	if len(out) == 0 {
		return nil
	}
	return out
}

func paginationFrom(resp any) map[string]any {
	m, ok := resp.(map[string]any)
	if !ok {
		return nil
	}

	// Instantly sometimes returns the cursor at the top-level.
	if next, _ := m["next_starting_after"].(string); next != "" {
		return map[string]any{
			"has_more":            true,
			"next_starting_after": next,
		}
	}

	// Other endpoints may wrap pagination in a nested object.
	if p, ok := m["pagination"].(map[string]any); ok {
		if next, _ := p["next_starting_after"].(string); next != "" {
			return map[string]any{
				"has_more":            true,
				"next_starting_after": next,
			}
		}
	}

	return nil
}
