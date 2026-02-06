package cmd

import "testing"

func TestDryRun_ExpandedAPICommands(t *testing.T) {
	base := []string{"--dry-run", "--output", "agent"}

	mustErr := func(args ...string) {
		t.Helper()
		tc := append(append([]string{}, base...), args...)
		res := execCLI(t, tc...)
		if res.Err == nil {
			t.Fatalf("expected error for %v; stdout=%q", tc, string(res.Stdout))
		}
	}
	mustOK := func(args ...string) {
		t.Helper()
		tc := append(append([]string{}, base...), args...)
		res := execCLI(t, tc...)
		if res.Err != nil {
			t.Fatalf("unexpected error for %v: %v stdout=%q", tc, res.Err, string(res.Stdout))
		}
	}

	// accounts extra
	mustErr("accounts", "analytics-daily", "--query", "nope")
	mustOK("accounts", "analytics-daily", "--limit", "1", "--starting-after", "c", "--query", "x=1")
	for _, sub := range []string{"warmup-enable", "warmup-disable", "test-vitals"} {
		mustErr("accounts", sub, " ")
		mustOK("accounts", sub, "a@example.com")
	}

	// account-campaign-mappings
	mustErr("account-campaign-mappings", "get", " ")
	mustOK("account-campaign-mappings", "get", "a@example.com")

	// campaigns extra
	mustErr("campaigns", "analytics-overview", "--query", "nope")
	mustOK("campaigns", "analytics-overview", "--query", "id=cid")
	mustErr("campaigns", "analytics-steps", "--query", "nope")
	mustOK("campaigns", "analytics-steps", "--query", "campaign_id=cid")

	// emails extra
	mustErr("emails", "forward", "--data-json", `{}`)
	mustErr("emails", "forward", "--confirm")
	mustErr("emails", "forward", "--confirm", "--data-json", "[") // invalid JSON
	mustOK("emails", "forward", "--confirm", "--data-json", `{"email_id":"eid","to":["a@example.com"]}`)
	mustErr("emails", "update", " ")
	mustErr("emails", "update", "eid")
	mustErr("emails", "update", "eid", "--data-json", "[]")
	mustErr("emails", "update", "eid", "--data-json", "[") // invalid JSON
	mustOK("emails", "update", "eid", "--data-json", `{"marked_as_done":true}`)
	mustErr("emails", "delete", "eid")
	mustErr("emails", "delete", "--confirm", " ")
	mustOK("emails", "delete", "--confirm", "eid")

	// leads extra
	mustErr("leads", "bulk-delete")
	mustErr("leads", "bulk-delete", "--confirm")
	mustErr("leads", "bulk-delete", "--confirm", "--query", "nope")
	mustOK("leads", "bulk-delete", "--confirm", "--query", "campaign_id=cid")
	for _, sub := range []string{"merge", "update-interest-status"} {
		mustErr("leads", sub)
		mustErr("leads", sub, "--confirm")
		mustErr("leads", sub, "--confirm", "--data-json", "[]")
		mustOK("leads", sub, "--confirm", "--data-json", `{"x":1}`)
	}

	// webhooks
	mustErr("webhooks", "list", "--query", "nope")
	mustOK("webhooks", "list", "--limit", "1", "--starting-after", "c", "--campaign", "cid", "--event-type", "email_sent", "--query", "x=1")
	mustErr("webhooks", "get", " ")
	mustOK("webhooks", "get", "wid")
	mustErr("webhooks", "create")                     // missing target url
	mustErr("webhooks", "create", "--data-json", "[") // invalid JSON
	mustErr("webhooks", "create", "--target-url", "x", "--headers-json", "{")
	mustErr("webhooks", "create", "--target-url", "x", "--headers-json", "[]")
	mustOK("webhooks", "create",
		"--data-json", `{"base":true}`,
		"--target-url", "https://example.com/hook",
		"--campaign", "cid",
		"--name", "n",
		"--event-type", "all_events",
		"--custom-interest-value", "2.5",
		"--headers-json", `{"X-Test":"1"}`,
	)
	mustErr("webhooks", "update", " ")
	mustErr("webhooks", "update", "wid") // empty body
	mustErr("webhooks", "update", "wid", "--data-json", "[")
	mustOK("webhooks", "update", "wid", "--data-json", `{"name":"n"}`)
	mustErr("webhooks", "delete", "wid")
	mustErr("webhooks", "delete", "--confirm", " ")
	mustOK("webhooks", "delete", "--confirm", "wid")
	mustOK("webhooks", "event-types")
	mustErr("webhooks", "test", "wid")
	mustErr("webhooks", "test", "--confirm", " ")
	mustOK("webhooks", "test", "--confirm", "wid")
	mustErr("webhooks", "resume", " ")
	mustOK("webhooks", "resume", "wid")
	mustErr("webhooks", "events", "list", "--query", "nope")
	mustOK("webhooks", "events", "list", "--limit", "1", "--starting-after", "c", "--query", "x=1")
	mustErr("webhooks", "events", "get", " ")
	mustOK("webhooks", "events", "get", "eid")
	mustErr("webhooks", "events", "summary", "--query", "nope")
	mustOK("webhooks", "events", "summary", "--query", "x=1")
	mustErr("webhooks", "events", "summary-by-date", "--query", "nope")
	mustOK("webhooks", "events", "summary-by-date", "--query", "x=1")

	// tags + mappings
	mustErr("custom-tags", "list", "--query", "nope")
	mustOK("custom-tags", "list", "--limit", "1", "--starting-after", "c", "--search", "x", "--query", "x=1")
	mustErr("custom-tags", "get", " ")
	mustOK("custom-tags", "get", "tid")
	mustErr("custom-tags", "create")                               // missing name
	mustErr("custom-tags", "create", "--data-json", "[")           // invalid JSON
	mustOK("custom-tags", "create", "--data-json", `{"name":"n"}`) // accept legacy name -> label mapping
	mustOK("custom-tags", "create", "--name", "n", "--color", "red", "--data-json", `{"x":1}`)
	mustErr("custom-tags", "update", " ")
	mustErr("custom-tags", "update", "tid")                               // no fields
	mustErr("custom-tags", "update", "tid", "--data-json", "[")           // invalid JSON
	mustOK("custom-tags", "update", "tid", "--data-json", `{"name":"n"}`) // accept legacy name -> label mapping
	mustOK("custom-tags", "update", "tid", "--name", "n", "--color", "blue")
	mustErr("custom-tags", "delete", "tid")
	mustErr("custom-tags", "delete", "--confirm", " ")
	mustOK("custom-tags", "delete", "--confirm", "tid")
	mustErr("custom-tags", "toggle-resource")                                               // missing required
	mustErr("custom-tags", "toggle-resource", "--data-json", "[")                           // invalid JSON
	mustErr("custom-tags", "toggle-resource", "--tag-id", "tid", "--resource-type", "lead") // missing resource id
	mustErr("custom-tags", "toggle-resource", "--tag-id", "tid", "--resource-id", "rid")    // missing resource type
	mustOK("custom-tags", "toggle-resource", "--tag-id", "tid", "--resource-id", "rid", "--resource-type", "lead", "--enabled=false", "--data-json", `{"x":1}`)
	mustOK("custom-tags", "mappings", "--limit", "1", "--starting-after", "c", "--query", "x=1")
	mustErr("custom-tags", "mappings", "--query", "nope")

	// block list entries
	mustErr("block-list-entries", "list", "--query", "nope")
	mustOK("block-list-entries", "list", "--limit", "1", "--starting-after", "c", "--search", "x", "--query", "x=1")
	mustErr("block-list-entries", "get", " ")
	mustOK("block-list-entries", "get", "bid")
	mustErr("block-list-entries", "create") // missing body
	mustErr("block-list-entries", "create", "--data-json", "[]")
	mustOK("block-list-entries", "create", "--data-json", `{"value":"x"}`)
	mustErr("block-list-entries", "update", " ")
	mustErr("block-list-entries", "update", "bid")                     // empty
	mustErr("block-list-entries", "update", "bid", "--data-json", "[") // invalid JSON
	mustOK("block-list-entries", "update", "bid", "--data-json", `{"value":"y"}`)
	mustErr("block-list-entries", "delete", "bid")
	mustErr("block-list-entries", "delete", "--confirm", " ")
	mustOK("block-list-entries", "delete", "--confirm", "bid")

	// lead labels
	mustErr("lead-labels", "list", "--query", "nope")
	mustOK("lead-labels", "list", "--limit", "1", "--starting-after", "c", "--search", "x", "--query", "x=1")
	mustErr("lead-labels", "get", " ")
	mustOK("lead-labels", "get", "lid")
	mustErr("lead-labels", "create")                                // missing name
	mustErr("lead-labels", "create", "--data-json", "[")            // invalid JSON
	mustOK("lead-labels", "create", "--data-json", `{"name":"n"}`)  // accept legacy name -> label mapping
	mustOK("lead-labels", "create", "--data-json", `{"label":"n"}`) // accept label alias
	mustOK("lead-labels", "create", "--data-json", `{"label":"n","interest_status_label":"neutral"}`)
	mustOK("lead-labels", "create", "--data-json", `{"label":"n","interest_status":"positive"}`)
	mustErr("lead-labels", "create", "--data-json", `{"label":"n","interest_status_label":"nope"}`)
	mustErr("lead-labels", "create", "--data-json", `{"label":"n","interest_status_label":1}`)
	mustErr("lead-labels", "create", "--data-json", `{"label":"n","interest_status_label":" "}`)
	mustOK("lead-labels", "create", "--name", "n", "--data-json", `{"x":1}`)
	mustErr("lead-labels", "update", " ")
	mustErr("lead-labels", "update", "lid")                                // no fields
	mustErr("lead-labels", "update", "lid", "--data-json", "[")            // invalid JSON
	mustOK("lead-labels", "update", "lid", "--data-json", `{"name":"n"}`)  // accept legacy name -> label mapping
	mustOK("lead-labels", "update", "lid", "--data-json", `{"label":"n"}`) // accept label alias
	mustOK("lead-labels", "update", "lid", "--data-json", `{"interest_status_label":"positive"}`)
	mustOK("lead-labels", "update", "lid", "--interest-status", "negative")
	mustOK("lead-labels", "update", "lid", "--data-json", `{"interest_status":"neutral"}`)
	mustErr("lead-labels", "update", "lid", "--data-json", `{"interest_status_label":"nope"}`)
	mustErr("lead-labels", "update", "lid", "--data-json", `{"interest_status_label":1}`)
	mustErr("lead-labels", "update", "lid", "--data-json", `{"interest_status_label":" "}`)
	mustOK("lead-labels", "update", "lid", "--name", "n")
	mustErr("lead-labels", "delete", "lid")
	mustErr("lead-labels", "delete", "--confirm", " ")
	mustOK("lead-labels", "delete", "--confirm", "lid")

	// subsequences
	mustErr("subsequences", "list") // missing parent
	mustErr("subsequences", "list", "--parent-campaign", "cid", "--query", "nope")
	mustOK("subsequences", "list", "--parent-campaign", "cid", "--limit", "1", "--starting-after", "c", "--search", "x", "--query", "x=1")
	mustErr("subsequences", "get", " ")
	mustOK("subsequences", "get", "sid")
	mustErr("subsequences", "create") // empty body
	mustErr("subsequences", "create", "--data-json", "[]")
	mustOK("subsequences", "create", "--data-json", `{"parent_campaign":"cid","name":"n"}`)
	mustErr("subsequences", "update", " ")
	mustErr("subsequences", "update", "sid")
	mustErr("subsequences", "update", "sid", "--data-json", "[") // invalid JSON
	mustOK("subsequences", "update", "sid", "--data-json", `{"name":"n"}`)
	mustErr("subsequences", "delete", "sid")
	mustErr("subsequences", "delete", "--confirm", " ")
	mustOK("subsequences", "delete", "--confirm", "sid")
	for _, sub := range []string{"pause", "resume"} {
		mustErr("subsequences", sub, "sid")
		mustErr("subsequences", sub, "--confirm", " ")
		mustErr("subsequences", sub, "--confirm", "sid", "--data-json", "[]")
		mustOK("subsequences", sub, "--confirm", "sid", "--data-json", `{"x":1}`)
	}
	mustErr("subsequences", "duplicate", "sid")
	mustErr("subsequences", "duplicate", "--confirm", " ")
	mustErr("subsequences", "duplicate", "--confirm", "sid", "--data-json", "[") // invalid JSON
	mustOK("subsequences", "duplicate", "--confirm", "sid", "--data-json", `{"x":1}`)

	// inbox placement
	mustErr("inbox-placement", "tests", "list", "--query", "nope")
	mustOK("inbox-placement", "tests", "list", "--limit", "1", "--starting-after", "c", "--query", "x=1")
	mustErr("inbox-placement", "tests", "get", " ")
	mustOK("inbox-placement", "tests", "get", "tid")
	mustErr("inbox-placement", "tests", "create")
	mustErr("inbox-placement", "tests", "create", "--confirm")
	mustErr("inbox-placement", "tests", "create", "--confirm", "--data-json", "[]")
	mustOK("inbox-placement", "tests", "create", "--confirm", "--data-json", `{"name":"t"}`)
	mustErr("inbox-placement", "tests", "update", " ")
	mustErr("inbox-placement", "tests", "update", "tid")
	mustErr("inbox-placement", "tests", "update", "tid", "--data-json", "[") // invalid JSON
	mustOK("inbox-placement", "tests", "update", "tid", "--data-json", `{"name":"t"}`)
	mustErr("inbox-placement", "tests", "delete", "tid")
	mustErr("inbox-placement", "tests", "delete", "--confirm", " ")
	mustOK("inbox-placement", "tests", "delete", "--confirm", "tid")
	mustOK("inbox-placement", "tests", "esps")
	mustErr("inbox-placement", "analytics", "list", "--query", "nope")
	mustErr("inbox-placement", "analytics", "list", "--limit", "1") // missing required test_id
	mustOK("inbox-placement", "analytics", "list", "--test-id", "tid", "--limit", "1", "--starting-after", "c", "--query", "x=1")
	mustErr("inbox-placement", "analytics", "get", " ")
	mustOK("inbox-placement", "analytics", "get", "aid")
	for _, sub := range []string{"stats-by-test-id", "deliverability-insights"} {
		mustErr("inbox-placement", "analytics", sub)                     // missing required test_id
		mustErr("inbox-placement", "analytics", sub, "--data-json", "[") // invalid JSON
		mustOK("inbox-placement", "analytics", sub, "--test-id", "tid", "--data-json", `{"x":1}`)
		mustOK("inbox-placement", "analytics", sub, "--data-json", `{"test_id":"tid","x":1}`) // test_id supplied via body
	}
	mustErr("inbox-placement", "analytics", "stats-by-date")                     // missing required test_id
	mustErr("inbox-placement", "analytics", "stats-by-date", "--data-json", "[") // invalid JSON
	mustErr("inbox-placement", "analytics", "stats-by-date", "--test-id", "tid", "--data-file", "/this/path/does/not/exist.json")
	mustOK("inbox-placement", "analytics", "stats-by-date", "--test-id", "tid", "--start-date", "2026-01-01", "--end-date", "2026-01-31")
	mustOK("inbox-placement", "analytics", "stats-by-date", "--test-id", "tid", "--data-json", `{"x":1}`)
	mustOK("inbox-placement", "analytics", "stats-by-date", "--data-json", `{"test_id":"tid","start_date":"2026-01-01","end_date":"2026-01-31"}`)
	mustErr("inbox-placement", "reports", "list", "--query", "nope")
	mustErr("inbox-placement", "reports", "list", "--limit", "1") // missing required test_id
	mustOK("inbox-placement", "reports", "list", "--test-id", "tid", "--limit", "1", "--starting-after", "c", "--query", "x=1")
	mustErr("inbox-placement", "reports", "get", " ")
	mustOK("inbox-placement", "reports", "get", "rid")

	// oauth
	mustErr("oauth", "google-init", "--data-json", "[]")
	mustOK("oauth", "google-init", "--data-json", `{}`)
	mustErr("oauth", "microsoft-init", "--data-json", "[]")
	mustOK("oauth", "microsoft-init", "--data-json", `{}`)
	mustErr("oauth", "session-status", " ")
	mustOK("oauth", "session-status", "sid")

	// api keys
	mustErr("api-keys", "list", "--query", "nope")
	mustOK("api-keys", "list", "--limit", "1", "--starting-after", "c", "--query", "x=1")
	mustErr("api-keys", "create", "--confirm")
	mustErr("api-keys", "create")                                  // missing confirm
	mustErr("api-keys", "create", "--confirm", "--data-json", "[") // invalid JSON
	mustOK("api-keys", "create", "--confirm", "--name", "n", "--data-json", `{"x":1}`)
	mustErr("api-keys", "delete", "kid")
	mustErr("api-keys", "delete", "--confirm", " ")
	mustOK("api-keys", "delete", "--confirm", "kid")

	// audit logs
	mustErr("audit-logs", "list", "--query", "nope")
	mustOK("audit-logs", "list", "--limit", "1", "--starting-after", "c", "--query", "x=1")

	// workspaces
	mustOK("workspaces", "current", "get")
	mustErr("workspaces", "current", "update")
	mustErr("workspaces", "current", "update", "--data-json", "[]")
	mustOK("workspaces", "current", "update", "--data-json", `{"name":"n"}`)
	mustErr("workspaces", "create")
	mustErr("workspaces", "create", "--confirm")
	mustErr("workspaces", "create", "--confirm", "--data-json", "[") // invalid JSON
	mustOK("workspaces", "create", "--confirm", "--data-json", `{"name":"n"}`)
	mustErr("workspaces", "change-owner")
	mustErr("workspaces", "change-owner", "--confirm")
	mustErr("workspaces", "change-owner", "--confirm", "--data-json", "[") // invalid JSON
	mustOK("workspaces", "change-owner", "--confirm", "--data-json", `{"owner_id":"o"}`)
	mustOK("workspaces", "whitelabel-domain", "get")
	mustErr("workspaces", "whitelabel-domain", "set")
	mustErr("workspaces", "whitelabel-domain", "set", "--confirm")
	mustErr("workspaces", "whitelabel-domain", "set", "--confirm", "--data-json", "[") // invalid JSON
	mustOK("workspaces", "whitelabel-domain", "set", "--confirm", "--domain", "x.example.com", "--data-json", `{"x":1}`)
	mustErr("workspaces", "whitelabel-domain", "delete")
	mustOK("workspaces", "whitelabel-domain", "delete", "--confirm")

	// members
	mustErr("workspace-members", "list", "--query", "nope")
	mustOK("workspace-members", "list", "--limit", "1", "--starting-after", "c", "--query", "x=1")
	mustErr("workspace-members", "get", " ")
	mustOK("workspace-members", "get", "mid")
	mustErr("workspace-members", "create")
	mustErr("workspace-members", "create", "--data-json", "[]")
	mustOK("workspace-members", "create", "--data-json", `{"email":"a@example.com"}`)
	mustErr("workspace-members", "update", " ")
	mustErr("workspace-members", "update", "mid")
	mustErr("workspace-members", "update", "mid", "--data-json", "[") // invalid JSON
	mustOK("workspace-members", "update", "mid", "--data-json", `{"role":"admin"}`)
	mustErr("workspace-members", "delete", "mid")
	mustErr("workspace-members", "delete", "--confirm", " ")
	mustOK("workspace-members", "delete", "--confirm", "mid")

	// group members
	mustErr("workspace-group-members", "list", "--query", "nope")
	mustOK("workspace-group-members", "list", "--limit", "1", "--starting-after", "c", "--query", "x=1")
	mustErr("workspace-group-members", "get", " ")
	mustOK("workspace-group-members", "get", "gid")
	mustErr("workspace-group-members", "create")
	mustErr("workspace-group-members", "create", "--data-json", "[]")
	mustOK("workspace-group-members", "create", "--data-json", `{"email":"a@example.com"}`)
	mustErr("workspace-group-members", "delete", "gid")
	mustErr("workspace-group-members", "delete", "--confirm", " ")
	mustOK("workspace-group-members", "delete", "--confirm", "gid")
	mustOK("workspace-group-members", "admin")

	// billing
	mustOK("workspace-billing", "plan-details")
	mustOK("workspace-billing", "subscription-details")

	// supersearch enrichment
	mustErr("supersearch-enrichment", "get", " ")
	mustOK("supersearch-enrichment", "get", "rid")
	mustErr("supersearch-enrichment", "history", " ")
	mustOK("supersearch-enrichment", "history", "rid")

	for _, sub := range []string{"create", "run", "ai", "count-leads", "enrich-leads"} {
		mustErr("supersearch-enrichment", sub)
		mustErr("supersearch-enrichment", sub, "--confirm")
		mustErr("supersearch-enrichment", sub, "--confirm", "--data-json", "[") // invalid JSON
		mustOK("supersearch-enrichment", sub, "--confirm", "--data-json", `{"x":1}`)
	}
	// Resource-id shortcut should satisfy minimum body for action endpoints.
	mustOK("supersearch-enrichment", "run", "--confirm", "--resource-id", "rid")
	mustOK("supersearch-enrichment", "run", "--confirm", "--resource-id", "rid", "--data-json", `{"x":1}`)

	mustErr("supersearch-enrichment", "update-settings", "rid") // missing confirm
	mustErr("supersearch-enrichment", "update-settings", "--confirm", " ")
	mustErr("supersearch-enrichment", "update-settings", "--confirm", "rid")
	mustErr("supersearch-enrichment", "update-settings", "--confirm", "rid", "--data-json", "[") // invalid JSON
	mustOK("supersearch-enrichment", "update-settings", "--confirm", "rid", "--data-json", `{"x":1}`)

	// crm + dfy
	mustErr("crm-actions", "phone-numbers", "list", "--query", "nope")
	mustOK("crm-actions", "phone-numbers", "list", "--limit", "1", "--starting-after", "c", "--query", "x=1")
	mustErr("crm-actions", "phone-numbers", "delete", "pid")
	mustErr("crm-actions", "phone-numbers", "delete", "--confirm", " ")
	mustOK("crm-actions", "phone-numbers", "delete", "--confirm", "pid")

	mustErr("dfy-email-account-orders", "list", "--query", "nope")
	mustOK("dfy-email-account-orders", "list", "--limit", "1", "--starting-after", "c", "--query", "x=1")
	mustErr("dfy-email-account-orders", "create")
	mustErr("dfy-email-account-orders", "create", "--confirm")
	mustErr("dfy-email-account-orders", "create", "--confirm", "--data-json", "[") // invalid JSON
	mustOK("dfy-email-account-orders", "create", "--confirm", "--data-json", `{"x":1}`)

	mustErr("dfy-email-account-orders", "accounts", "list", "--query", "nope")
	mustOK("dfy-email-account-orders", "accounts", "list", "--limit", "1", "--starting-after", "c", "--query", "x=1")
	mustErr("dfy-email-account-orders", "accounts", "cancel")
	mustErr("dfy-email-account-orders", "accounts", "cancel", "--confirm")
	mustErr("dfy-email-account-orders", "accounts", "cancel", "--confirm", "--data-json", "[") // invalid JSON
	mustOK("dfy-email-account-orders", "accounts", "cancel", "--confirm", "--data-json", `{"x":1}`)

	for _, sub := range []string{"check", "similar", "pre-warmed-up-list"} {
		mustErr("dfy-email-account-orders", "domains", sub)
		mustErr("dfy-email-account-orders", "domains", sub, "--confirm")
		mustErr("dfy-email-account-orders", "domains", sub, "--confirm", "--data-json", "[") // invalid JSON
		mustOK("dfy-email-account-orders", "domains", sub, "--confirm", "--data-json", `{"x":1}`)
	}
}
