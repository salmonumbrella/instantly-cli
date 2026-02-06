package cmd

import "testing"

func TestPaginationFrom_TopLevel(t *testing.T) {
	resp := map[string]any{
		"items":               []any{map[string]any{"id": 1}},
		"next_starting_after": "cursor123",
	}
	got := paginationFrom(resp)
	if got == nil {
		t.Fatalf("expected pagination, got nil")
	}
	if got["has_more"] != true {
		t.Fatalf("has_more = %v, want true", got["has_more"])
	}
	if got["next_starting_after"] != "cursor123" {
		t.Fatalf("next_starting_after = %v, want cursor123", got["next_starting_after"])
	}
}

func TestPaginationFrom_Nested(t *testing.T) {
	resp := map[string]any{
		"items": []any{},
		"pagination": map[string]any{
			"next_starting_after": "nested456",
		},
	}
	got := paginationFrom(resp)
	if got == nil {
		t.Fatalf("expected pagination, got nil")
	}
	if got["next_starting_after"] != "nested456" {
		t.Fatalf("next_starting_after = %v, want nested456", got["next_starting_after"])
	}
}

func TestPaginationFrom_None(t *testing.T) {
	resp := map[string]any{
		"items": []any{},
	}
	got := paginationFrom(resp)
	if got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestPaginationFrom_NonMap(t *testing.T) {
	if got := paginationFrom([]any{}); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}
