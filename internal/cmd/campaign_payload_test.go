package cmd

import "testing"

func TestConvertLineBreaksToHTML(t *testing.T) {
	in := "a\r\nb\n\nc\n"
	got := convertLineBreaksToHTML(in)
	if got == "" {
		t.Fatalf("expected html")
	}
	if want := "<p>a<br />b</p><p>c</p>"; got != want {
		t.Fatalf("got=%q want=%q", got, want)
	}

	// Covers empty paragraph skipping.
	in = "\n\n\n\nx\n\n"
	got = convertLineBreaksToHTML(in)
	if got != "<p>x</p>" {
		t.Fatalf("got=%q", got)
	}
}

func TestBuildCreateCampaignPayload(t *testing.T) {
	p := buildCreateCampaignPayload("n", "sub\nj", "body", []string{"a@example.com"}, 1, 2)
	if p["name"] != "n" {
		t.Fatalf("p=%#v", p)
	}
	if p["daily_limit"].(int) != 1 || p["email_gap"].(int) != 2 {
		t.Fatalf("p=%#v", p)
	}
	seq := p["sequences"].([]any)
	if len(seq) != 1 {
		t.Fatalf("seq=%#v", seq)
	}
}
