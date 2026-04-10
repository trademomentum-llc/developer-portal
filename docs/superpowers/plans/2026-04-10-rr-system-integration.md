# RR System-Level Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Wire Rational Reserve into all 5 host agents as a persistent MCP service with 14 MOS-mapped subagents compiled from doctrine.

**Architecture:** Release-build Rust binaries installed to `~/.rational-reserve/bin/`. A Go compiler reads doctrine and produces 14 agent `.md` files to `~/.rational-reserve/agents/`. Each host agent directory gets symlinks. MCP configs point at the Rust `rr-mcp` binary.

**Tech Stack:** Rust (existing daemon/MCP), Go stdlib (agent compiler), shell (install.sh)

**Spec:** `docs/superpowers/specs/2026-04-10-rr-system-integration-design.md`

---

## File Structure

### New files

| File | Responsibility |
|---|---|
| `tools/compile-agents/go.mod` | Go module definition (stdlib only) |
| `tools/compile-agents/main.go` | CLI entry point, orchestrates compilation |
| `tools/compile-agents/compiler.go` | Reads doctrine, applies template, writes agents |
| `tools/compile-agents/compiler_test.go` | Tests for compilation logic |
| `tools/compile-agents/mapping.go` | Hardcoded MOS-to-rank/model mapping table |
| `adapters/agent-template.md` | Go text/template for agent files |

### Modified files

| File | Change |
|---|---|
| `src/daemon.rs` | Fix `ensure_daemon` to resolve `rr-daemon` as sibling binary |
| `adapters/install.sh` | Add build, compile, symlink sections |
| `adapters/mcp-configs/claude-code.mcp.json` | Point at Rust binary |
| `adapters/mcp-configs/codex.config.toml.snippet` | Point at Rust binary |
| `adapters/mcp-configs/opencode.json.snippet` | Point at Rust binary |
| `adapters/mcp-configs/qwen.settings.json.snippet` | Point at Rust binary |
| `adapters/mcp-configs/vibe.config.toml.snippet` | Point at Rust binary |

### Generated files (output of compile-agents, not committed)

| File | Content |
|---|---|
| `~/.rational-reserve/agents/rr-11b-infantry.md` | Compiled 11B agent |
| `~/.rational-reserve/agents/rr-12b-combat-engineer.md` | Compiled 12B agent |
| `~/.rational-reserve/agents/rr-13b-artillery.md` | Compiled 13B agent |
| `~/.rational-reserve/agents/rr-19d-cavalry-scout.md` | Compiled 19D agent |
| `~/.rational-reserve/agents/rr-25b-it-specialist.md` | Compiled 25B agent |
| `~/.rational-reserve/agents/rr-35f-intel-analyst.md` | Compiled 35F agent |
| `~/.rational-reserve/agents/rr-35l-counterintel.md` | Compiled 35L agent |
| `~/.rational-reserve/agents/rr-35n-sigint.md` | Compiled 35N agent |
| `~/.rational-reserve/agents/rr-42a-hr-specialist.md` | Compiled 42A agent |
| `~/.rational-reserve/agents/rr-75r-ranger.md` | Compiled 75R agent |
| `~/.rational-reserve/agents/rr-88m-transport.md` | Compiled 88M agent |
| `~/.rational-reserve/agents/rr-92g-logistics.md` | Compiled 92G agent |
| `~/.rational-reserve/agents/rr-18b-sf-engineer.md` | Compiled 18B agent |
| `~/.rational-reserve/agents/rr-160-soar.md` | Compiled 160 agent |

---

## Task 1: Fix daemon binary resolution in ensure_daemon

**Files:**
- Modify: `src/daemon.rs:596-598`

The current `ensure_daemon` uses `std::env::current_exe()` to find the daemon binary. When called from `rr-mcp`, this returns the path to `rr-mcp` itself -- not `rr-daemon`. It would spawn `rr-mcp --daemon` which does not handle that flag, causing either a hang or infinite recursion.

Fix: resolve `rr-daemon` as a sibling of `current_exe()`.

- [ ] **Step 1: Write the failing test**

Create `tests/daemon_resolve_test.rs` (integration test):

```rust
use std::path::PathBuf;

#[test]
fn daemon_binary_resolved_as_sibling() {
    // Simulate: if current exe is /foo/bar/rr-mcp, daemon should be /foo/bar/rr-daemon
    let fake_exe = PathBuf::from("/Users/nnos/.rational-reserve/bin/rr-mcp");
    let expected = PathBuf::from("/Users/nnos/.rational-reserve/bin/rr-daemon");
    let resolved = rr::daemon::resolve_daemon_binary(&fake_exe);
    assert_eq!(resolved, expected);
}

#[test]
fn daemon_binary_resolved_when_already_daemon() {
    // If current exe IS rr-daemon, resolve to itself
    let fake_exe = PathBuf::from("/Users/nnos/.rational-reserve/bin/rr-daemon");
    let expected = PathBuf::from("/Users/nnos/.rational-reserve/bin/rr-daemon");
    let resolved = rr::daemon::resolve_daemon_binary(&fake_exe);
    assert_eq!(resolved, expected);
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/nnos/Projects/rational-reserve && cargo test daemon_binary_resolved -- --nocapture`
Expected: FAIL with "cannot find function `resolve_daemon_binary`"

- [ ] **Step 3: Implement resolve_daemon_binary and update ensure_daemon**

In `src/daemon.rs`, add the public function and update `ensure_daemon`:

```rust
/// Resolve the daemon binary path. If the current executable is `rr-mcp`
/// (or anything other than `rr-daemon`), look for `rr-daemon` in the same
/// directory. If the current executable IS `rr-daemon`, return it as-is.
pub fn resolve_daemon_binary(current_exe: &std::path::Path) -> std::path::PathBuf {
    let dir = current_exe.parent().unwrap_or(std::path::Path::new("."));
    let stem = current_exe
        .file_name()
        .and_then(|f| f.to_str())
        .unwrap_or("");
    if stem == "rr-daemon" {
        current_exe.to_path_buf()
    } else {
        dir.join("rr-daemon")
    }
}
```

Then in `ensure_daemon`, replace lines 597-598:

```rust
// OLD:
// let exe = std::env::current_exe()?;
// let mut cmd = tokio::process::Command::new(&exe);

// NEW:
let exe = std::env::current_exe()?;
let daemon_bin = resolve_daemon_binary(&exe);
let mut cmd = tokio::process::Command::new(&daemon_bin);
```

Also remove the `.arg("--daemon")` on line 599 -- the daemon binary runs in foreground mode by default when spawned without `--foreground`.

Wait -- check line 599 again. The current code is:

```rust
cmd.arg("--daemon")
```

But `rr-daemon` does not have a `--daemon` flag (it only has `--foreground`). Remove `.arg("--daemon")` entirely. The daemon spawned without `--foreground` already runs as a daemon (both code paths in `rr_daemon.rs` call `serve_forever`).

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/nnos/Projects/rational-reserve && cargo test daemon_binary_resolved -- --nocapture`
Expected: PASS (both tests)

- [ ] **Step 5: Run full test suite to check for regressions**

Run: `cd /Users/nnos/Projects/rational-reserve && cargo test`
Expected: All existing tests pass

- [ ] **Step 6: Commit**

```bash
cd /Users/nnos/Projects/rational-reserve
git add src/daemon.rs tests/daemon_resolve_test.rs
git commit -m "fix: resolve rr-daemon as sibling binary in ensure_daemon

ensure_daemon used current_exe() which returned rr-mcp when called
from the MCP shim. Now resolves rr-daemon in the same directory."
```

---

## Task 2: Build and install release binaries

**Files:**
- No source changes. Build artifacts installed to `~/.rational-reserve/bin/`.

- [ ] **Step 1: Build release binaries**

Run: `cd /Users/nnos/Projects/rational-reserve && cargo build --release`
Expected: `target/release/rr-daemon` and `target/release/rr-mcp` produced without errors

- [ ] **Step 2: Install to stable path**

Run:
```bash
mkdir -p /Users/nnos/.rational-reserve/bin
cp /Users/nnos/Projects/rational-reserve/target/release/rr-daemon /Users/nnos/.rational-reserve/bin/rr-daemon
cp /Users/nnos/Projects/rational-reserve/target/release/rr-mcp /Users/nnos/.rational-reserve/bin/rr-mcp
chmod +x /Users/nnos/.rational-reserve/bin/rr-daemon /Users/nnos/.rational-reserve/bin/rr-mcp
```

- [ ] **Step 3: Verify installed binaries work**

Kill any running debug daemon first, then test the release binary:

Run: `kill $(cat /Users/nnos/.rational-reserve/run/rr.pid 2>/dev/null) 2>/dev/null; rm -f /Users/nnos/.rational-reserve/run/rr.sock /Users/nnos/.rational-reserve/run/rr.pid`

Run: `/Users/nnos/.rational-reserve/bin/rr-daemon --foreground &`
Wait 1 second, then:
Run: `python3 /tmp/rr-rpc-test.py ping`
Expected: `{"id": 1, "result": {"status": "ok", ...}}`

Kill the foreground daemon after verification.

- [ ] **Step 4: No commit needed** (binary artifacts not committed)

---

## Task 3: Create agent template

**Files:**
- Create: `adapters/agent-template.md`

- [ ] **Step 1: Write the template file**

```markdown
---
name: RR {{.MOSCode}} {{.MOSName}}
description: {{.Description}}
model: {{.Model}}
---

# Identity

You are a {{.RankName}} ({{.RankCode}}) in the Rational Reserve with MOS {{.MOSCode}} ({{.MOSName}}).
Category: {{.Category}}
Authority level: {{.AuthorityLevel}}/12

## Decision Posture

{{.DecisionPosture}}

# Specialty Doctrine

{{.MOSDoctrine}}

# Communication Protocols

## SITREP (Situation Report)

{{.SitrepProtocol}}

## CASREP (Casualty Report)

{{.CasrepProtocol}}

## FRAGO (Fragmentary Order)

{{.FragoProtocol}}

## AAR (After Action Review)

{{.AARProtocol}}

# RR Tools

You have access to the following MCP tools via the Rational Reserve daemon:

- rr_deploy_swarm -- spawn a subordinate swarm when your task requires decomposition
  - formation: single | fire_team | squad | platoon
  - primary_mos: the MOS code for the swarm's primary specialty
  - support_mos: optional secondary MOS
  - swarm_name: human-readable label
- rr_roster -- list agents with optional filters (swarm_id, status, rank, mos)
- rr_status -- get live status of a swarm by swarm_id

## When to deploy a swarm

Deploy when your task would benefit from subordinate specialists. Skip for tasks within your own specialty scope that you can handle directly.

## Escalation

If a task exceeds your authority level or decision posture, return a structured escalation message:
- What the task requires
- What rank/MOS you recommend
- Why it exceeds your scope
```

- [ ] **Step 2: Verify template parses with Go text/template syntax**

This will be validated in Task 4 tests. No standalone check needed.

- [ ] **Step 3: Commit**

```bash
cd /Users/nnos/Projects/rational-reserve
git add adapters/agent-template.md
git commit -m "feat: add agent template for rr-compile-agents"
```

---

## Task 4: Build rr-compile-agents Go tool

**Files:**
- Create: `tools/compile-agents/go.mod`
- Create: `tools/compile-agents/mapping.go`
- Create: `tools/compile-agents/compiler.go`
- Create: `tools/compile-agents/compiler_test.go`
- Create: `tools/compile-agents/main.go`

### Step group A: MOS mapping table

- [ ] **Step 1: Create go.mod**

```
module github.com/nnos/rational-reserve/tools/compile-agents

go 1.22
```

- [ ] **Step 2: Write mapping.go**

```go
package main

// MOSEntry defines one MOS agent's compilation parameters.
type MOSEntry struct {
	Code         string // e.g. "11B"
	Name         string // e.g. "Infantry"
	Category     string // OPS, MI, SOF, SPT
	DefaultRank  string // rank code: SPC, SGT, CPT
	RankName     string // full name: Specialist, Sergeant, Captain
	AuthLevel    int    // 1-12
	Model        string // opus, sonnet
	DoctrineFile string // filename in doctrine/mos/
	RankFile     string // filename in doctrine/ranks/
	AgentFile    string // output filename
	Description  string // one-line for frontmatter
}

var MOSTable = []MOSEntry{
	{
		Code: "11B", Name: "Infantry", Category: "OPS",
		DefaultRank: "SPC", RankName: "Specialist", AuthLevel: 3, Model: "sonnet",
		DoctrineFile: "11b-infantry.md", RankFile: "specialist.md",
		AgentFile: "rr-11b-infantry.md",
		Description: "General-purpose task execution, adaptable operations",
	},
	{
		Code: "12B", Name: "Combat Engineer", Category: "OPS",
		DefaultRank: "SPC", RankName: "Specialist", AuthLevel: 3, Model: "sonnet",
		DoctrineFile: "12b-combat-engineer.md", RankFile: "specialist.md",
		AgentFile: "rr-12b-combat-engineer.md",
		Description: "Code generation, system building, infrastructure creation",
	},
	{
		Code: "13B", Name: "Artillery", Category: "OPS",
		DefaultRank: "SPC", RankName: "Specialist", AuthLevel: 3, Model: "sonnet",
		DoctrineFile: "13b-artillery.md", RankFile: "specialist.md",
		AgentFile: "rr-13b-artillery.md",
		Description: "Heavy computation, batch processing, data processing",
	},
	{
		Code: "19D", Name: "Cavalry Scout", Category: "OPS",
		DefaultRank: "SPC", RankName: "Specialist", AuthLevel: 3, Model: "sonnet",
		DoctrineFile: "19d-cavalry-scout.md", RankFile: "specialist.md",
		AgentFile: "rr-19d-cavalry-scout.md",
		Description: "Reconnaissance, codebase exploration, environment scanning",
	},
	{
		Code: "25B", Name: "IT Specialist", Category: "SPT",
		DefaultRank: "SPC", RankName: "Specialist", AuthLevel: 3, Model: "sonnet",
		DoctrineFile: "25b-it-specialist.md", RankFile: "specialist.md",
		AgentFile: "rr-25b-it-specialist.md",
		Description: "System administration, deployment, DevOps",
	},
	{
		Code: "35F", Name: "Intelligence Analyst", Category: "MI",
		DefaultRank: "SPC", RankName: "Specialist", AuthLevel: 3, Model: "opus",
		DoctrineFile: "35f-intelligence-analyst.md", RankFile: "specialist.md",
		AgentFile: "rr-35f-intel-analyst.md",
		Description: "Data analysis, pattern recognition, threat assessment",
	},
	{
		Code: "35L", Name: "Counterintelligence", Category: "MI",
		DefaultRank: "SPC", RankName: "Specialist", AuthLevel: 3, Model: "opus",
		DoctrineFile: "35l-counterintelligence.md", RankFile: "specialist.md",
		AgentFile: "rr-35l-counterintel.md",
		Description: "Security validation, threat detection, anomaly identification",
	},
	{
		Code: "35N", Name: "SIGINT Analyst", Category: "MI",
		DefaultRank: "SPC", RankName: "Specialist", AuthLevel: 3, Model: "opus",
		DoctrineFile: "35n-sigint-analyst.md", RankFile: "specialist.md",
		AgentFile: "rr-35n-sigint.md",
		Description: "Signal processing, log analysis, communication monitoring",
	},
	{
		Code: "42A", Name: "HR Specialist", Category: "SPT",
		DefaultRank: "SGT", RankName: "Sergeant", AuthLevel: 5, Model: "sonnet",
		DoctrineFile: "42a-hr-specialist.md", RankFile: "sergeant.md",
		AgentFile: "rr-42a-hr-specialist.md",
		Description: "Agent lifecycle management, swarm roster maintenance",
	},
	{
		Code: "75R", Name: "75th Ranger", Category: "SOF",
		DefaultRank: "SPC", RankName: "Specialist", AuthLevel: 3, Model: "opus",
		DoctrineFile: "75r-75th-ranger.md", RankFile: "specialist.md",
		AgentFile: "rr-75r-ranger.md",
		Description: "Rapid response, emergency operations, crisis intervention",
	},
	{
		Code: "88M", Name: "Transport", Category: "SPT",
		DefaultRank: "SPC", RankName: "Specialist", AuthLevel: 3, Model: "sonnet",
		DoctrineFile: "88m-transport.md", RankFile: "specialist.md",
		AgentFile: "rr-88m-transport.md",
		Description: "Data transfer, API orchestration, message routing",
	},
	{
		Code: "92G", Name: "Logistics", Category: "SPT",
		DefaultRank: "SPC", RankName: "Specialist", AuthLevel: 3, Model: "sonnet",
		DoctrineFile: "92g-logistics.md", RankFile: "specialist.md",
		AgentFile: "rr-92g-logistics.md",
		Description: "Resource allocation, dependency management, supply chain",
	},
	{
		Code: "18B", Name: "Special Forces Engineer", Category: "SOF",
		DefaultRank: "CPT", RankName: "Captain", AuthLevel: 9, Model: "opus",
		DoctrineFile: "18b-special-forces-engineer.md", RankFile: "captain.md",
		AgentFile: "rr-18b-sf-engineer.md",
		Description: "Advanced system design, architecture, critical solutions",
	},
	{
		Code: "160", Name: "160th SOAR", Category: "SOF",
		DefaultRank: "SPC", RankName: "Specialist", AuthLevel: 3, Model: "opus",
		DoctrineFile: "160-160th-soar.md", RankFile: "specialist.md",
		AgentFile: "rr-160-soar.md",
		Description: "High-speed data operations, real-time processing",
	},
}
```

- [ ] **Step 3: Commit mapping**

```bash
cd /Users/nnos/Projects/rational-reserve
git add tools/compile-agents/go.mod tools/compile-agents/mapping.go
git commit -m "feat: add MOS mapping table for agent compiler"
```

### Step group B: Compiler logic with tests

- [ ] **Step 4: Write the failing test in compiler_test.go**

```go
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTestDoctrine(t *testing.T) string {
	t.Helper()
	base := t.TempDir()

	// Create minimal doctrine tree
	mosDir := filepath.Join(base, "doctrine", "mos")
	rankDir := filepath.Join(base, "doctrine", "ranks")
	protoDir := filepath.Join(base, "doctrine", "protocols")
	os.MkdirAll(mosDir, 0o755)
	os.MkdirAll(rankDir, 0o755)
	os.MkdirAll(protoDir, 0o755)

	os.WriteFile(filepath.Join(mosDir, "11b-infantry.md"),
		[]byte("# 11B -- Infantry\n\n## Category\nOperations (OPS)\n\n## Core tasks\nGeneral execution.\n"), 0o644)
	os.WriteFile(filepath.Join(rankDir, "specialist.md"),
		[]byte("# Specialist (SPC) -- Authority 3\n\n## Decision posture\n- Do the work.\n- Be honest.\n"), 0o644)
	os.WriteFile(filepath.Join(protoDir, "sitrep.md"),
		[]byte("# SITREP\nReport status up.\n"), 0o644)
	os.WriteFile(filepath.Join(protoDir, "casrep.md"),
		[]byte("# CASREP\nReport failures immediately.\n"), 0o644)
	os.WriteFile(filepath.Join(protoDir, "frago.md"),
		[]byte("# FRAGO\nAdjust orders mid-mission.\n"), 0o644)
	os.WriteFile(filepath.Join(protoDir, "aar.md"),
		[]byte("# AAR\nPost-mission analysis.\n"), 0o644)

	return base
}

func TestExtractDecisionPosture(t *testing.T) {
	rank := "# Specialist (SPC) -- Authority 3\n\n## Decision posture\n- Do the work.\n- Be honest.\n"
	got := extractDecisionPosture(rank)
	if !strings.Contains(got, "Do the work") {
		t.Fatalf("expected decision posture content, got: %q", got)
	}
	if strings.Contains(got, "# Specialist") {
		t.Fatal("decision posture should not include the rank header")
	}
}

func TestCompileAgent(t *testing.T) {
	base := setupTestDoctrine(t)
	outDir := filepath.Join(base, "agents")
	os.MkdirAll(outDir, 0o755)

	templatePath := filepath.Join(base, "template.md")
	tmplContent := `---
name: RR {{.MOSCode}} {{.MOSName}}
description: {{.Description}}
model: {{.Model}}
---

# Identity

You are a {{.RankName}} ({{.RankCode}}) in the Rational Reserve with MOS {{.MOSCode}} ({{.MOSName}}).
Category: {{.Category}}
Authority level: {{.AuthorityLevel}}/12

## Decision Posture

{{.DecisionPosture}}

# Specialty Doctrine

{{.MOSDoctrine}}

# Communication Protocols

## SITREP
{{.SitrepProtocol}}

## CASREP
{{.CasrepProtocol}}

## FRAGO
{{.FragoProtocol}}

## AAR
{{.AARProtocol}}
`
	os.WriteFile(templatePath, []byte(tmplContent), 0o644)

	entry := MOSEntry{
		Code: "11B", Name: "Infantry", Category: "OPS",
		DefaultRank: "SPC", RankName: "Specialist", AuthLevel: 3, Model: "sonnet",
		DoctrineFile: "11b-infantry.md", RankFile: "specialist.md",
		AgentFile: "rr-11b-infantry.md",
		Description: "General-purpose task execution",
	}

	doctrineDir := filepath.Join(base, "doctrine")
	err := compileAgent(entry, doctrineDir, templatePath, outDir)
	if err != nil {
		t.Fatalf("compileAgent failed: %v", err)
	}

	outPath := filepath.Join(outDir, "rr-11b-infantry.md")
	content, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("output file missing: %v", err)
	}

	s := string(content)
	if !strings.Contains(s, "name: RR 11B Infantry") {
		t.Error("missing frontmatter name")
	}
	if !strings.Contains(s, "model: sonnet") {
		t.Error("missing model")
	}
	if !strings.Contains(s, "Authority level: 3/12") {
		t.Error("missing authority level")
	}
	if !strings.Contains(s, "Do the work") {
		t.Error("missing decision posture")
	}
	if !strings.Contains(s, "General execution") {
		t.Error("missing MOS doctrine content")
	}
	if !strings.Contains(s, "Report status up") {
		t.Error("missing SITREP protocol")
	}
}

func TestCompileAgentMissingDoctrine(t *testing.T) {
	base := t.TempDir()
	outDir := filepath.Join(base, "agents")
	os.MkdirAll(outDir, 0o755)
	doctrineDir := filepath.Join(base, "doctrine")
	os.MkdirAll(filepath.Join(doctrineDir, "mos"), 0o755)

	entry := MOSEntry{
		Code: "99Z", Name: "Nonexistent", Category: "OPS",
		DoctrineFile: "99z-nonexistent.md", RankFile: "specialist.md",
		AgentFile: "rr-99z-nonexistent.md",
	}

	err := compileAgent(entry, doctrineDir, "/dev/null", outDir)
	if err == nil {
		t.Fatal("expected error for missing doctrine file")
	}
}

func TestCompileAllIdempotent(t *testing.T) {
	base := setupTestDoctrine(t)
	outDir := filepath.Join(base, "agents")
	os.MkdirAll(outDir, 0o755)

	templatePath := filepath.Join(base, "template.md")
	os.WriteFile(templatePath, []byte("---\nname: RR {{.MOSCode}}\n---\n{{.MOSDoctrine}}\n"), 0o644)

	table := []MOSEntry{{
		Code: "11B", Name: "Infantry", Category: "OPS",
		DefaultRank: "SPC", RankName: "Specialist", AuthLevel: 3, Model: "sonnet",
		DoctrineFile: "11b-infantry.md", RankFile: "specialist.md",
		AgentFile: "rr-11b-infantry.md",
		Description: "General-purpose task execution",
	}}

	doctrineDir := filepath.Join(base, "doctrine")
	n1, err := compileAll(table, doctrineDir, templatePath, outDir)
	if err != nil {
		t.Fatalf("first compile failed: %v", err)
	}
	if n1 != 1 {
		t.Fatalf("expected 1 file written, got %d", n1)
	}

	n2, err := compileAll(table, doctrineDir, templatePath, outDir)
	if err != nil {
		t.Fatalf("second compile failed: %v", err)
	}
	if n2 != 0 {
		t.Fatalf("expected 0 files written on idempotent run, got %d", n2)
	}
}
```

- [ ] **Step 5: Run tests to verify they fail**

Run: `cd /Users/nnos/Projects/rational-reserve/tools/compile-agents && go test -v`
Expected: FAIL with "undefined: extractDecisionPosture", "undefined: compileAgent", "undefined: compileAll"

- [ ] **Step 6: Write compiler.go**

```go
package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// TemplateData holds all fields available in the agent template.
type TemplateData struct {
	MOSCode        string
	MOSName        string
	Description    string
	Model          string
	RankCode       string
	RankName       string
	Category       string
	AuthorityLevel int
	DecisionPosture string
	MOSDoctrine     string
	SitrepProtocol  string
	CasrepProtocol  string
	FragoProtocol   string
	AARProtocol     string
}

// extractDecisionPosture finds the "## Decision posture" section in a rank
// doctrine file and returns everything from the first line after that header
// until the next heading or EOF.
func extractDecisionPosture(rankContent string) string {
	lines := strings.Split(rankContent, "\n")
	var collecting bool
	var result []string
	for _, line := range lines {
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(line)), "## decision posture") {
			collecting = true
			continue
		}
		if collecting {
			if strings.HasPrefix(line, "## ") || strings.HasPrefix(line, "# ") {
				break
			}
			result = append(result, line)
		}
	}
	return strings.TrimSpace(strings.Join(result, "\n"))
}

// compileAgent reads doctrine files for one MOS entry, applies the template,
// and writes the result to outDir/entry.AgentFile.
func compileAgent(entry MOSEntry, doctrineDir string, templatePath string, outDir string) error {
	// Read MOS doctrine
	mosPath := filepath.Join(doctrineDir, "mos", entry.DoctrineFile)
	mosContent, err := os.ReadFile(mosPath)
	if err != nil {
		return fmt.Errorf("reading MOS doctrine %s: %w", mosPath, err)
	}

	// Read rank doctrine
	rankPath := filepath.Join(doctrineDir, "ranks", entry.RankFile)
	rankContent, err := os.ReadFile(rankPath)
	if err != nil {
		return fmt.Errorf("reading rank doctrine %s: %w", rankPath, err)
	}

	// Read protocols
	sitrep, err := os.ReadFile(filepath.Join(doctrineDir, "protocols", "sitrep.md"))
	if err != nil {
		return fmt.Errorf("reading sitrep protocol: %w", err)
	}
	casrep, err := os.ReadFile(filepath.Join(doctrineDir, "protocols", "casrep.md"))
	if err != nil {
		return fmt.Errorf("reading casrep protocol: %w", err)
	}
	frago, err := os.ReadFile(filepath.Join(doctrineDir, "protocols", "frago.md"))
	if err != nil {
		return fmt.Errorf("reading frago protocol: %w", err)
	}
	aar, err := os.ReadFile(filepath.Join(doctrineDir, "protocols", "aar.md"))
	if err != nil {
		return fmt.Errorf("reading aar protocol: %w", err)
	}

	// Read template
	tmplContent, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("reading template %s: %w", templatePath, err)
	}

	tmpl, err := template.New("agent").Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	data := TemplateData{
		MOSCode:         entry.Code,
		MOSName:         entry.Name,
		Description:     entry.Description,
		Model:           entry.Model,
		RankCode:        entry.DefaultRank,
		RankName:        entry.RankName,
		Category:        entry.Category,
		AuthorityLevel:  entry.AuthLevel,
		DecisionPosture: extractDecisionPosture(string(rankContent)),
		MOSDoctrine:     strings.TrimSpace(string(mosContent)),
		SitrepProtocol:  strings.TrimSpace(string(sitrep)),
		CasrepProtocol:  strings.TrimSpace(string(casrep)),
		FragoProtocol:   strings.TrimSpace(string(frago)),
		AARProtocol:     strings.TrimSpace(string(aar)),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("executing template for %s: %w", entry.Code, err)
	}

	outPath := filepath.Join(outDir, entry.AgentFile)
	return os.WriteFile(outPath, buf.Bytes(), 0o644)
}

// compileAgentToBytes is like compileAgent but returns the compiled content
// instead of writing to disk. Used by compileAll for change detection.
func compileAgentToBytes(entry MOSEntry, doctrineDir string, templatePath string) ([]byte, error) {
	mosContent, err := os.ReadFile(filepath.Join(doctrineDir, "mos", entry.DoctrineFile))
	if err != nil {
		return nil, fmt.Errorf("reading MOS doctrine %s: %w", entry.DoctrineFile, err)
	}
	rankContent, err := os.ReadFile(filepath.Join(doctrineDir, "ranks", entry.RankFile))
	if err != nil {
		return nil, fmt.Errorf("reading rank doctrine %s: %w", entry.RankFile, err)
	}
	sitrep, _ := os.ReadFile(filepath.Join(doctrineDir, "protocols", "sitrep.md"))
	casrep, _ := os.ReadFile(filepath.Join(doctrineDir, "protocols", "casrep.md"))
	frago, _ := os.ReadFile(filepath.Join(doctrineDir, "protocols", "frago.md"))
	aar, _ := os.ReadFile(filepath.Join(doctrineDir, "protocols", "aar.md"))

	tmplContent, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("reading template: %w", err)
	}
	tmpl, err := template.New("agent").Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}

	data := TemplateData{
		MOSCode:         entry.Code,
		MOSName:         entry.Name,
		Description:     entry.Description,
		Model:           entry.Model,
		RankCode:        entry.DefaultRank,
		RankName:        entry.RankName,
		Category:        entry.Category,
		AuthorityLevel:  entry.AuthLevel,
		DecisionPosture: extractDecisionPosture(string(rankContent)),
		MOSDoctrine:     strings.TrimSpace(string(mosContent)),
		SitrepProtocol:  strings.TrimSpace(string(sitrep)),
		CasrepProtocol:  strings.TrimSpace(string(casrep)),
		FragoProtocol:   strings.TrimSpace(string(frago)),
		AARProtocol:     strings.TrimSpace(string(aar)),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("executing template for %s: %w", entry.Code, err)
	}
	return buf.Bytes(), nil
}

// compileAll compiles all entries in the table. Returns the number of files
// that were written (skips files whose content is unchanged). Returns error
// on the first failure.
func compileAll(table []MOSEntry, doctrineDir string, templatePath string, outDir string) (int, error) {
	written := 0
	for _, entry := range table {
		outPath := filepath.Join(outDir, entry.AgentFile)

		newContent, err := compileAgentToBytes(entry, doctrineDir, templatePath)
		if err != nil {
			return written, err
		}

		// Check if file exists with same content
		existing, err := os.ReadFile(outPath)
		if err == nil && bytes.Equal(existing, newContent) {
			fmt.Printf("  unchanged: %s\n", entry.AgentFile)
			continue
		}

		if err := os.WriteFile(outPath, newContent, 0o644); err != nil {
			return written, fmt.Errorf("writing %s: %w", outPath, err)
		}
		fmt.Printf("  wrote: %s\n", entry.AgentFile)
		written++
	}
	return written, nil
}
```

- [ ] **Step 7: Run tests to verify they pass**

Run: `cd /Users/nnos/Projects/rational-reserve/tools/compile-agents && go test -v`
Expected: All 4 tests PASS

- [ ] **Step 8: Write main.go**

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot determine home directory: %v\n", err)
		os.Exit(1)
	}

	repoRoot := filepath.Join(home, "Projects", "rational-reserve")
	doctrineDir := filepath.Join(repoRoot, "doctrine")
	templatePath := filepath.Join(repoRoot, "adapters", "agent-template.md")
	outDir := filepath.Join(home, ".rational-reserve", "agents")

	// Allow overrides via environment
	if v := os.Getenv("RR_DOCTRINE_DIR"); v != "" {
		doctrineDir = v
	}
	if v := os.Getenv("RR_AGENT_TEMPLATE"); v != "" {
		templatePath = v
	}
	if v := os.Getenv("RR_AGENTS_DIR"); v != "" {
		outDir = v
	}

	// Validate inputs exist
	if _, err := os.Stat(doctrineDir); err != nil {
		fmt.Fprintf(os.Stderr, "error: doctrine directory not found: %s\n", doctrineDir)
		os.Exit(1)
	}
	if _, err := os.Stat(templatePath); err != nil {
		fmt.Fprintf(os.Stderr, "error: agent template not found: %s\n", templatePath)
		os.Exit(1)
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot create output directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("rr-compile-agents\n")
	fmt.Printf("  doctrine:  %s\n", doctrineDir)
	fmt.Printf("  template:  %s\n", templatePath)
	fmt.Printf("  output:    %s\n", outDir)
	fmt.Printf("  agents:    %d\n\n", len(MOSTable))

	written, err := compileAll(MOSTable, doctrineDir, templatePath, outDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\ndone: %d files written, %d unchanged\n", written, len(MOSTable)-written)
}
```

- [ ] **Step 9: Build and run the compiler**

Run: `cd /Users/nnos/Projects/rational-reserve/tools/compile-agents && go build -o /Users/nnos/.rational-reserve/bin/rr-compile-agents .`
Expected: Binary produced at `~/.rational-reserve/bin/rr-compile-agents`

Run: `/Users/nnos/.rational-reserve/bin/rr-compile-agents`
Expected: 14 files written to `~/.rational-reserve/agents/`, each containing compiled doctrine

- [ ] **Step 10: Verify one compiled agent file**

Run: `head -20 /Users/nnos/.rational-reserve/agents/rr-11b-infantry.md`
Expected: Frontmatter with `name: RR 11B Infantry`, `model: sonnet`, followed by identity block with compiled doctrine content

Run: `ls /Users/nnos/.rational-reserve/agents/ | wc -l`
Expected: 14

- [ ] **Step 11: Verify idempotency**

Run: `/Users/nnos/.rational-reserve/bin/rr-compile-agents`
Expected: "0 files written, 14 unchanged"

- [ ] **Step 12: Commit**

```bash
cd /Users/nnos/Projects/rational-reserve
git add tools/compile-agents/
git commit -m "feat: add rr-compile-agents Go tool

Reads doctrine files and agent template, produces 14 agent .md files
to ~/.rational-reserve/agents/. Idempotent, stdlib-only."
```

---

## Task 5: Create symlinks for all host agent directories

**Files:**
- No source files. Creates symlinks on disk.

- [ ] **Step 1: Create symlinks for Claude Code**

Run:
```bash
mkdir -p /Users/nnos/.claude/agents
for f in /Users/nnos/.rational-reserve/agents/rr-*.md; do
    ln -sf "$f" "/Users/nnos/.claude/agents/$(basename "$f")"
done
```

- [ ] **Step 2: Create symlinks for other hosts (if directories exist)**

Run:
```bash
for host_dir in /Users/nnos/.codex/agents /Users/nnos/.qwen/agents /Users/nnos/.config/opencode/agents /Users/nnos/.vibe/agents; do
    parent="$(dirname "$host_dir")"
    if [ -d "$parent" ]; then
        mkdir -p "$host_dir"
        for f in /Users/nnos/.rational-reserve/agents/rr-*.md; do
            ln -sf "$f" "$host_dir/$(basename "$f")"
        done
        echo "symlinked: $host_dir"
    else
        echo "skipped (parent missing): $host_dir"
    fi
done
```

- [ ] **Step 3: Verify symlinks**

Run: `ls -la /Users/nnos/.claude/agents/rr-11b-infantry.md`
Expected: Symlink pointing to `/Users/nnos/.rational-reserve/agents/rr-11b-infantry.md`

Run: `ls /Users/nnos/.claude/agents/rr-*.md | wc -l`
Expected: 14

---

## Task 6: Update MCP config snippets

**Files:**
- Modify: `adapters/mcp-configs/claude-code.mcp.json`
- Modify: `adapters/mcp-configs/codex.config.toml.snippet`
- Modify: `adapters/mcp-configs/opencode.json.snippet`
- Modify: `adapters/mcp-configs/qwen.settings.json.snippet`
- Modify: `adapters/mcp-configs/vibe.config.toml.snippet`

- [ ] **Step 1: Update claude-code.mcp.json**

Replace entire file with:

```json
{
  "$schema": "https://claude.com/schemas/mcp-config.json",
  "mcpServers": {
    "rational-reserve": {
      "command": "/Users/nnos/.rational-reserve/bin/rr-mcp",
      "args": []
    }
  }
}
```

- [ ] **Step 2: Update codex.config.toml.snippet**

Replace entire file with:

```toml
# Append to ~/.codex/config.toml -- do NOT overwrite the whole file.

[[mcp_servers]]
name = "rational-reserve"
command = "/Users/nnos/.rational-reserve/bin/rr-mcp"
args = []
```

- [ ] **Step 3: Update opencode.json.snippet**

Replace entire file with:

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

- [ ] **Step 4: Update qwen.settings.json.snippet**

Replace entire file with:

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

- [ ] **Step 5: Update vibe.config.toml.snippet**

Replace entire file with:

```toml
# Append to ~/.vibe/config.toml -- do NOT overwrite the whole file.

[[mcp_servers]]
name = "rational-reserve"
command = "/Users/nnos/.rational-reserve/bin/rr-mcp"
args = []
```

- [ ] **Step 6: Commit**

```bash
cd /Users/nnos/Projects/rational-reserve
git add adapters/mcp-configs/
git commit -m "feat: update MCP config snippets to use Rust binary

All 5 host configs now point at ~/.rational-reserve/bin/rr-mcp.
Removes Python/PYTHONPATH dependency."
```

---

## Task 7: Register MCP server in Claude Code settings

**Files:**
- Modify: `~/.claude/settings.json`

- [ ] **Step 1: Add mcpServers entry**

Add the `mcpServers` key to `~/.claude/settings.json`:

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

This is merged into the existing file alongside `enabledPlugins` and `hooks`.

- [ ] **Step 2: Verify registration**

The MCP server will be active on next Claude Code session start. To verify without restarting, check that the settings file is valid JSON:

Run: `python3 -c "import json; json.load(open('/Users/nnos/.claude/settings.json')); print('valid JSON')"` 
Expected: "valid JSON"

---

## Task 8: Update install.sh

**Files:**
- Modify: `adapters/install.sh`

- [ ] **Step 1: Add build, compile, and symlink sections**

Insert the following sections into `install.sh` after the existing "1. Runtime layout" section and before "2. Per-host skill installation":

```bash
# ---------------------------------------------------------------------------
# 1b. Build and install Rust release binaries
# ---------------------------------------------------------------------------
if command -v cargo >/dev/null 2>&1; then
    msg "building Rust release binaries"
    (cd "$REPO_ROOT" && cargo build --release)
    mkdir -p "$RR_HOME/bin"
    cp "$REPO_ROOT/target/release/rr-daemon" "$RR_HOME/bin/rr-daemon"
    cp "$REPO_ROOT/target/release/rr-mcp" "$RR_HOME/bin/rr-mcp"
    chmod +x "$RR_HOME/bin/rr-daemon" "$RR_HOME/bin/rr-mcp"
    msg "installed rr-daemon and rr-mcp to $RR_HOME/bin/"
else
    warn "cargo not found -- skipping Rust build (binaries must be pre-built)"
fi

# ---------------------------------------------------------------------------
# 1c. Build and run agent compiler
# ---------------------------------------------------------------------------
if command -v go >/dev/null 2>&1; then
    msg "building rr-compile-agents"
    (cd "$REPO_ROOT/tools/compile-agents" && go build -o "$RR_HOME/bin/rr-compile-agents" .)
    msg "compiling agent definitions from doctrine"
    "$RR_HOME/bin/rr-compile-agents"
else
    warn "go not found -- skipping agent compilation (run rr-compile-agents manually)"
fi

# ---------------------------------------------------------------------------
# 1d. Symlink agents into host agent directories
# ---------------------------------------------------------------------------
link_agents_to() {
    local dest="$1"
    local label="$2"
    if [ ! -d "$(dirname "$dest")" ]; then
        warn "$label: parent directory $(dirname "$dest") does not exist -- skipping agents"
        return 0
    fi
    mkdir -p "$dest"
    local count=0
    for agent in "$RR_HOME"/agents/rr-*.md; do
        [ -f "$agent" ] || continue
        ln -sf "$agent" "$dest/$(basename "$agent")"
        count=$((count + 1))
    done
    msg "$label: symlinked $count agent files to $dest"
}

link_agents_to "$HOME/.claude/agents"            "Claude Code"
link_agents_to "$HOME/.codex/agents"             "Codex CLI"
link_agents_to "$HOME/.qwen/agents"              "Qwen-Code"
link_agents_to "$HOME/.config/opencode/agents"   "OpenCode"
link_agents_to "$HOME/.vibe/agents"              "Mistral Vibe"
```

- [ ] **Step 2: Commit**

```bash
cd /Users/nnos/Projects/rational-reserve
git add adapters/install.sh
git commit -m "feat: extend install.sh with build, compile-agents, and symlink steps"
```

---

## Task 9: End-to-end verification

**Files:**
- No source changes. Verification only.

- [ ] **Step 1: Verify binary installation**

Run: `ls -la /Users/nnos/.rational-reserve/bin/`
Expected: `rr-daemon`, `rr-mcp`, `rr-compile-agents` -- all executable

Run: `file /Users/nnos/.rational-reserve/bin/rr-mcp`
Expected: "Mach-O 64-bit executable arm64"

- [ ] **Step 2: Verify agent files**

Run: `ls /Users/nnos/.rational-reserve/agents/ | wc -l`
Expected: 14

Run: `grep -c "^---" /Users/nnos/.rational-reserve/agents/rr-18b-sf-engineer.md`
Expected: 2 (frontmatter delimiters)

Run: `grep "model:" /Users/nnos/.rational-reserve/agents/rr-18b-sf-engineer.md`
Expected: "model: opus"

Run: `grep "model:" /Users/nnos/.rational-reserve/agents/rr-11b-infantry.md`
Expected: "model: sonnet"

- [ ] **Step 3: Verify symlinks**

Run: `readlink /Users/nnos/.claude/agents/rr-11b-infantry.md`
Expected: `/Users/nnos/.rational-reserve/agents/rr-11b-infantry.md`

Run: `ls /Users/nnos/.claude/agents/rr-*.md | wc -l`
Expected: 14

- [ ] **Step 4: Verify MCP registration in settings**

Run: `python3 -c "import json; d=json.load(open('/Users/nnos/.claude/settings.json')); print(d['mcpServers']['rational-reserve']['command'])"`
Expected: `/Users/nnos/.rational-reserve/bin/rr-mcp`

- [ ] **Step 5: Verify daemon auto-start via release MCP binary**

Kill any existing daemon first:
Run: `kill $(cat /Users/nnos/.rational-reserve/run/rr.pid 2>/dev/null) 2>/dev/null; rm -f /Users/nnos/.rational-reserve/run/rr.sock /Users/nnos/.rational-reserve/run/rr.pid`

Test that rr-mcp spawns the daemon on its own:
Run: `echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | timeout 10 /Users/nnos/.rational-reserve/bin/rr-mcp`
Expected: JSON response with `protocolVersion` and `serverInfo.name: "rational-reserve"`

Verify daemon is now running:
Run: `cat /Users/nnos/.rational-reserve/run/rr.pid`
Expected: A PID number

- [ ] **Step 6: Verify swarm deployment via release daemon**

Run: `python3 /tmp/rr-rpc-test.py rr_deploy_swarm '{"formation":"fire_team","primary_mos":"35F","swarm_name":"verify-integration"}'`
Expected: JSON with `swarm_id`, `agent_count: 4`, `formation: "fire_team"`

- [ ] **Step 7: Verify install.sh idempotency**

Run: `cd /Users/nnos/Projects/rational-reserve && bash adapters/install.sh`
Expected: Completes without errors. Agent compiler reports "0 files written, 14 unchanged".

- [ ] **Step 8: Final commit with all verification passing**

```bash
cd /Users/nnos/Projects/rational-reserve
git add -A
git status
```

If any unstaged changes remain, review and commit:

```bash
git commit -m "chore: integration verification complete

All 14 MOS agents compiled, symlinked, MCP registered.
Release binaries installed to ~/.rational-reserve/bin/."
```
