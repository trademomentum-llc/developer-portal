# RR System-Level Integration Design

**Date:** 2026-04-10
**Status:** Draft
**Scope:** Wire Rational Reserve into all 5 host agents as a persistent system-level service with 14 MOS-mapped subagents

---

## 1. Problem Statement

The Rational Reserve daemon and MCP shim are built and tested but not wired into the host agent ecosystem. The existing adapter layer has stale Python-based MCP config templates and no agent definitions. This design adds:

1. Persistent MCP server registration across all 5 host agents
2. 14 subagent definitions (one per MOS) compiled from doctrine
3. A single canonical agent location with symlinks to each host

---

## 2. Architecture

### 2.1 Runtime Layout (additions in bold)

```
~/.rational-reserve/
  bin/
    rr-daemon              <-- release build (NEW)
    rr-mcp                 <-- release build (NEW)
    rr-compile-agents      <-- Go binary (NEW)
  agents/                  <-- canonical agent files (NEW)
    rr-11b-infantry.md
    rr-12b-combat-engineer.md
    rr-13b-artillery.md
    rr-19d-cavalry-scout.md
    rr-25b-it-specialist.md
    rr-35f-intel-analyst.md
    rr-35l-counterintel.md
    rr-35n-sigint.md
    rr-42a-hr-specialist.md
    rr-75r-ranger.md
    rr-88m-transport.md
    rr-92g-logistics.md
    rr-18b-sf-engineer.md
    rr-160-soar.md
  state/
    rr.db
  run/
    rr.pid
    rr.sock
  logs/
    c2.jsonl
  doctrine/                <-- already exists (copied by install.sh)
```

### 2.2 Source Layout (additions)

```
~/Projects/rational-reserve/
  adapters/
    agent-template.md              <-- template for rr-compile-agents (NEW)
    mcp-configs/
      claude-code.mcp.json         <-- updated: Rust binary path
      codex.config.toml.snippet    <-- updated: Rust binary path
      opencode.json.snippet        <-- updated: Rust binary path
      qwen.settings.json.snippet   <-- updated: Rust binary path
      vibe.config.toml.snippet     <-- updated: Rust binary path
    install.sh                     <-- updated: build, compile agents, symlink
  tools/
    compile-agents/                <-- Go source (NEW)
      go.mod
      main.go
      template.go
```

---

## 3. MCP Server Registration

All 5 hosts register the same binary with no environment variables.

### 3.1 Claude Code

Location: `~/.claude/.mcp.json`

```json
{
  "mcpServers": {
    "rational-reserve": {
      "command": "/Users/nnos/.rational-reserve/bin/rr-mcp",
      "args": []
    }
  }
}
```

### 3.2 Codex CLI

Location: `~/.codex/config.toml` (append)

```toml
[[mcp_servers]]
name = "rational-reserve"
command = "/Users/nnos/.rational-reserve/bin/rr-mcp"
args = []
```

### 3.3 Qwen-Code

Location: `~/.qwen/settings.json`

```json
{
  "mcpServers": {
    "rational-reserve": {
      "command": "/Users/nnos/.rational-reserve/bin/rr-mcp",
      "args": []
    }
  }
}
```

### 3.4 OpenCode

Location: `~/.config/opencode/opencode.json`

```json
{
  "mcp": {
    "rational-reserve": {
      "type": "local",
      "command": ["/Users/nnos/.rational-reserve/bin/rr-mcp"]
    }
  }
}
```

### 3.5 Mistral Vibe

Location: `~/.vibe/config.toml` (append)

```toml
[[mcp_servers]]
name = "rational-reserve"
command = "/Users/nnos/.rational-reserve/bin/rr-mcp"
args = []
```

### 3.6 Daemon Auto-Start

The `rr-mcp` binary calls `ensure_daemon()` internally. On first MCP tool call in any session, the daemon starts if not already running. No launchd/systemd service needed. The daemon binary path is resolved by the MCP shim from `~/.rational-reserve/bin/rr-daemon`.

**Open question:** The current Rust `rr-mcp` binary hardcodes the daemon binary lookup. After the release build is installed to `~/.rational-reserve/bin/`, the `ensure_daemon()` function must resolve the daemon binary from that path, not from `target/debug/`. This may require a code change in `src/bin/rr_mcp.rs` or `src/daemon.rs` to use `~/.rational-reserve/bin/rr-daemon` as the daemon path. Verify during implementation.

---

## 4. Agent Definitions

### 4.1 Mapping

14 agents, one per MOS. Each agent is a markdown file with compiled doctrine.

| MOS | Agent file | Category | Default rank | Model |
|---|---|---|---|---|
| 11B | rr-11b-infantry.md | OPS | SPC | sonnet |
| 12B | rr-12b-combat-engineer.md | OPS | SPC | sonnet |
| 13B | rr-13b-artillery.md | OPS | SPC | sonnet |
| 19D | rr-19d-cavalry-scout.md | OPS | SPC | sonnet |
| 25B | rr-25b-it-specialist.md | SPT | SPC | sonnet |
| 35F | rr-35f-intel-analyst.md | MI | SPC | opus |
| 35L | rr-35l-counterintel.md | MI | SPC | opus |
| 35N | rr-35n-sigint.md | MI | SPC | opus |
| 42A | rr-42a-hr-specialist.md | SPT | SGT | sonnet |
| 75R | rr-75r-ranger.md | SOF | SPC | opus |
| 88M | rr-88m-transport.md | SPT | SPC | sonnet |
| 92G | rr-92g-logistics.md | SPT | SPC | sonnet |
| 18B | rr-18b-sf-engineer.md | SOF | CPT | opus |
| 160 | rr-160-soar.md | SOF | SPC | opus |

### 4.2 Agent File Structure

Each compiled agent file contains:

```markdown
---
name: RR {MOS_CODE} {MOS_NAME}
description: {one-line from doctrine}
model: {opus|sonnet}
---

# Identity

You are a {RANK} ({RANK_NAME}) in the Rational Reserve with MOS {MOS_CODE} ({MOS_NAME}).
Category: {OPS|MI|SOF|SPT}
Authority level: {N}/12

## Decision Posture

{Compiled from doctrine/ranks/{rank}.md -- the decision posture section}

# Specialty Doctrine

{Full content of doctrine/mos/{mos_code}.md}

# Communication Protocols

## SITREP (Situation Report)
{Compiled from doctrine/protocols/sitrep.md}

## CASREP (Casualty Report)
{Compiled from doctrine/protocols/casrep.md}

## FRAGO (Fragmentary Order)
{Compiled from doctrine/protocols/frago.md}

## AAR (After Action Review)
{Compiled from doctrine/protocols/aar.md}

# RR Tools

You have access to the following MCP tools via the Rational Reserve daemon:

- `rr_deploy_swarm` -- spawn a subordinate swarm when your task requires decomposition
  - formation: single | fire_team | squad | platoon
  - primary_mos: the MOS code for the swarm's primary specialty
  - support_mos: optional secondary MOS
  - swarm_name: human-readable label
- `rr_roster` -- list agents with optional filters (swarm_id, status, rank, mos)
- `rr_status` -- get live status of a swarm by swarm_id

## When to deploy a swarm

Deploy when your task would benefit from subordinate specialists. Skip for tasks within your own specialty scope that you can handle directly.

## Escalation

If a task exceeds your authority level or decision posture, return a structured escalation message:
- What the task requires
- What rank/MOS you recommend
- Why it exceeds your scope
```

### 4.3 Doctrine Compilation

Content is embedded statically. No runtime file reads.

**Source files per agent:**
1. `doctrine/mos/{mos_code}.md` -- full MOS doctrine
2. `doctrine/ranks/{default_rank}.md` -- rank doctrine (decision posture section only)
3. `doctrine/protocols/*.md` -- all 4 protocol templates (shared across all agents)

**Regeneration:** Run `~/.rational-reserve/bin/rr-compile-agents` after any doctrine change. The binary reads from `~/Projects/rational-reserve/doctrine/` and the template at `adapters/agent-template.md`, writes to `~/.rational-reserve/agents/`.

---

## 5. Symlink Strategy

Each host's agent directory gets symlinks to the canonical files in `~/.rational-reserve/agents/`.

```
~/.claude/agents/rr-*.md           -> ~/.rational-reserve/agents/rr-*.md
~/.codex/agents/rr-*.md            -> ~/.rational-reserve/agents/rr-*.md
~/.qwen/agents/rr-*.md             -> ~/.rational-reserve/agents/rr-*.md
~/.config/opencode/agents/rr-*.md  -> ~/.rational-reserve/agents/rr-*.md
~/.vibe/agents/rr-*.md             -> ~/.rational-reserve/agents/rr-*.md
```

Symlinks are created by `install.sh`. Only created for hosts whose parent directory exists on disk. Non-RR agent files in those directories are untouched.

If a host does not follow symlinks for agent discovery, `install.sh` falls back to a file copy for that host and logs a warning.

---

## 6. Updated install.sh

The existing `install.sh` is extended with these new sections (inserted before the MCP config snippet output):

### 6.1 Build Rust release binaries

```
cd ~/Projects/rational-reserve
cargo build --release
mkdir -p ~/.rational-reserve/bin
cp target/release/rr-daemon ~/.rational-reserve/bin/rr-daemon
cp target/release/rr-mcp ~/.rational-reserve/bin/rr-mcp
```

### 6.2 Build and run agent compiler

```
cd ~/Projects/rational-reserve/tools/compile-agents
go build -o ~/.rational-reserve/bin/rr-compile-agents .
~/.rational-reserve/bin/rr-compile-agents
```

### 6.3 Create agent symlinks

```
for host_dir in ~/.claude/agents ~/.codex/agents ~/.qwen/agents \
                ~/.config/opencode/agents ~/.vibe/agents; do
    if [ -d "$(dirname "$host_dir")" ]; then
        mkdir -p "$host_dir"
        for agent in ~/.rational-reserve/agents/rr-*.md; do
            ln -sf "$agent" "$host_dir/$(basename "$agent")"
        done
    fi
done
```

### 6.4 Updated MCP config snippets

All snippets updated to reference `~/.rational-reserve/bin/rr-mcp` instead of `python -m rr_mcp.server`. No PYTHONPATH needed.

---

## 7. rr-compile-agents Tool

### 7.1 Language and Location

Go binary. Source at `~/Projects/rational-reserve/tools/compile-agents/`. stdlib only (reads files, writes files, applies a text template). Installed to `~/.rational-reserve/bin/rr-compile-agents`.

### 7.2 Inputs

- Doctrine directory: `~/Projects/rational-reserve/doctrine/`
- Agent template: `~/Projects/rational-reserve/adapters/agent-template.md`
- Agent mapping: hardcoded table of 14 MOS entries with default rank and model assignment

### 7.3 Outputs

- 14 `.md` files written to `~/.rational-reserve/agents/`
- Stdout log of which files were written or unchanged

### 7.4 Behavior

- Idempotent -- safe to re-run
- Overwrites existing files only if content changed
- Exits non-zero if any doctrine file is missing
- No network access, no dependencies beyond stdlib

---

## 8. Daemon Binary Resolution

The `rr-mcp` binary must resolve the daemon binary from `~/.rational-reserve/bin/rr-daemon` when calling `ensure_daemon()`. Current behavior may hardcode a relative path or use `PATH` lookup.

**Implementation note:** Check `src/daemon.rs` and `src/bin/rr_mcp.rs` for how the daemon binary path is resolved. If it uses a relative path or `PATH`, add a compile-time or runtime override that defaults to `~/.rational-reserve/bin/rr-daemon`. The `RR_HOME` environment variable (already used by `install.sh`) can serve as the root for path resolution.

---

## 9. Verification Checklist

After implementation, verify:

- [ ] `cargo build --release` succeeds
- [ ] `~/.rational-reserve/bin/rr-mcp` and `rr-daemon` are present and executable
- [ ] `rr-compile-agents` builds and produces 14 files in `~/.rational-reserve/agents/`
- [ ] Each agent file contains compiled doctrine (not file references)
- [ ] Symlinks exist in all host agent directories (for hosts present on disk)
- [ ] `~/.claude/settings.json` mcpServers entry points at the release binary
- [ ] MCP config snippets for all 5 hosts reference the Rust binary
- [ ] Starting a new Claude Code session shows `rational-reserve` as an available MCP server
- [ ] `rr_deploy_swarm` works from a Claude Code session via MCP
- [ ] Agent files are loadable by Claude Code's Agent tool
- [ ] `install.sh` is idempotent (running twice produces same result)
- [ ] Daemon auto-starts when MCP shim is invoked
- [ ] Swarm state is visible across agents (deploy from one, query from another)

---

## 10. Out of Scope

- v0.3/v0.4 MCP tools (execute_mission, SITREP, FRAGO, CASREP, disband) -- future work
- LLM routing within agents -- agents use the model specified in frontmatter, not the rank-based multi-LLM selection from REQUIREMENTS_RR.md Section 9.1 (that is a daemon-level concern for v0.3)
- Agent-to-agent direct communication -- all coordination goes through the daemon or the dispatching parent
- Automated MCP config file editing -- install.sh prints snippets, user merges manually (matches existing install.sh behavior)
