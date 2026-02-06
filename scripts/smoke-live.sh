#!/usr/bin/env bash
set -euo pipefail

if [[ -z "${INSTANTLY_API_KEY:-}" ]]; then
  echo "INSTANTLY_API_KEY is required" >&2
  exit 1
fi

cd "$(dirname "$0")/.."

BIN="${BIN:-./instantly}"

echo "[smoke] accounts list"
"$BIN" accounts list --limit 1 --output agent >/dev/null

echo "[smoke] campaigns list"
"$BIN" campaigns list --limit 1 --output agent >/dev/null

echo "[smoke] emails unread-count"
"$BIN" emails unread-count --output agent >/dev/null

echo "[smoke] api get /accounts"
"$BIN" api get /accounts --query limit=1 --output agent >/dev/null

echo "[smoke] ok"

