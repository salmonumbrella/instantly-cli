package cmd

import "testing"

func TestCommands_MissingAPIKeyPaths(t *testing.T) {
	// Ensure we do not accidentally use a real key from the environment.
	t.Setenv("INSTANTLY_API_KEY", "")

	base := []string{"--output", "json"}

	cases := [][]string{
		append(append([]string{}, base...), "accounts", "list"),
		append(append([]string{}, base...), "accounts", "get", "a@example.com"),
		append(append([]string{}, base...), "accounts", "create", "--email", "a@example.com"),
		append(append([]string{}, base...), "accounts", "update", "a@example.com", "--daily-limit", "1"),
		append(append([]string{}, base...), "accounts", "warmup-enable", "a@example.com"),
		append(append([]string{}, base...), "accounts", "warmup-disable", "a@example.com"),
		append(append([]string{}, base...), "accounts", "test-vitals", "a@example.com"),
		append(append([]string{}, base...), "accounts", "delete", "--confirm", "a@example.com"),

		append(append([]string{}, base...), "campaigns", "list"),
		append(append([]string{}, base...), "campaigns", "get", "cid"),
		append(append([]string{}, base...), "campaigns", "create", "--name", "n", "--subject", "s", "--body", "b", "--senders", "a@example.com"),
		append(append([]string{}, base...), "campaigns", "update", "cid", "--name", "n"),
		append(append([]string{}, base...), "campaigns", "activate", "cid"),
		append(append([]string{}, base...), "campaigns", "pause", "cid"),
		append(append([]string{}, base...), "campaigns", "delete", "--confirm", "cid"),
		append(append([]string{}, base...), "campaigns", "search-by-contact", "a@example.com"),

		append(append([]string{}, base...), "leads", "list"),
		append(append([]string{}, base...), "leads", "get", "lid"),
		append(append([]string{}, base...), "leads", "create", "--email", "a@example.com"),
		append(append([]string{}, base...), "leads", "update", "lid", "--first-name", "A"),
		append(append([]string{}, base...), "leads", "delete", "--confirm", "lid"),

		append(append([]string{}, base...), "lead-lists", "list"),
		append(append([]string{}, base...), "lead-lists", "get", "lid"),
		append(append([]string{}, base...), "lead-lists", "create", "--name", "n"),
		append(append([]string{}, base...), "lead-lists", "update", "lid", "--name", "n"),
		append(append([]string{}, base...), "lead-lists", "verification-stats", "lid"),
		append(append([]string{}, base...), "lead-lists", "delete", "--confirm", "lid"),

		append(append([]string{}, base...), "emails", "list"),
		append(append([]string{}, base...), "emails", "get", "eid"),
		append(append([]string{}, base...), "emails", "unread-count"),
		append(append([]string{}, base...), "emails", "reply", "--confirm", "--reply-to", "u", "--eaccount", "a@example.com", "--subject", "s", "--text", "t"),
		append(append([]string{}, base...), "emails", "verify", "--email", "a@example.com", "--max-wait", "0s"),
		append(append([]string{}, base...), "emails", "mark-thread-read", "tid"),

		append(append([]string{}, base...), "analytics", "campaign"),
		append(append([]string{}, base...), "analytics", "daily"),
		append(append([]string{}, base...), "analytics", "warmup"),

		append(append([]string{}, base...), "jobs", "list"),
		append(append([]string{}, base...), "jobs", "get", "jid"),

		append(append([]string{}, base...), "api", "get", "/accounts"),
		append(append([]string{}, base...), "api", "delete", "/accounts/a@example.com"),

		// accounts extra
		append(append([]string{}, base...), "accounts", "analytics-daily"),
		append(append([]string{}, base...), "account-campaign-mappings", "get", "a@example.com"),

		// campaigns extra
		append(append([]string{}, base...), "campaigns", "analytics-overview"),
		append(append([]string{}, base...), "campaigns", "analytics-steps"),

		// emails extra
		append(append([]string{}, base...), "emails", "forward", "--confirm", "--data-json", `{"email_id":"eid","to":["a@example.com"]}`),
		append(append([]string{}, base...), "emails", "update", "eid", "--data-json", `{"marked_as_done":true}`),
		append(append([]string{}, base...), "emails", "delete", "--confirm", "eid"),

		// leads extra
		append(append([]string{}, base...), "leads", "bulk-delete", "--confirm", "--query", "campaign_id=cid"),
		append(append([]string{}, base...), "leads", "merge", "--confirm", "--data-json", `{"source_lead_id":"a","target_lead_id":"b"}`),
		append(append([]string{}, base...), "leads", "update-interest-status", "--confirm", "--data-json", `{"lead_id":"a","lt_interest_status":1}`),

		// webhooks
		append(append([]string{}, base...), "webhooks", "list"),
		append(append([]string{}, base...), "webhooks", "get", "wid"),
		append(append([]string{}, base...), "webhooks", "create", "--target-url", "https://example.com/hook", "--event-type", "all_events"),
		append(append([]string{}, base...), "webhooks", "update", "wid", "--data-json", `{"name":"n"}`),
		append(append([]string{}, base...), "webhooks", "delete", "--confirm", "wid"),
		append(append([]string{}, base...), "webhooks", "event-types"),
		append(append([]string{}, base...), "webhooks", "test", "--confirm", "wid"),
		append(append([]string{}, base...), "webhooks", "resume", "wid"),
		append(append([]string{}, base...), "webhooks", "events", "list"),
		append(append([]string{}, base...), "webhooks", "events", "get", "eid"),
		append(append([]string{}, base...), "webhooks", "events", "summary"),
		append(append([]string{}, base...), "webhooks", "events", "summary-by-date"),

		// tags / mappings
		append(append([]string{}, base...), "custom-tags", "list"),
		append(append([]string{}, base...), "custom-tags", "get", "tid"),
		append(append([]string{}, base...), "custom-tags", "create", "--name", "n"),
		append(append([]string{}, base...), "custom-tags", "update", "tid", "--name", "n"),
		append(append([]string{}, base...), "custom-tags", "delete", "--confirm", "tid"),
		append(append([]string{}, base...), "custom-tags", "toggle-resource", "--tag-id", "tid", "--resource-id", "rid", "--resource-type", "lead", "--enabled=false"),
		append(append([]string{}, base...), "custom-tags", "mappings"),

		// block list entries
		append(append([]string{}, base...), "block-list-entries", "list"),
		append(append([]string{}, base...), "block-list-entries", "get", "bid"),
		append(append([]string{}, base...), "block-list-entries", "create", "--data-json", `{"value":"x"}`),
		append(append([]string{}, base...), "block-list-entries", "update", "bid", "--data-json", `{"value":"y"}`),
		append(append([]string{}, base...), "block-list-entries", "delete", "--confirm", "bid"),

		// lead labels
		append(append([]string{}, base...), "lead-labels", "list"),
		append(append([]string{}, base...), "lead-labels", "get", "lid"),
		append(append([]string{}, base...), "lead-labels", "create", "--name", "n"),
		append(append([]string{}, base...), "lead-labels", "update", "lid", "--name", "n"),
		append(append([]string{}, base...), "lead-labels", "delete", "--confirm", "lid"),

		// subsequences
		append(append([]string{}, base...), "subsequences", "list", "--parent-campaign", "cid"),
		append(append([]string{}, base...), "subsequences", "get", "sid"),
		append(append([]string{}, base...), "subsequences", "create", "--data-json", `{"parent_campaign":"cid","name":"n"}`),
		append(append([]string{}, base...), "subsequences", "update", "sid", "--data-json", `{"name":"n"}`),
		append(append([]string{}, base...), "subsequences", "delete", "--confirm", "sid"),
		append(append([]string{}, base...), "subsequences", "pause", "--confirm", "sid"),
		append(append([]string{}, base...), "subsequences", "resume", "--confirm", "sid"),
		append(append([]string{}, base...), "subsequences", "duplicate", "--confirm", "sid"),

		// inbox placement
		append(append([]string{}, base...), "inbox-placement", "tests", "list"),
		append(append([]string{}, base...), "inbox-placement", "tests", "get", "tid"),
		append(append([]string{}, base...), "inbox-placement", "tests", "create", "--confirm", "--data-json", `{"name":"t"}`),
		append(append([]string{}, base...), "inbox-placement", "tests", "update", "tid", "--data-json", `{"name":"t"}`),
		append(append([]string{}, base...), "inbox-placement", "tests", "delete", "--confirm", "tid"),
		append(append([]string{}, base...), "inbox-placement", "tests", "esps"),
		append(append([]string{}, base...), "inbox-placement", "analytics", "list", "--test-id", "tid"),
		append(append([]string{}, base...), "inbox-placement", "analytics", "get", "aid"),
		append(append([]string{}, base...), "inbox-placement", "analytics", "stats-by-test-id", "--test-id", "tid"),
		append(append([]string{}, base...), "inbox-placement", "analytics", "deliverability-insights", "--test-id", "tid"),
		append(append([]string{}, base...), "inbox-placement", "analytics", "stats-by-date", "--test-id", "tid"),
		append(append([]string{}, base...), "inbox-placement", "reports", "list", "--test-id", "tid"),
		append(append([]string{}, base...), "inbox-placement", "reports", "get", "rid"),

		// oauth
		append(append([]string{}, base...), "oauth", "google-init", "--data-json", `{}`),
		append(append([]string{}, base...), "oauth", "microsoft-init", "--data-json", `{}`),
		append(append([]string{}, base...), "oauth", "session-status", "sid"),

		// api keys
		append(append([]string{}, base...), "api-keys", "list"),
		append(append([]string{}, base...), "api-keys", "create", "--confirm", "--name", "n"),
		append(append([]string{}, base...), "api-keys", "delete", "--confirm", "kid"),

		// audit logs
		append(append([]string{}, base...), "audit-logs", "list"),

		// workspaces
		append(append([]string{}, base...), "workspaces", "current", "get"),
		append(append([]string{}, base...), "workspaces", "current", "update", "--data-json", `{"name":"n"}`),
		append(append([]string{}, base...), "workspaces", "create", "--confirm", "--data-json", `{"name":"n"}`),
		append(append([]string{}, base...), "workspaces", "change-owner", "--confirm", "--data-json", `{"owner_id":"o"}`),
		append(append([]string{}, base...), "workspaces", "whitelabel-domain", "get"),
		append(append([]string{}, base...), "workspaces", "whitelabel-domain", "set", "--confirm", "--domain", "x.example.com"),
		append(append([]string{}, base...), "workspaces", "whitelabel-domain", "delete", "--confirm"),

		// members
		append(append([]string{}, base...), "workspace-members", "list"),
		append(append([]string{}, base...), "workspace-members", "get", "mid"),
		append(append([]string{}, base...), "workspace-members", "create", "--data-json", `{"email":"a@example.com"}`),
		append(append([]string{}, base...), "workspace-members", "update", "mid", "--data-json", `{"role":"admin"}`),
		append(append([]string{}, base...), "workspace-members", "delete", "--confirm", "mid"),

		// group members
		append(append([]string{}, base...), "workspace-group-members", "list"),
		append(append([]string{}, base...), "workspace-group-members", "get", "gid"),
		append(append([]string{}, base...), "workspace-group-members", "create", "--data-json", `{"email":"a@example.com"}`),
		append(append([]string{}, base...), "workspace-group-members", "delete", "--confirm", "gid"),
		append(append([]string{}, base...), "workspace-group-members", "admin"),

		// billing
		append(append([]string{}, base...), "workspace-billing", "plan-details"),
		append(append([]string{}, base...), "workspace-billing", "subscription-details"),

		// supersearch enrichment
		append(append([]string{}, base...), "supersearch-enrichment", "get", "rid"),
		append(append([]string{}, base...), "supersearch-enrichment", "history", "rid"),
		append(append([]string{}, base...), "supersearch-enrichment", "create", "--confirm", "--data-json", `{"x":1}`),
		append(append([]string{}, base...), "supersearch-enrichment", "update-settings", "--confirm", "rid", "--data-json", `{"x":1}`),
		append(append([]string{}, base...), "supersearch-enrichment", "run", "--confirm", "--data-json", `{"x":1}`),
		append(append([]string{}, base...), "supersearch-enrichment", "ai", "--confirm", "--data-json", `{"x":1}`),
		append(append([]string{}, base...), "supersearch-enrichment", "count-leads", "--confirm", "--data-json", `{"x":1}`),
		append(append([]string{}, base...), "supersearch-enrichment", "enrich-leads", "--confirm", "--data-json", `{"x":1}`),

		// crm + dfy
		append(append([]string{}, base...), "crm-actions", "phone-numbers", "list"),
		append(append([]string{}, base...), "crm-actions", "phone-numbers", "delete", "--confirm", "pid"),
		append(append([]string{}, base...), "dfy-email-account-orders", "list"),
		append(append([]string{}, base...), "dfy-email-account-orders", "create", "--confirm", "--data-json", `{"x":1}`),
		append(append([]string{}, base...), "dfy-email-account-orders", "accounts", "list"),
		append(append([]string{}, base...), "dfy-email-account-orders", "accounts", "cancel", "--confirm", "--data-json", `{"x":1}`),
		append(append([]string{}, base...), "dfy-email-account-orders", "domains", "check", "--confirm", "--data-json", `{"x":1}`),
	}

	for _, tc := range cases {
		res := execCLI(t, tc...)
		if res.Err == nil {
			t.Fatalf("expected error for %v; stdout=%q", tc, string(res.Stdout))
		}
	}
}
