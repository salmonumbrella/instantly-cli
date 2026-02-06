package agentfmt

import (
	"strings"
)

// ListEnvelope wraps list outputs.
type ListEnvelope struct {
	Kind  string         `json:"kind"`
	Items any            `json:"items"`
	Meta  map[string]any `json:"meta,omitempty"`
}

// ItemEnvelope wraps single-item outputs.
type ItemEnvelope struct {
	Kind string         `json:"kind"`
	Item any            `json:"item"`
	Meta map[string]any `json:"meta,omitempty"`
}

// DataEnvelope wraps untyped outputs.
type DataEnvelope struct {
	Kind string         `json:"kind"`
	Data any            `json:"data"`
	Meta map[string]any `json:"meta,omitempty"`
}

// KindFromCommandPath converts a cobra CommandPath to a dotted kind string.
func KindFromCommandPath(path string) string {
	path = strings.TrimSpace(path)
	path = strings.TrimPrefix(path, "instantly ")
	parts := strings.Fields(path)
	if len(parts) == 0 {
		return "unknown"
	}
	return strings.Join(parts, ".")
}

// Envelope wraps an API response in an agent-friendly structure.
//
// Heuristic:
// - If response is an object with an "items" array, treat as list.
// - Otherwise treat as single item/data.
func Envelope(kind string, resp any, meta map[string]any) any {
	if m, ok := resp.(map[string]any); ok {
		if items, hasItems := m["items"]; hasItems {
			return ListEnvelope{Kind: kind, Items: items, Meta: meta}
		}
		return ItemEnvelope{Kind: kind, Item: resp, Meta: meta}
	}
	return DataEnvelope{Kind: kind, Data: resp, Meta: meta}
}
