package main

import "testing"

// Each rule must have at least one positive and one negative test (REQ A2).
// We assert per-rule rather than per-helper so renaming a helper doesn't
// silently drop coverage.

func ruleByCode(t *testing.T, code string) Rule {
	t.Helper()
	for _, r := range Rules() {
		if r.Code == code {
			return r
		}
	}
	t.Fatalf("rule %s not in catalog", code)
	return Rule{}
}

func TestRule_NV_S_001_EnvFile(t *testing.T) {
	r := ruleByCode(t, "NV-S-001")
	pos := []string{".env", "web/.env", ".env.production", "infra/.env.local"}
	neg := []string{".env.example", ".env.sample", "src/env.go", "README.md"}
	for _, p := range pos {
		if !r.Match(p, 0) {
			t.Errorf("expected match for %q", p)
		}
	}
	for _, p := range neg {
		if r.Match(p, 0) {
			t.Errorf("unexpected match for %q", p)
		}
	}
}

func TestRule_NV_S_002_PrivateKey(t *testing.T) {
	r := ruleByCode(t, "NV-S-002")
	pos := []string{"deploy/server.key", "secrets/tls.pem", "infra/id_rsa", "id_ed25519.pub"}
	neg := []string{"docs/keys.md", "src/keyboard.ts"}
	for _, p := range pos {
		if !r.Match(p, 0) {
			t.Errorf("expected match for %q", p)
		}
	}
	for _, p := range neg {
		if r.Match(p, 0) {
			t.Errorf("unexpected match for %q", p)
		}
	}
}

func TestRule_NV_S_003_Credentials(t *testing.T) {
	r := ruleByCode(t, "NV-S-003")
	pos := []string{
		".aws/credentials",
		"home/.aws/credentials",
		"credentials.json",
		"infra/my-service-account.json",
		".htpasswd",
	}
	neg := []string{"README.md", "src/account.go", "service-account.md"}
	for _, p := range pos {
		if !r.Match(p, 0) {
			t.Errorf("expected match for %q", p)
		}
	}
	for _, p := range neg {
		if r.Match(p, 0) {
			t.Errorf("unexpected match for %q", p)
		}
	}
}

func TestRule_NV_R_001_DependencyDir(t *testing.T) {
	r := ruleByCode(t, "NV-R-001")
	pos := []string{"node_modules/react/index.js", "web/node_modules/x.js", "venv/lib/foo.py", ".venv/bin/python", "__pycache__/foo.cpython.pyc", "target/release/app"}
	neg := []string{"vendor/foo.go", "src/node_modules.md", "src/main.go"}
	for _, p := range pos {
		if !r.Match(p, 0) {
			t.Errorf("expected match for %q", p)
		}
	}
	for _, p := range neg {
		if r.Match(p, 0) {
			t.Errorf("unexpected match for %q", p)
		}
	}
}

func TestRule_NV_R_002_CompiledOutput(t *testing.T) {
	r := ruleByCode(t, "NV-R-002")
	pos := []string{"build/app.exe", "lib.so", "bin/foo.dylib", "obj/file.o", "Main.class", "scripts/foo.pyc"}
	neg := []string{"README.md", "src/Exe.go", "docs/dll.md"}
	for _, p := range pos {
		if !r.Match(p, 0) {
			t.Errorf("expected match for %q", p)
		}
	}
	for _, p := range neg {
		if r.Match(p, 0) {
			t.Errorf("unexpected match for %q", p)
		}
	}
}

func TestRule_NV_P_001_BuildDir(t *testing.T) {
	r := ruleByCode(t, "NV-P-001")
	pos := []string{"dist/index.html", "build/main.js", ".next/static/foo.js", ".turbo/foo", "coverage/lcov.info", "out/index.html"}
	neg := []string{"src/dist.ts", "README.md"}
	for _, p := range pos {
		if !r.Match(p, 0) {
			t.Errorf("expected match for %q", p)
		}
	}
	for _, p := range neg {
		if r.Match(p, 0) {
			t.Errorf("unexpected match for %q", p)
		}
	}
}

func TestRule_NV_N_001_EditorScratch(t *testing.T) {
	r := ruleByCode(t, "NV-N-001")
	pos := []string{"src/foo.swp", "main.go~", "config.bak", "patch.orig", "server.log", "npm-debug.log.1"}
	neg := []string{"src/main.go", "README.md", "docs/log.md"}
	for _, p := range pos {
		if !r.Match(p, 0) {
			t.Errorf("expected match for %q", p)
		}
	}
	for _, p := range neg {
		if r.Match(p, 0) {
			t.Errorf("unexpected match for %q", p)
		}
	}
}

func TestRule_NV_N_002_OSJunk(t *testing.T) {
	r := ruleByCode(t, "NV-N-002")
	pos := []string{".DS_Store", "src/.DS_Store", "Thumbs.db", "desktop.ini"}
	neg := []string{"src/main.go", "README.md"}
	for _, p := range pos {
		if !r.Match(p, 0) {
			t.Errorf("expected match for %q", p)
		}
	}
	for _, p := range neg {
		if r.Match(p, 0) {
			t.Errorf("unexpected match for %q", p)
		}
	}
}

func TestRule_NV_N_003_HugeFile(t *testing.T) {
	r := ruleByCode(t, "NV-N-003")
	if !r.Match("video.mp4", 6*1024*1024) {
		t.Error("expected block for 6 MiB file")
	}
	if r.Match("video.mp4", 4*1024*1024) {
		t.Error("unexpected block for 4 MiB file")
	}
}

func TestRule_GR_N_001_LargeFile(t *testing.T) {
	r := ruleByCode(t, "GR-N-001")
	if !r.Match("logo.png", 2*1024*1024) {
		t.Error("expected warn for 2 MiB file")
	}
	if r.Match("tiny.png", 500*1024) {
		t.Error("unexpected warn for 500 KiB file")
	}
	if r.Match("huge.bin", 6*1024*1024) {
		t.Error("warn should NOT fire on 6 MiB (NV-N-003 blocks it instead)")
	}
}

func TestRule_NV_X_001_RuntimeData(t *testing.T) {
	r := ruleByCode(t, "NV-X-001")
	pos := []string{"data/seed.csv", "results/run1.json", "uploads/file.bin"}
	neg := []string{"src/data.go", "docs/data.md", "internal/results/foo.go"}
	for _, p := range pos {
		if !r.Match(p, 0) {
			t.Errorf("expected match for %q", p)
		}
	}
	for _, p := range neg {
		if r.Match(p, 0) {
			t.Errorf("unexpected match for %q", p)
		}
	}
}

func TestRule_GR_P_001_GeneratedProto(t *testing.T) {
	r := ruleByCode(t, "GR-P-001")
	pos := []string{"daemon/proto/foo.pb.go", "api/foo_pb2.py", "rpc/svc_pb2_grpc.py", "x.pb.cc", "x.pb.h"}
	neg := []string{"daemon/proto/foo.proto", "src/main.go"}
	for _, p := range pos {
		if !r.Match(p, 0) {
			t.Errorf("expected match for %q", p)
		}
	}
	for _, p := range neg {
		if r.Match(p, 0) {
			t.Errorf("unexpected match for %q", p)
		}
	}
}

func TestRule_GR_Y_001_IDEConfig(t *testing.T) {
	r := ruleByCode(t, "GR-Y-001")
	pos := []string{".vscode/settings.json", ".idea/workspace.xml"}
	neg := []string{".vscode/extensions.json", ".idea/runConfigurations/foo.xml"}
	for _, p := range pos {
		if !r.Match(p, 0) {
			t.Errorf("expected match for %q", p)
		}
	}
	for _, p := range neg {
		if r.Match(p, 0) {
			t.Errorf("unexpected match for %q", p)
		}
	}
}

func TestSplitFindings(t *testing.T) {
	rs := Rules()
	all := []Finding{
		{Rule: rs[0]}, // block
		{Rule: ruleByCode(t, "GR-P-001")}, // warn
		{Rule: rs[1]}, // block
	}
	b, w := SplitFindings(all)
	if len(b) != 2 || len(w) != 1 {
		t.Fatalf("got %d block / %d warn, want 2 / 1", len(b), len(w))
	}
}

// Smoke test on the entire catalog: every rule has Code, Reason, Match, and
// a Severity of either "block" or "warn".
func TestCatalogShape(t *testing.T) {
	for _, r := range Rules() {
		if r.Code == "" || r.Match == nil || r.Reason == "" {
			t.Errorf("rule %+v is missing required fields", r)
		}
		if r.Severity != "block" && r.Severity != "warn" {
			t.Errorf("rule %s has invalid severity %q", r.Code, r.Severity)
		}
	}
}
