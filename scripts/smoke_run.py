#!/usr/bin/env python3
"""
Real-credentials smoke runner for the Instantly CLI.

Design goals:
- Agent-only: concise, machine-readable summary (JSON report)
- Safe by default: prefer GET/live reads; use --dry-run for risky/costly writes
- Best-effort: continue past failures, report what failed/skipped

Secrets:
- Reads INSTANTLY_API_KEY from environment; never prints it.
"""

from __future__ import annotations

import json
import os
import subprocess
import sys
import time
from dataclasses import dataclass, asdict
from typing import Any, Dict, List, Optional, Tuple


@dataclass
class StepResult:
    name: str
    argv: List[str]
    mode: str  # live|dry_run|skip
    ok: bool
    rc: int
    seconds: float
    note: str = ""
    stdout_json: Any = None
    stderr_tail: str = ""


def _tail(s: str, n: int = 800) -> str:
    s = s.strip()
    if len(s) <= n:
        return s
    return s[-n:]


def run_cli(
    bin_path: str,
    argv: List[str],
    env: Dict[str, str],
    parse_json: bool = False,
    timeout_s: int = 90,
) -> Tuple[bool, int, float, Any, str]:
    start = time.time()
    p = subprocess.run(
        [bin_path, *argv],
        env=env,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
        timeout=timeout_s,
    )
    seconds = time.time() - start
    out = p.stdout
    err = p.stderr
    parsed = None
    if parse_json:
        try:
            parsed = json.loads(out) if out.strip() else None
        except Exception:
            parsed = None
    return p.returncode == 0, p.returncode, seconds, parsed, _tail(err)


def first_in(obj: Any, path: List[str], default=None):
    cur = obj
    for k in path:
        if not isinstance(cur, dict) or k not in cur:
            return default
        cur = cur[k]
    return cur


def find_first_id(obj: Any) -> Optional[str]:
    # Heuristic: find first string id-like field named "id".
    if isinstance(obj, dict):
        v = obj.get("id")
        if isinstance(v, str) and v.strip():
            return v.strip()
        for vv in obj.values():
            got = find_first_id(vv)
            if got:
                return got
    elif isinstance(obj, list):
        for it in obj:
            got = find_first_id(it)
            if got:
                return got
    return None


def find_first_email(obj: Any) -> Optional[str]:
    if isinstance(obj, dict):
        for k in ("email", "eaccount"):
            v = obj.get(k)
            if isinstance(v, str) and "@" in v:
                return v
        for vv in obj.values():
            got = find_first_email(vv)
            if got:
                return got
    elif isinstance(obj, list):
        for it in obj:
            got = find_first_email(it)
            if got:
                return got
    return None


def main() -> int:
    repo_root = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
    bin_path = os.environ.get("BIN", os.path.join(repo_root, "instantly"))

    api_key = os.environ.get("INSTANTLY_API_KEY", "").strip()
    if not api_key:
        print("INSTANTLY_API_KEY is required", file=sys.stderr)
        return 2

    env = os.environ.copy()
    # Make parsing consistent.
    env["INSTANTLY_OUTPUT"] = "json"
    # Avoid accidental interactive/costly retries.
    env.setdefault("INSTANTLY_MAX_429_RETRIES", "0")
    env.setdefault("INSTANTLY_MAX_5XX_RETRIES", "0")

    results: List[StepResult] = []

    def step(name: str, argv: List[str], mode: str, parse_json: bool = False, note: str = "") -> Any:
        if mode == "skip":
            results.append(
                StepResult(
                    name=name,
                    argv=argv,
                    mode=mode,
                    ok=True,
                    rc=0,
                    seconds=0.0,
                    note=note,
                )
            )
            return None

        ok, rc, seconds, out_json, err_tail = run_cli(
            bin_path=bin_path,
            argv=argv,
            env=env,
            parse_json=parse_json,
        )
        results.append(
            StepResult(
                name=name,
                argv=argv,
                mode=mode,
                ok=ok,
                rc=rc,
                seconds=seconds,
                note=note,
                stdout_json=out_json if parse_json else None,
                stderr_tail=err_tail,
            )
        )
        return out_json

    # Baseline
    ws = step("workspaces.current.get", ["workspaces", "current", "get"], "live", parse_json=True)
    step("workspace-billing.plan-details", ["workspace-billing", "plan-details"], "live")
    step("workspace-billing.subscription-details", ["workspace-billing", "subscription-details"], "live")
    step("audit-logs.list", ["audit-logs", "list", "--limit", "1"], "live")

    # OAuth (dry-run)
    step("oauth.google-init", ["--dry-run", "oauth", "google-init", "--output", "agent"], "dry_run")
    step("oauth.microsoft-init", ["--dry-run", "oauth", "microsoft-init", "--output", "agent"], "dry_run")
    step("oauth.session-status", ["--dry-run", "oauth", "session-status", "SESSION_ID", "--output", "agent"], "dry_run")

    # Accounts
    accounts = step("accounts.list", ["accounts", "list", "--limit", "1"], "live", parse_json=True)
    sender_email = find_first_email(accounts)
    if sender_email:
        step("accounts.get", ["accounts", "get", sender_email], "live")
        step("accounts.analytics-daily", ["accounts", "analytics-daily", "--limit", "1"], "live")
        step("accounts.warmup-enable", ["--dry-run", "accounts", "warmup-enable", sender_email, "--output", "agent"], "dry_run")
        step("accounts.warmup-disable", ["--dry-run", "accounts", "warmup-disable", sender_email, "--output", "agent"], "dry_run")
        step("accounts.test-vitals", ["--dry-run", "accounts", "test-vitals", sender_email, "--output", "agent"], "dry_run")
    else:
        step("accounts.get", [], "skip", note="no account email found from accounts list")

    # Campaigns + analytics
    campaigns = step("campaigns.list", ["campaigns", "list", "--limit", "1"], "live", parse_json=True)
    campaign_id = find_first_id(campaigns)
    if campaign_id:
        step("campaigns.get", ["campaigns", "get", campaign_id], "live")
        step("campaigns.activate", ["--dry-run", "campaigns", "activate", campaign_id, "--output", "agent"], "dry_run")
        step("campaigns.pause", ["--dry-run", "campaigns", "pause", campaign_id, "--output", "agent"], "dry_run")
        step("campaigns.analytics-overview", ["campaigns", "analytics-overview", "--query", f"campaign_id={campaign_id}"], "live")
        step("campaigns.analytics-steps", ["campaigns", "analytics-steps", "--query", f"campaign_id={campaign_id}"], "live")
    else:
        step("campaigns.get", [], "skip", note="no campaign id found from campaigns list")

    step("campaigns.search-by-contact", ["campaigns", "search-by-contact", "test@example.com"], "live")
    if campaign_id:
        step("analytics.campaign", ["analytics", "campaign", "--campaign-id", campaign_id], "live")
        step("analytics.daily", ["analytics", "daily", "--campaign-id", campaign_id], "live")
    else:
        step("analytics.campaign", ["analytics", "campaign"], "live")
        step("analytics.daily", ["analytics", "daily"], "live")
    if sender_email:
        step("analytics.warmup", ["analytics", "warmup", "--email", sender_email, "--start-date", "2026-01-01", "--end-date", "2026-01-31"], "live")

    # Lead lists lifecycle (live)
    lead_list = step(
        "lead-lists.create",
        ["lead-lists", "create", "--name", f"smoke-{int(time.time())}"],
        "live",
        parse_json=True,
    )
    lead_list_id = find_first_id(lead_list)
    if lead_list_id:
        step("lead-lists.get", ["lead-lists", "get", lead_list_id], "live")
        step("lead-lists.update", ["lead-lists", "update", lead_list_id, "--name", f"smoke-updated-{int(time.time())}"], "live")
        step("lead-lists.verification-stats", ["lead-lists", "verification-stats", lead_list_id], "live")
        step("lead-lists.list", ["lead-lists", "list", "--limit", "1"], "live")
    else:
        step("lead-lists.get", [], "skip", note="lead-lists create did not return an id")

    # Leads lifecycle (live)
    lead_email = f"smoke+{int(time.time())}@example.com"
    lead = step("leads.create", ["leads", "create", "--email", lead_email], "live", parse_json=True)
    lead_id = find_first_id(lead)
    step("leads.list", ["leads", "list", "--limit", "1"], "live")
    if lead_id:
        step("leads.get", ["leads", "get", lead_id], "live")
        step("leads.update", ["leads", "update", lead_id, "--first-name", "Smoke"], "live")
        step("leads.delete", ["leads", "delete", "--confirm", lead_id], "live")
    else:
        step("leads.get", [], "skip", note="leads create did not return an id")

    # Destructive bulk endpoints: dry-run only
    step("leads.bulk-delete", ["--dry-run", "leads", "bulk-delete", "--confirm", "--query", f"email={lead_email}", "--output", "agent"], "dry_run")
    step("leads.merge", ["--dry-run", "leads", "merge", "--confirm", "--data-json", '{"source_lead_id":"A","target_lead_id":"B"}', "--output", "agent"], "dry_run")
    step("leads.update-interest-status", ["--dry-run", "leads", "update-interest-status", "--confirm", "--data-json", '{"lead_id":"A","lt_interest_status":1}', "--output", "agent"], "dry_run")

    # Emails (live reads; write endpoints dry-run)
    emails = step("emails.list", ["emails", "list", "--limit", "1"], "live", parse_json=True)
    step("emails.unread-count", ["emails", "unread-count"], "live")
    email_id = find_first_id(emails)
    if email_id:
        step("emails.get", ["emails", "get", email_id], "live")
        step("emails.update", ["--dry-run", "emails", "update", email_id, "--data-json", '{"marked_as_done":true}', "--output", "agent"], "dry_run")
        step("emails.delete", ["--dry-run", "emails", "delete", "--confirm", email_id, "--output", "agent"], "dry_run")
    else:
        step("emails.get", [], "skip", note="no email id found from emails list")
    if sender_email:
        step("emails.reply", ["--dry-run", "emails", "reply", "--confirm", "--reply-to", "someone@example.com", "--eaccount", sender_email, "--subject", "Smoke", "--text", "Smoke", "--output", "agent"], "dry_run")
    step("emails.forward", ["--dry-run", "emails", "forward", "--confirm", "--data-json", '{"email_id":"EID","to":["someone@example.com"]}', "--output", "agent"], "dry_run")
    step("emails.verify", ["emails", "verify", "--email", "test@example.com", "--max-wait", "0s"], "live")

    # Webhooks lifecycle (live)
    wh = step(
        "webhooks.create",
        ["webhooks", "create", "--target-url", "https://example.com/webhook", "--event-type", "all_events"],
        "live",
        parse_json=True,
    )
    webhook_id = find_first_id(wh)
    step("webhooks.list", ["webhooks", "list", "--limit", "1"], "live")
    if webhook_id:
        step("webhooks.get", ["webhooks", "get", webhook_id], "live")
        step("webhooks.update", ["webhooks", "update", webhook_id, "--data-json", '{"name":"smoke-updated"}'], "live")
        step("webhooks.event-types", ["webhooks", "event-types"], "live")
        step("webhooks.test", ["--dry-run", "webhooks", "test", "--confirm", webhook_id, "--output", "agent"], "dry_run")
        step("webhooks.resume", ["webhooks", "resume", webhook_id], "live")
        step("webhooks.delete", ["webhooks", "delete", "--confirm", webhook_id], "live")
    else:
        step("webhooks.get", [], "skip", note="webhooks create did not return an id")

    # Webhook events (live reads)
    we = step("webhook-events.list", ["webhooks", "events", "list", "--limit", "1"], "live", parse_json=True)
    event_id = find_first_id(we)
    if event_id:
        step("webhook-events.get", ["webhooks", "events", "get", event_id], "live")
    step("webhook-events.summary", ["webhooks", "events", "summary"], "live")
    step("webhook-events.summary-by-date", ["webhooks", "events", "summary-by-date"], "live")

    # Tags + mappings lifecycle (live)
    tag = step("custom-tags.create", ["custom-tags", "create", "--name", f"smoke-{int(time.time())}"], "live", parse_json=True)
    tag_id = find_first_id(tag)
    step("custom-tags.list", ["custom-tags", "list", "--limit", "1"], "live")
    if tag_id:
        step("custom-tags.get", ["custom-tags", "get", tag_id], "live")
        step("custom-tags.update", ["custom-tags", "update", tag_id, "--name", f"smoke-updated-{int(time.time())}"], "live")
        step("custom-tags.mappings", ["custom-tags", "mappings", "--limit", "1"], "live")
        step("custom-tags.toggle-resource", ["--dry-run", "custom-tags", "toggle-resource", "--tag-id", tag_id, "--resource-id", "RID", "--resource-type", "lead", "--enabled=false", "--output", "agent"], "dry_run")
        step("custom-tags.delete", ["custom-tags", "delete", "--confirm", tag_id], "live")
    else:
        step("custom-tags.get", [], "skip", note="custom-tags create did not return an id")

    # Block list entries lifecycle (live)
    bl = step("block-list-entries.create", ["block-list-entries", "create", "--data-json", '{"value":"smoke@example.com"}'], "live", parse_json=True)
    bl_id = find_first_id(bl)
    step("block-list-entries.list", ["block-list-entries", "list", "--limit", "1"], "live")
    if bl_id:
        step("block-list-entries.get", ["block-list-entries", "get", bl_id], "live")
        step("block-list-entries.update", ["block-list-entries", "update", bl_id, "--data-json", '{"value":"smoke2@example.com"}'], "live")
        step("block-list-entries.delete", ["block-list-entries", "delete", "--confirm", bl_id], "live")
    else:
        step("block-list-entries.get", [], "skip", note="block-list-entries create did not return an id")

    # Lead labels lifecycle (live)
    ll = step("lead-labels.create", ["lead-labels", "create", "--name", f"smoke-{int(time.time())}"], "live", parse_json=True)
    label_id = find_first_id(ll)
    step("lead-labels.list", ["lead-labels", "list", "--limit", "1"], "live")
    if label_id:
        step("lead-labels.get", ["lead-labels", "get", label_id], "live")
        step("lead-labels.update", ["lead-labels", "update", label_id, "--name", f"smoke-updated-{int(time.time())}"], "live")
        step("lead-labels.delete", ["lead-labels", "delete", "--confirm", label_id], "live")
    else:
        step("lead-labels.get", [], "skip", note="lead-labels create did not return an id")

    # Subsequences (read if we have a campaign; writes dry-run)
    if campaign_id:
        subs = step("subsequences.list", ["subsequences", "list", "--parent-campaign", campaign_id, "--limit", "1"], "live", parse_json=True)
        sub_id = find_first_id(subs)
        if sub_id:
            step("subsequences.get", ["subsequences", "get", sub_id], "live")
            step("subsequences.pause", ["--dry-run", "subsequences", "pause", sub_id, "--output", "agent"], "dry_run")
            step("subsequences.resume", ["--dry-run", "subsequences", "resume", sub_id, "--output", "agent"], "dry_run")
            step("subsequences.duplicate", ["--dry-run", "subsequences", "duplicate", sub_id, "--output", "agent"], "dry_run")
        else:
            step("subsequences.get", [], "skip", note="no subsequence id found")
    else:
        step("subsequences.list", [], "skip", note="no campaign id available for subsequences")

    # Inbox placement (live reads; writes dry-run)
    ip_tests = step("inbox-placement.tests.list", ["inbox-placement", "tests", "list", "--limit", "1"], "live", parse_json=True)
    test_id = find_first_id(ip_tests)
    step("inbox-placement.tests.esps", ["inbox-placement", "tests", "esps"], "live")
    if test_id:
        step("inbox-placement.tests.get", ["inbox-placement", "tests", "get", test_id], "live")
        step("inbox-placement.analytics.list", ["inbox-placement", "analytics", "list", "--test-id", test_id, "--limit", "1"], "live", parse_json=True)
        step("inbox-placement.reports.list", ["inbox-placement", "reports", "list", "--test-id", test_id, "--limit", "1"], "live", parse_json=True)
    else:
        step("inbox-placement.analytics.list", [], "skip", note="no inbox placement test id available")
        step("inbox-placement.reports.list", [], "skip", note="no inbox placement test id available")

    # These are POST endpoints and require a JSON body with test_id.
    if test_id:
        step("inbox-placement.analytics.stats-by-test-id", ["inbox-placement", "analytics", "stats-by-test-id", "--test-id", test_id], "live")
        step("inbox-placement.analytics.deliverability-insights", ["inbox-placement", "analytics", "deliverability-insights", "--test-id", test_id], "live")
        step(
            "inbox-placement.analytics.stats-by-date",
            [
                "inbox-placement",
                "analytics",
                "stats-by-date",
                "--test-id",
                test_id,
                "--start-date",
                "2026-01-01",
                "--end-date",
                "2026-01-31",
            ],
            "live",
        )
    else:
        step("inbox-placement.analytics.stats-by-test-id", [], "skip", note="no inbox placement test id available")
        step("inbox-placement.analytics.deliverability-insights", [], "skip", note="no inbox placement test id available")
    step("inbox-placement.tests.create", ["--dry-run", "inbox-placement", "tests", "create", "--confirm", "--data-json", '{"name":"smoke"}', "--output", "agent"], "dry_run")

    # Background jobs
    jobs = step("jobs.list", ["jobs", "list", "--limit", "1"], "live", parse_json=True)
    job_id = find_first_id(jobs)
    if job_id:
        step("jobs.get", ["jobs", "get", job_id], "live")

    # CRM actions
    step("crm-actions.phone-numbers.list", ["crm-actions", "phone-numbers", "list", "--limit", "1"], "live")

    # DFY (live reads + dry-run writes)
    step("dfy-email-account-orders.list", ["dfy-email-account-orders", "list", "--limit", "1"], "live")
    step("dfy-email-account-orders.accounts.list", ["dfy-email-account-orders", "accounts", "list", "--limit", "1"], "live")
    step("dfy-email-account-orders.create", ["--dry-run", "dfy-email-account-orders", "create", "--confirm", "--data-json", '{"_smoke":true}', "--output", "agent"], "dry_run")
    step("dfy-email-account-orders.accounts.cancel", ["--dry-run", "dfy-email-account-orders", "accounts", "cancel", "--confirm", "--data-json", '{"_smoke":true}', "--output", "agent"], "dry_run")
    step("dfy-email-account-orders.domains.check", ["--dry-run", "dfy-email-account-orders", "domains", "check", "--confirm", "--data-json", '{"_smoke":true}', "--output", "agent"], "dry_run")
    step("dfy-email-account-orders.domains.similar", ["--dry-run", "dfy-email-account-orders", "domains", "similar", "--confirm", "--data-json", '{"_smoke":true}', "--output", "agent"], "dry_run")
    step("dfy-email-account-orders.domains.pre-warmed-up-list", ["--dry-run", "dfy-email-account-orders", "domains", "pre-warmed-up-list", "--confirm", "--data-json", '{"_smoke":true}', "--output", "agent"], "dry_run")

    # Supersearch enrichment (dry-run only)
    step("supersearch-enrichment.create", ["--dry-run", "supersearch-enrichment", "create", "--confirm", "--data-json", '{"name":"smoke"}', "--output", "agent"], "dry_run")
    step("supersearch-enrichment.run", ["--dry-run", "supersearch-enrichment", "run", "--confirm", "--resource-id", "RID", "--data-json", '{"_smoke":true}', "--output", "agent"], "dry_run")
    step("supersearch-enrichment.ai", ["--dry-run", "supersearch-enrichment", "ai", "--confirm", "--resource-id", "RID", "--data-json", '{"_smoke":true}', "--output", "agent"], "dry_run")
    step("supersearch-enrichment.count-leads", ["--dry-run", "supersearch-enrichment", "count-leads", "--confirm", "--resource-id", "RID", "--data-json", '{"_smoke":true}', "--output", "agent"], "dry_run")
    step("supersearch-enrichment.enrich-leads", ["--dry-run", "supersearch-enrichment", "enrich-leads", "--confirm", "--resource-id", "RID", "--data-json", '{"_smoke":true}', "--output", "agent"], "dry_run")

    # Account campaign mappings
    if sender_email:
        step("account-campaign-mappings.get", ["account-campaign-mappings", "get", sender_email], "live")

    # Workspace members / group members (live reads, write endpoints dry-run)
    members = step("workspace-members.list", ["workspace-members", "list", "--limit", "1"], "live", parse_json=True)
    # Prefer a non-owner member id (some workspaces return 404 when fetching the owner membership).
    member_id = None
    if isinstance(members, dict) and isinstance(members.get("items"), list):
        for it in members["items"]:
            if isinstance(it, dict) and it.get("role") != "owner" and isinstance(it.get("id"), str):
                member_id = it["id"]
                break
    if not member_id:
        member_id = find_first_id(members)
    if member_id:
        step("workspace-members.get", ["workspace-members", "get", member_id], "live")
    step("workspace-group-members.list", ["workspace-group-members", "list", "--limit", "1"], "live", parse_json=True)
    step("workspace-group-members.admin", ["workspace-group-members", "admin"], "live")
    step("workspace-members.create", ["--dry-run", "workspace-members", "create", "--data-json", '{"_smoke":true}', "--output", "agent"], "dry_run")
    step("workspace-members.delete", ["--dry-run", "workspace-members", "delete", "--confirm", "MEMBER_ID", "--output", "agent"], "dry_run")

    # Escape hatch: prove raw works
    step("api.get.accounts", ["api", "get", "/accounts", "--query", "limit=1"], "live")
    step("api.get.workspaces.current", ["api", "get", "/workspaces/current"], "live")

    # Cleanup: lead list
    if lead_list_id:
        step("lead-lists.delete", ["lead-lists", "delete", "--confirm", lead_list_id], "live")

    # Summary + report file
    ok = sum(1 for r in results if r.ok and r.mode != "skip")
    failed = [r for r in results if not r.ok]
    skipped = sum(1 for r in results if r.mode == "skip")

    report = {
        "bin": bin_path,
        "ts": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
        "workspace": ws,
        "counts": {"total": len(results), "ok": ok, "failed": len(failed), "skipped": skipped},
        "results": [asdict(r) for r in results],
    }

    out_path = os.environ.get("SMOKE_REPORT", "/tmp/instantly_smoke_report.json")
    with open(out_path, "w") as f:
        json.dump(report, f, indent=2, sort_keys=True, default=str)

    # Print a small summary to stdout (no secrets).
    print(json.dumps({"report": out_path, "counts": report["counts"]}))

    # Non-zero exit if any failures.
    return 1 if failed else 0


if __name__ == "__main__":
    raise SystemExit(main())
