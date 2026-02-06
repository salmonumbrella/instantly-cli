# Instantly CLI — Cold email automation in your terminal.

Instantly.ai in your terminal. Manage sender accounts, campaigns, leads, emails, analytics, webhooks, and workspace settings.

## Features

- **Accounts** - manage sender accounts, warmup, vitals, daily analytics
- **Campaigns** - create, activate, pause, search by contact
- **Leads** - CRUD, bulk delete, merge, interest status
- **Lead Lists** - manage lists and verification stats
- **Emails** - read, reply, forward, verify, mark read
- **Analytics** - campaign, daily, warmup reporting
- **Webhooks** - create, test, resume, event history
- **Inbox Placement** - deliverability tests and reports
- **Workspaces** - manage workspace, members, groups, billing
- **Custom Tags** - tag resources, view mappings
- **Supersearch** - enrichment, AI search, lead counting
- **Agent-first** - stable JSON envelopes, dry-run, jq filtering, schema introspection

## Installation

### Homebrew

```bash
brew install salmonumbrella/tap/instantly-cli
```

### From source

```bash
make build
./instantly version
```

## Quick Start

### 1. Set API Key

```bash
export INSTANTLY_API_KEY="your-key-here"
```

Or pass `--api-key` per command.

### 2. Test It

```bash
instantly accounts list --limit 1
```

## Configuration

### Environment Variables

- `INSTANTLY_API_KEY` - API key (required)
- `INSTANTLY_OUTPUT` - Default output format: `agent` (default), `json`, `jsonl`, `text`

## Rate Limiting

The CLI handles rate limiting automatically:

- **Exponential backoff** - retries with increasing delays plus jitter
- **429 retries** - configurable via `--max-429-retries` (default: 0)
- **5xx retries** - configurable via `--max-5xx-retries` (default: 0)
- **Idempotency keys** - safe write retries via `--idempotency-key`

## Commands

### Accounts

```bash
instantly accounts list [--limit <n>] [--starting-after <cursor>] [--search <text>]
instantly accounts get <email>
instantly accounts create --email <email> [--first-name <name>] [--last-name <name>]
instantly accounts update <email> [--daily-limit <n>] [--sending-gap <n>]
instantly accounts warmup-enable <email>
instantly accounts warmup-disable <email>
instantly accounts test-vitals <email>
instantly accounts delete <email> --confirm
instantly accounts analytics-daily [--limit <n>] [--starting-after <cursor>]
```

### Campaigns

```bash
instantly campaigns list [--limit <n>] [--starting-after <cursor>] [--search <text>]
instantly campaigns get <campaign_id>
instantly campaigns create --name <name> --subject <subject> --body <body> \
  [--senders auto|email1,email2] [--daily-limit <n>] [--email-gap <n>]
instantly campaigns update <campaign_id> [--name <name>] [--daily-limit <n>]
instantly campaigns activate <campaign_id>
instantly campaigns pause <campaign_id>
instantly campaigns delete <campaign_id> --confirm
instantly campaigns search-by-contact <contact_email>
instantly campaigns analytics-overview
instantly campaigns analytics-steps
```

### Leads

```bash
instantly leads list [--limit <n>] [--campaign <id>] [--list-id <id>] [--search <text>] [--status <status>]
instantly leads get <lead_id>
instantly leads create --email <email> [--campaign <id>] [--first-name <name>] [--last-name <name>]
instantly leads update <lead_id> [--first-name <name>] [--last-name <name>]
instantly leads delete <lead_id> --confirm
instantly leads bulk-delete --confirm
instantly leads merge --confirm [--data-json <json>]
instantly leads update-interest-status --confirm [--data-json <json>]
```

### Lead Lists

```bash
instantly lead-lists list [--limit <n>] [--starting-after <cursor>] [--search <text>]
instantly lead-lists get <list_id>
instantly lead-lists create --name <name> [--enrich]
instantly lead-lists update <list_id> [--name <name>]
instantly lead-lists verification-stats <list_id>
instantly lead-lists delete <list_id> --confirm
```

### Emails

```bash
instantly emails list [--limit <n>] [--campaign-id <id>] [--eaccount <email>] [--unread]
instantly emails get <email_id>
instantly emails unread-count
instantly emails reply --reply-to <uuid> --eaccount <email> --subject <subject> --confirm
instantly emails forward --confirm [--data-json <json>]
instantly emails verify --email <email> [--max-wait <duration>]
instantly emails mark-thread-read <thread_id>
instantly emails delete <email_id> --confirm
```

### Analytics

```bash
instantly analytics campaign [--campaign-id <id>] [--start-date YYYY-MM-DD] [--end-date YYYY-MM-DD]
instantly analytics daily [--campaign-id <id>] [--start-date YYYY-MM-DD] [--end-date YYYY-MM-DD]
instantly analytics warmup [--email <email>] [--start-date YYYY-MM-DD] [--end-date YYYY-MM-DD]
```

### Webhooks

```bash
instantly webhooks list [--limit <n>] [--campaign <id>] [--event-type <type>]
instantly webhooks get <webhook_id>
instantly webhooks create [--target-url <url>] [--campaign <id>] [--event-type <type>]
instantly webhooks update <webhook_id> [--data-json <json>]
instantly webhooks delete <webhook_id> --confirm
instantly webhooks event-types
instantly webhooks test <webhook_id> --confirm
instantly webhooks resume <webhook_id>
instantly webhooks events list [--limit <n>]
instantly webhooks events get <event_id>
instantly webhooks events summary
instantly webhooks events summary-by-date
```

### Custom Tags

```bash
instantly custom-tags list [--limit <n>] [--search <text>]
instantly custom-tags get <tag_id>
instantly custom-tags create [--name <name>] [--color <color>]
instantly custom-tags update <tag_id> [--name <name>] [--color <color>]
instantly custom-tags delete <tag_id> --confirm
instantly custom-tags toggle-resource [--tag-id <id>] [--resource-id <id>] [--resource-type <type>]
instantly custom-tags mappings [--limit <n>]
```

### Block List Entries

```bash
instantly block-list-entries list [--limit <n>] [--search <text>]
instantly block-list-entries get <entry_id>
instantly block-list-entries create [--data-json <json>]
instantly block-list-entries update <entry_id> [--data-json <json>]
instantly block-list-entries delete <entry_id> --confirm
```

### Lead Labels

```bash
instantly lead-labels list [--limit <n>] [--search <text>]
instantly lead-labels get <label_id>
instantly lead-labels create --name <name> [--interest-status positive|neutral|negative]
instantly lead-labels update <label_id> [--name <name>]
instantly lead-labels delete <label_id> --confirm
```

### Subsequences

```bash
instantly subsequences list --parent-campaign <id> [--limit <n>]
instantly subsequences get <subsequence_id>
instantly subsequences create [--data-json <json>]
instantly subsequences update <subsequence_id> [--data-json <json>]
instantly subsequences delete <subsequence_id> --confirm
instantly subsequences pause <subsequence_id> --confirm
instantly subsequences resume <subsequence_id> --confirm
instantly subsequences duplicate <subsequence_id> --confirm
```

### Inbox Placement

```bash
instantly inbox-placement tests list [--limit <n>]
instantly inbox-placement tests get <test_id>
instantly inbox-placement tests create --confirm [--data-json <json>]
instantly inbox-placement tests delete <test_id> --confirm
instantly inbox-placement tests esps
instantly inbox-placement analytics list --test-id <id>
instantly inbox-placement analytics get <analytics_id>
instantly inbox-placement analytics stats-by-test-id --test-id <id>
instantly inbox-placement analytics deliverability-insights --test-id <id>
instantly inbox-placement analytics stats-by-date --test-id <id> [--start-date YYYY-MM-DD] [--end-date YYYY-MM-DD]
instantly inbox-placement reports list --test-id <id>
instantly inbox-placement reports get <report_id>
```

### OAuth

```bash
instantly oauth google-init [--data-json <json>]
instantly oauth microsoft-init [--data-json <json>]
instantly oauth session-status <session_id>
```

### API Keys

```bash
instantly api-keys list [--limit <n>]
instantly api-keys create --name <name> --confirm
instantly api-keys delete <api_key_id> --confirm
```

### Workspaces

```bash
instantly workspaces current get
instantly workspaces current update [--data-json <json>]
instantly workspaces create --confirm [--data-json <json>]
instantly workspaces change-owner --confirm [--data-json <json>]
instantly workspaces whitelabel-domain get
instantly workspaces whitelabel-domain set --domain <domain> --confirm
instantly workspaces whitelabel-domain delete --confirm
```

### Workspace Members

```bash
instantly workspace-members list [--limit <n>]
instantly workspace-members get <member_id>
instantly workspace-members create [--data-json <json>]
instantly workspace-members update <member_id> [--data-json <json>]
instantly workspace-members delete <member_id> --confirm
```

### Workspace Group Members

```bash
instantly workspace-group-members list [--limit <n>]
instantly workspace-group-members get <group_member_id>
instantly workspace-group-members create [--data-json <json>]
instantly workspace-group-members delete <group_member_id> --confirm
instantly workspace-group-members admin
```

### Workspace Billing

```bash
instantly workspace-billing plan-details
instantly workspace-billing subscription-details
```

### Account Campaign Mappings

```bash
instantly account-campaign-mappings get <email>
```

### CRM Actions

```bash
instantly crm-actions phone-numbers list [--limit <n>]
instantly crm-actions phone-numbers delete <phone_number_id> --confirm
```

### DFY Email Account Orders

```bash
instantly dfy-email-account-orders list [--limit <n>]
instantly dfy-email-account-orders create --confirm [--data-json <json>]
instantly dfy-email-account-orders accounts list [--limit <n>]
instantly dfy-email-account-orders accounts cancel --confirm
instantly dfy-email-account-orders domains check --confirm [--data-json <json>]
instantly dfy-email-account-orders domains similar --confirm [--data-json <json>]
instantly dfy-email-account-orders domains pre-warmed-up-list --confirm
```

### Supersearch Enrichment

```bash
instantly supersearch-enrichment create --confirm [--data-json <json>]
instantly supersearch-enrichment get <resource_id>
instantly supersearch-enrichment history <resource_id>
instantly supersearch-enrichment update-settings <resource_id> --confirm [--data-json <json>]
instantly supersearch-enrichment run --confirm [--resource-id <id>]
instantly supersearch-enrichment ai --confirm [--resource-id <id>]
instantly supersearch-enrichment count-leads --confirm [--resource-id <id>]
instantly supersearch-enrichment enrich-leads --confirm [--resource-id <id>]
```

### Audit Logs

```bash
instantly audit-logs list [--limit <n>]
```

### Background Jobs

```bash
instantly jobs list [--limit <n>]
instantly jobs get <job_id>
```

### Escape Hatch (Raw API)

```bash
instantly api get <path> [--query key=value]...
instantly api post <path> [--data <json>] [--data-file <file>]
instantly api patch <path> [--data <json>] [--data-file <file>]
instantly api delete <path>
```

### Introspection

```bash
instantly schema                   # Machine-readable command tree
instantly version                  # Version/build info
```

## Output Formats

### Agent (default)

Stable JSON envelopes for automation:

```bash
$ instantly accounts list --limit 1
{
  "kind": "accounts",
  "items": [...],
  "meta": { "has_more": true, "next_cursor": "..." }
}
```

### JSON

```bash
$ instantly accounts list --limit 1 --output json
[{"email": "sender@example.com", "status": 1, ...}]
```

### JSONL

```bash
$ instantly accounts list --output jsonl
{"email": "sender@example.com", ...}
{"email": "sender2@example.com", ...}
```

### Text

```bash
$ instantly version --output text
0.1.0 (abc1234) 2026-02-06T10:00:00Z
```

Data goes to stdout, errors to stderr for clean piping.

## Examples

### List active campaigns

```bash
instantly campaigns list --jq '.items[] | select(.status == "active") | .name'
```

### Create a campaign with sender accounts

```bash
instantly campaigns create \
  --name "Q1 Outreach" \
  --subject "Quick question" \
  --body "Hi {{firstName}}, ..." \
  --senders sender1@example.com,sender2@example.com \
  --daily-limit 50
```

### Check warmup analytics

```bash
instantly analytics warmup \
  --email sender@example.com \
  --start-date 2026-01-01 \
  --end-date 2026-01-31
```

### Automation

```bash
# Dry-run: preview the request without sending
instantly --dry-run campaigns delete abc123 --confirm

# Pipeline: get all campaign IDs
instantly campaigns list --jq '[.items[].id]'

# Field projection shorthand
instantly accounts list --fields email,status,daily_limit

# Agent-friendly schema introspection
instantly schema --output json
```

### Debug Mode

```bash
instantly --debug accounts list --limit 1
# Shows: HTTP method, URL, headers, response status
```

### Dry-Run Mode

Preview mutations without executing:

```bash
instantly --dry-run leads create --email test@example.com --campaign abc123
# Shows the exact HTTP request that would be made
# No API key required
```

### JQ Filtering

```bash
# Extract just email addresses
instantly accounts list --jq '.items[].email'

# Filter by status
instantly campaigns list --jq '.items[] | select(.status == "paused")'

# Count leads per campaign
instantly campaigns list --jq '.items | length'
```

## Global Flags

All commands support these flags:

- `--output <format>`, `-o` - Output format: `agent` (default), `json`, `jsonl`, `text`
- `--json`, `-j` - Shorthand for `--output json`
- `--quiet` - Suppress stderr and text output
- `--silent` - Suppress all output
- `--debug` - Enable debug logging to stderr
- `--timeout <duration>` - HTTP timeout (default: 60s)
- `--base-url <url>` - API base URL
- `--api-key <key>` - API key (or set `INSTANTLY_API_KEY`)
- `--dry-run` - Print request without network call (no API key required)
- `--jq <expr>` - JQ filter expression for JSON/agent output
- `--fields <fields>` - Comma-separated field projection shorthand
- `--max-429-retries <n>` - Max retries for 429 responses (default: 0)
- `--max-5xx-retries <n>` - Max retries for 5xx responses (default: 0)
- `--retry-delay <duration>` - Base retry delay (default: 1s)
- `--max-retry-delay <duration>` - Max retry delay (default: 30s)
- `--idempotency-key <key>` - Idempotency key for safe write retries
- `--help` - Show help for any command
- `--version` - Show version info (via `instantly version`)

## Shell Completions

```bash
# Bash (macOS/Homebrew)
instantly completion bash > $(brew --prefix)/etc/bash_completion.d/instantly

# Zsh
instantly completion zsh > "${fpath[1]}/_instantly"

# Fish
instantly completion fish > ~/.config/fish/completions/instantly.fish

# PowerShell
instantly completion powershell | Out-String | Invoke-Expression
```

## Development

After cloning, install git hooks:

```bash
lefthook install
```

This installs [lefthook](https://github.com/evilmartians/lefthook) pre-commit and pre-push hooks for linting and testing.

## License

MIT

## Links

- [Instantly API Documentation](https://developer.instantly.ai/introduction)
