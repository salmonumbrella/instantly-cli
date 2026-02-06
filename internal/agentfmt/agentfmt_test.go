package agentfmt

import "testing"

func TestKindFromCommandPath(t *testing.T) {
	if got := KindFromCommandPath("instantly campaigns list"); got != "campaigns.list" {
		t.Fatalf("got %q", got)
	}
	if got := KindFromCommandPath("  instantly   emails   reply "); got != "emails.reply" {
		t.Fatalf("got %q", got)
	}
	if got := KindFromCommandPath(""); got != "unknown" {
		t.Fatalf("got %q", got)
	}
}

func TestEnvelope_ListItemData(t *testing.T) {
	kind := "x.y"

	list := Envelope(kind, map[string]any{"items": []any{map[string]any{"id": 1}}}, map[string]any{"m": 1})
	if _, ok := list.(ListEnvelope); !ok {
		t.Fatalf("list envelope type = %T", list)
	}

	item := Envelope(kind, map[string]any{"id": 1}, nil)
	if _, ok := item.(ItemEnvelope); !ok {
		t.Fatalf("item envelope type = %T", item)
	}

	data := Envelope(kind, []any{1, 2}, nil)
	if _, ok := data.(DataEnvelope); !ok {
		t.Fatalf("data envelope type = %T", data)
	}
}
