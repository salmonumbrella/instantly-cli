# Instantly CLI

Agent-friendly CLI for `instantly.ai` that calls the Instantly HTTP API directly (`/api/v2`).

## Install

```bash
brew install salmonumbrella/tap/instantly-cli
```

### From source

```bash
make build
./instantly version
```

## Auth

Set your API key via env var (preferred):

```bash
export INSTANTLY_API_KEY="..."
```

Or pass `--api-key` per command.

## Output (agent-first)

Default output is `agent`. You can also use:

- `--output agent`: wraps results in a stable envelope `{kind, items|item|data, meta}`
- `--output jsonl`: newline-delimited JSON
- `--output text`: minimal (mostly `ok`)

### Filtering

Use `--jq` (gojq) to post-process JSON/agent output:

```bash
instantly campaigns list --jq '.items | map(.id)'
```

Or use `--fields` as a shorthand projection:

```bash
instantly campaigns list --fields id,name,status
```

## Safety

Destructive operations require `--confirm` (delete, sending emails, costly actions).

## Dry Run (no network)

`--dry-run` prevents network calls and returns the request that would be made:

```bash
instantly --dry-run campaigns delete 123 --confirm
instantly --dry-run api post /leads --data '{"email":"a@example.com"}'
```

`--dry-run` does not require `INSTANTLY_API_KEY`.

## Escape Hatch

Low-level access for any endpoint:

```bash
instantly api get /accounts --query limit=1
instantly api post /campaigns --data '{"name":"Test"}'
```

## Commands

Top-level groups:

- `accounts`
- `campaigns`
- `leads`
- `lead-lists`
- `emails`
- `analytics`
- `jobs`
- `api` (raw)

Run `instantly help` for the full command/flag list.

## Agent Introspection

Agents can fetch a machine-readable command/flag tree:

```bash
instantly schema --output json
```

## Live Smoke Test

Read-only smoke calls against the real API:

```bash
make smoke-live
```
