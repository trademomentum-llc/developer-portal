# rr-policy-guards

PreToolUse policy hooks for the Rational Reserve platform.

## What this is

A Claude Code plugin that enforces project-wide policy by validating tool
invocations before they execute. Currently includes one hook:

- **rr-emoji-guard** -- blocks any file write (`Write`, `Edit`, `MultiEdit`)
  whose content contains a non-ASCII character. The rule is absolute: every
  file in this user's projects must be pure ASCII. Em dashes, smart quotes,
  box-drawing characters, arrows, check marks, and emoji are all rejected.

A second hook (rr-brew-guard) is planned for M1 of the developer-portal
build and will live alongside the emoji guard in this same plugin.

## Layout

```
plugins/rr-policy-guards/
+-- plugin.json                      plugin manifest
+-- hooks/
|   +-- hooks.json                   plugin hook configuration
+-- tools/
|   +-- emoji-guard/
|       +-- go.mod                   stdlib only
|       +-- main.go                  PreToolUse entrypoint
|       +-- parser.go                ASCII scanner + content extractor
|       +-- audit.go                 JSONL audit log writer
|       +-- main_test.go             integration tests
|       +-- parser_test.go           pure-function unit tests
+-- bin/
    +-- rr-emoji-guard               compiled binary (gitignored)
```

## Build

```
cd tools/emoji-guard
go test ./...
go build -o ../../bin/rr-emoji-guard .
```

The binary is a single static Go executable with no external dependencies.

## How it is wired into Claude Code

The hook is registered directly in `~/.claude/settings.json`:

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Write|Edit|MultiEdit",
        "hooks": [
          {
            "type": "command",
            "command": "/Users/nnos/Projects/developer-portal/plugins/rr-policy-guards/bin/rr-emoji-guard",
            "timeout": 5
          }
        ]
      }
    ]
  }
}
```

The plugin's own `hooks/hooks.json` exists for future packaging as a
marketplace plugin, but Claude Code currently picks up local plugins via
direct settings.json entries.

## Behavior

For each Write / Edit / MultiEdit invocation:

1. Read the PreToolUse JSON from stdin
2. If the tool is not Write / Edit / MultiEdit, exit 0 (out of scope)
3. Extract the content being written based on the tool name
4. Scan rune by rune for any non-ASCII byte (> 0x7F) or invalid UTF-8
5. If a violation is found, print a clear error to stderr including line,
   column, code point, and category, and exit 2 (block)
6. If no violation, exit 0 (allow)
7. Either way, append a JSON line to the audit log

## Override

Set `RR_EMOJI_GUARD_BYPASS=1` for a single command to allow non-ASCII through.
The bypass is still recorded in the audit log so any usage is auditable.

## Audit log

Default location: `~/.rational-reserve/logs/emoji-guard.jsonl`
Override: `RR_EMOJI_GUARD_AUDIT_LOG=/path/to/log.jsonl`

Each line is one JSON object: `{"ts","action","reason","tool","session"}`.
The log is append-only, mode 0600, never rotated automatically.

## Restart required

Claude Code loads hooks at session start. After registering or updating
this hook you must exit Claude Code and start a new session for the change
to take effect. The current session is not protected by the hook.

## Testing

```
go test ./tools/emoji-guard/...
```

Unit tests cover Scan, ExtractContent, and run() with table-driven cases for
clean ASCII, em dash, box drawing, smart quotes, emoji, smart quotes, line
and column tracking, invalid UTF-8, bypass, malformed input, and audit log
writing.

## License

MIT.
