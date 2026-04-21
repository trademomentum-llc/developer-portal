#!/usr/bin/env bash
set -euo pipefail

SETTINGS="$HOME/.claude/settings.json"
HOOK_CMD="/Users/nnos/Projects/developer-portal/plugins/rr-policy-guards/bin/rr-tofu-guard"

test -f "$SETTINGS" || exit 0

jq --arg cmd "$HOOK_CMD" '
  .hooks.PreToolUse |= map(
    if .matcher == "Bash" then
      .hooks |= map(select(.command != $cmd))
    else
      .
    end
  ) |
  .hooks.PreToolUse |= map(select(.matcher != "Bash" or (.hooks | length > 0)))
' "$SETTINGS" > "$SETTINGS.tmp" && mv "$SETTINGS.tmp" "$SETTINGS"

echo "tofu-guard hook removed from $SETTINGS"
