#!/usr/bin/env bash
set -euo pipefail

SETTINGS="$HOME/.claude/settings.json"
HOOK_CMD="/Users/nnos/Projects/developer-portal/plugins/rr-policy-guards/bin/rr-tofu-guard"

test -f "$SETTINGS" || echo '{}' > "$SETTINGS"

jq --arg cmd "$HOOK_CMD" '
  .hooks //= {} |
  .hooks.PreToolUse //= [] |
  if any(.hooks.PreToolUse[]; .matcher == "Bash" and (.hooks[]? | .command == $cmd))
  then .
  elif any(.hooks.PreToolUse[]; .matcher == "Bash")
  then .hooks.PreToolUse |= map(
    if .matcher == "Bash"
    then .hooks += [{"type": "command", "command": $cmd}]
    else .
    end
  )
  else .hooks.PreToolUse += [{
    "matcher": "Bash",
    "hooks": [{"type": "command", "command": $cmd}]
  }]
  end
' "$SETTINGS" > "$SETTINGS.tmp" && mv "$SETTINGS.tmp" "$SETTINGS"

echo "tofu-guard hook merged into $SETTINGS"
