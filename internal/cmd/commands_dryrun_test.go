package cmd

import (
	"strings"
	"testing"
)

func TestDryRun_AccountsCommands(t *testing.T) {
	res := execCLI(t, "--dry-run", "--output", "agent", "accounts", "list", "--limit", "1", "--starting-after", "c", "--search", "x", "--status", "1", "--provider", "2", "--query", "extra=yes")
	if res.Err != nil {
		t.Fatalf("err=%v stdout=%q", res.Err, string(res.Stdout))
	}

	res = execCLI(t, "--dry-run", "--output", "agent", "accounts", "list", "--query", "nope")
	if res.Err == nil {
		t.Fatalf("expected error")
	}

	res = execCLI(t, "--dry-run", "--output", "agent", "accounts", "get", " ")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "accounts", "get", "a@example.com")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}

	// create: missing email
	res = execCLI(t, "--dry-run", "--output", "agent", "accounts", "create")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
	// create: invalid json object
	res = execCLI(t, "--dry-run", "--output", "agent", "accounts", "create", "--data-json", "[]")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
	// create: success
	res = execCLI(t, "--dry-run", "--output", "agent", "accounts", "create",
		"--email", "a@example.com",
		"--first-name", "A",
		"--last-name", "B",
		"--provider-code", "1",
		"--imap-host", "imap",
		"--imap-port", "993",
		"--imap-username", "u",
		"--imap-password", "p",
		"--smtp-host", "smtp",
		"--smtp-port", "587",
		"--smtp-username", "u",
		"--smtp-password", "p",
	)
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}
	// create: success with --data-json merge
	res = execCLI(t, "--dry-run", "--output", "agent", "accounts", "create",
		"--data-json", `{"email":"b@example.com","smtp_host":"smtp2"}`,
		"--first-name", "B",
	)
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}

	// update: no fields
	res = execCLI(t, "--dry-run", "--output", "agent", "accounts", "update", "a@example.com")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "accounts", "update", " ", "--daily-limit", "1")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
	// update: invalid json object
	res = execCLI(t, "--dry-run", "--output", "agent", "accounts", "update", "a@example.com", "--data-json", "[]")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
	// update: success
	res = execCLI(t, "--dry-run", "--output", "agent", "accounts", "update", "a@example.com", "--daily-limit", "10")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "accounts", "update", "a@example.com",
		"--data-json", `{"sending_gap":5}`,
		"--first-name", "A",
		"--last-name", "B",
		"--sending-gap", "3",
		"--tracking-domain-name", "t.example.com",
	)
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}

	// warmup/test-vitals
	for _, sub := range []string{"warmup-enable", "warmup-disable", "test-vitals"} {
		res = execCLI(t, "--dry-run", "--output", "agent", "accounts", sub, " ")
		if res.Err == nil {
			t.Fatalf("expected error for %s", sub)
		}
		res = execCLI(t, "--dry-run", "--output", "agent", "accounts", sub, "a@example.com")
		if res.Err != nil {
			t.Fatalf("err=%v for %s", res.Err, sub)
		}
	}

	// delete confirm gate
	res = execCLI(t, "--dry-run", "--output", "agent", "accounts", "delete", "a@example.com")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "accounts", "delete", "--confirm", " ")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "accounts", "delete", "--confirm", "a@example.com")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}
}

func TestDryRun_CampaignsCommands(t *testing.T) {
	res := execCLI(t, "--dry-run", "--output", "agent", "campaigns", "list", "--limit", "1", "--starting-after", "c", "--search", "x", "--query", "q=1")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "campaigns", "list", "--query", "nope")
	if res.Err == nil {
		t.Fatalf("expected error")
	}

	res = execCLI(t, "--dry-run", "--output", "agent", "campaigns", "get", " ")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "campaigns", "get", "cid")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}

	// create missing flags
	res = execCLI(t, "--dry-run", "--output", "agent", "campaigns", "create")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
	// create missing --subject
	res = execCLI(t, "--dry-run", "--output", "agent", "campaigns", "create", "--name", "n", "--body", "b", "--subject", "")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
	// create missing --body
	res = execCLI(t, "--dry-run", "--output", "agent", "campaigns", "create", "--name", "n", "--subject", "s", "--body", "")
	if res.Err == nil {
		t.Fatalf("expected error")
	}

	// create with senders=auto (dry-run path uses <auto>)
	res = execCLI(t, "--dry-run", "--output", "agent", "campaigns", "create", "--name", "n", "--subject", "s", "--body", "b", "--senders", "auto")
	if res.Err != nil {
		t.Fatalf("err=%v stdout=%q", res.Err, string(res.Stdout))
	}
	if !strings.Contains(string(res.Stdout), `\u003cauto\u003e`) {
		t.Fatalf("expected <auto> placeholder, got %q", string(res.Stdout))
	}

	// create with explicit senders but empty list
	res = execCLI(t, "--dry-run", "--output", "agent", "campaigns", "create", "--name", "n", "--subject", "s", "--body", "b", "--senders", ",,,")
	if res.Err == nil {
		t.Fatalf("expected error")
	}

	// update no fields
	res = execCLI(t, "--dry-run", "--output", "agent", "campaigns", "update", "cid")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "campaigns", "update", " ", "--daily-limit", "5")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "campaigns", "update", "cid", "--daily-limit", "5", "--email-gap", "1", "--open-tracking", "--link-tracking", "--name", "x")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}

	// activate/pause
	for _, sub := range []string{"activate", "pause"} {
		res = execCLI(t, "--dry-run", "--output", "agent", "campaigns", sub, " ")
		if res.Err == nil {
			t.Fatalf("expected error")
		}
		res = execCLI(t, "--dry-run", "--output", "agent", "campaigns", sub, "cid")
		if res.Err != nil {
			t.Fatalf("err=%v", res.Err)
		}
	}

	// delete confirm
	res = execCLI(t, "--dry-run", "--output", "agent", "campaigns", "delete", "cid")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "campaigns", "delete", "--confirm", " ")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "campaigns", "delete", "--confirm", "cid")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}

	// search-by-contact
	res = execCLI(t, "--dry-run", "--output", "agent", "campaigns", "search-by-contact", " ")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "campaigns", "search-by-contact", "a@example.com")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}
}

func TestDryRun_LeadsAndListsAndJobsAndAnalytics(t *testing.T) {
	// leads list: invalid body merge
	res := execCLI(t, "--dry-run", "--output", "agent", "leads", "list", "--body-json", "[]")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
	// leads list: success
	res = execCLI(t, "--dry-run", "--output", "agent", "leads", "list", "--limit", "1", "--starting-after", "c", "--campaign", "cid", "--list-id", "lid", "--status", "active", "--search", "s", "--distinct")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}
	// leads list: success with body merge (covers mergeMaps call)
	res = execCLI(t, "--dry-run", "--output", "agent", "leads", "list", "--body-json", `{"foo":"bar"}`)
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}

	// leads get/create/update/delete
	res = execCLI(t, "--dry-run", "--output", "agent", "leads", "get", " ")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "leads", "get", "lid")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "leads", "create")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "leads", "create", "--email", "a@example.com", "--campaign", "cid", "--first-name", "A", "--last-name", "B", "--company-name", "C")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "leads", "update", " ")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
	// update: no fields
	res = execCLI(t, "--dry-run", "--output", "agent", "leads", "update", "lid")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "leads", "update", "lid", "--first-name", "A")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "leads", "update", "lid", "--last-name", "B", "--company-name", "C")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "leads", "delete", "lid")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "leads", "delete", "--confirm", " ")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "leads", "delete", "--confirm", "lid")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}

	// leads list
	res = execCLI(t, "--dry-run", "--output", "agent", "leads", "list")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "leads", "list", "--limit", "1", "--starting-after", "c", "--campaign", "cid", "--list-id", "lid", "--status", "s", "--search", "q", "--distinct")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}
	// leads list: body merge errors
	res = execCLI(t, "--dry-run", "--output", "agent", "leads", "list", "--body-json", "{")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "leads", "list", "--body-file", "/this/path/does/not/exist.json")
	if res.Err == nil {
		t.Fatalf("expected error")
	}

	// lead-lists
	res = execCLI(t, "--dry-run", "--output", "agent", "lead-lists", "list", "--limit", "1", "--starting-after", "c", "--search", "x", "--query", "q=1")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "lead-lists", "list", "--query", "nope")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "lead-lists", "get", " ")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "lead-lists", "get", "lid")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "lead-lists", "create")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "lead-lists", "create", "--name", "n", "--enrich")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "lead-lists", "update", "lid")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "lead-lists", "update", " ", "--name", "n")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "lead-lists", "update", "lid", "--name", "n", "--enrich")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "lead-lists", "verification-stats", " ")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "lead-lists", "verification-stats", "lid")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "lead-lists", "delete", "lid")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "lead-lists", "delete", "--confirm", " ")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "lead-lists", "delete", "--confirm", "lid")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}

	// jobs
	res = execCLI(t, "--dry-run", "--output", "agent", "jobs", "list", "--limit", "1", "--starting-after", "c", "--query", "q=1")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "jobs", "list", "--query", "nope")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "jobs", "get", " ")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "jobs", "get", "jid")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}

	// analytics
	res = execCLI(t, "--dry-run", "--output", "agent", "analytics", "campaign", "--campaign-id", "c", "--start-date", "2020-01-01", "--end-date", "2020-01-02", "--exclude-total-leads-count")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "analytics", "daily", "--campaign-id", "c", "--start-date", "2020-01-01", "--end-date", "2020-01-02", "--campaign-status", "-99")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "analytics", "warmup", "--emails", "a@example.com,b@example.com", "--start-date", "2020-01-01", "--end-date", "2020-01-02")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "analytics", "warmup", "--email", "a@example.com")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}
}

func TestDryRun_EmailsAndAPI(t *testing.T) {
	// emails list with unread flag set
	res := execCLI(t, "--dry-run", "--output", "agent", "emails", "list", "--limit", "1", "--starting-after", "c", "--search", "s", "--campaign-id", "cid", "--eaccount", "a@example.com", "--unread", "--query", "q=1")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "emails", "list", "--query", "nope")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
	// list without unread flag (covers isUnreadSet false path)
	res = execCLI(t, "--dry-run", "--output", "agent", "emails", "list", "--limit", "1")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}

	res = execCLI(t, "--dry-run", "--output", "agent", "emails", "get", " ")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "emails", "get", "eid")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}

	res = execCLI(t, "--dry-run", "--output", "agent", "emails", "unread-count")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}

	// reply confirm + payload shapes
	res = execCLI(t, "--dry-run", "--output", "agent", "emails", "reply", "--reply-to", "u", "--eaccount", "a@example.com", "--subject", "s", "--text", "t")
	if res.Err == nil {
		t.Fatalf("expected error without --confirm")
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "emails", "reply", "--confirm", "--eaccount", "a@example.com", "--subject", "s", "--text", "t")
	if res.Err == nil {
		t.Fatalf("expected error (missing --reply-to)")
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "emails", "reply", "--confirm", "--reply-to", "u", "--subject", "s", "--text", "t")
	if res.Err == nil {
		t.Fatalf("expected error (missing --eaccount)")
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "emails", "reply", "--confirm", "--reply-to", "u", "--eaccount", "a@example.com", "--text", "t")
	if res.Err == nil {
		t.Fatalf("expected error (missing --subject)")
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "emails", "reply", "--confirm", "--reply-to", "u", "--eaccount", "a@example.com", "--subject", "s", "--text", "t")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}
	if !strings.Contains(string(res.Stdout), "\"text\": \"t\"") {
		t.Fatalf("stdout=%q", string(res.Stdout))
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "emails", "reply", "--confirm", "--reply-to", "u", "--eaccount", "a@example.com", "--subject", "s", "--html", "<p>h</p>")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}
	if !strings.Contains(string(res.Stdout), `\u003cp\u003eh\u003c/p\u003e`) {
		t.Fatalf("stdout=%q", string(res.Stdout))
	}
	// legacy --body maps to text
	res = execCLI(t, "--dry-run", "--output", "agent", "emails", "reply", "--confirm", "--reply-to", "u", "--eaccount", "a@example.com", "--subject", "s", "--body", "legacy")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}
	if !strings.Contains(string(res.Stdout), "\"text\": \"legacy\"") {
		t.Fatalf("stdout=%q", string(res.Stdout))
	}
	// no content flags
	res = execCLI(t, "--dry-run", "--output", "agent", "emails", "reply", "--confirm", "--reply-to", "u", "--eaccount", "a@example.com", "--subject", "s")
	if res.Err == nil {
		t.Fatalf("expected error")
	}

	// verify requires email
	res = execCLI(t, "--dry-run", "--output", "agent", "emails", "verify")
	if res.Err == nil {
		t.Fatalf("expected error")
	}

	// mark-thread-read
	res = execCLI(t, "--dry-run", "--output", "agent", "emails", "mark-thread-read", " ")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "emails", "mark-thread-read", "tid")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}

	// api escape hatch
	res = execCLI(t, "--dry-run", "--output", "agent", "api", "get", "/accounts", "--query", "limit=1")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "api", "post", "/x", "--data", `{"a":1}`)
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "api", "patch", "/x", "--data", `{"a":1}`)
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "api", "delete", "/x")
	if res.Err != nil {
		t.Fatalf("err=%v", res.Err)
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "api", "post", "/x", "--data", "{")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "api", "post", "/x", "--data", `{}`, "--data-file", "x")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
	res = execCLI(t, "--dry-run", "--output", "agent", "api", "get", "/x", "--query", "nope")
	if res.Err == nil {
		t.Fatalf("expected error")
	}
}
