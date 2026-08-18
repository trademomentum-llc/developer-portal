package main

import "testing"

func TestExtract_NonGit(t *testing.T) {
	if inv := ExtractCommit("ls -la"); inv.IsCommit {
		t.Errorf("ls should not be IsCommit")
	}
	if inv := ExtractCommit(""); inv.IsCommit {
		t.Errorf("empty should not be IsCommit")
	}
}

func TestExtract_GitNonCommit(t *testing.T) {
	for _, cmd := range []string{"git status", "git diff", "git log --oneline"} {
		if ExtractCommit(cmd).IsCommit {
			t.Errorf("%q should not be IsCommit", cmd)
		}
	}
}

func TestExtract_SimpleCommit(t *testing.T) {
	inv := ExtractCommit(`git commit -m "feat: add"`)
	if !inv.IsCommit {
		t.Fatal("expected IsCommit")
	}
	if len(inv.MessageArgs) != 1 || inv.MessageArgs[0] != "feat: add" {
		t.Errorf("got %v", inv.MessageArgs)
	}
}

func TestExtract_MultipleMessageFlags(t *testing.T) {
	inv := ExtractCommit(`git commit -m "feat: add" -m "body line"`)
	if len(inv.MessageArgs) != 2 {
		t.Errorf("got %v", inv.MessageArgs)
	}
}

func TestExtract_LongFormMessage(t *testing.T) {
	inv := ExtractCommit(`git commit --message="feat: add"`)
	if len(inv.MessageArgs) != 1 || inv.MessageArgs[0] != "feat: add" {
		t.Errorf("got %v", inv.MessageArgs)
	}
}

func TestExtract_FileFlag(t *testing.T) {
	inv := ExtractCommit("git commit -F /tmp/msg.txt")
	if inv.FileArg != "/tmp/msg.txt" {
		t.Errorf("got %q", inv.FileArg)
	}
}

func TestExtract_DashC(t *testing.T) {
	inv := ExtractCommit(`git -C /repo commit -m "feat: x" -m "why"`)
	if inv.RepoDir != "/repo" {
		t.Errorf("got RepoDir = %q", inv.RepoDir)
	}
	if !inv.IsCommit {
		t.Error("expected IsCommit after -C")
	}
}

func TestExtract_Amend(t *testing.T) {
	inv := ExtractCommit("git commit --amend --no-edit")
	if !inv.Amend || !inv.NoEdit {
		t.Errorf("got %+v", inv)
	}

	inv = ExtractCommit("git -C /x commit --amend")
	if !inv.Amend || inv.RepoDir != "/x" {
		t.Errorf("got %+v", inv)
	}

	inv = ExtractCommit(`git commit --amend -m "x"`)
	if !inv.Amend || len(inv.MessageArgs) != 1 || inv.MessageArgs[0] != "x" {
		t.Errorf("got %+v", inv)
	}
}

func TestExtract_EnvVarPrefix(t *testing.T) {
	// Leading env-var assignments must be skipped so we still find `git commit`.
	inv := ExtractCommit(`GIT_AUTHOR_NAME=Test GIT_COMMITTER_NAME=Test git commit -m "feat: x" -m "y"`)
	if !inv.IsCommit {
		t.Error("expected IsCommit after env prefix")
	}
}

func TestExtract_QuotedNewline(t *testing.T) {
	// A literal newline inside double quotes should be preserved as part of
	// the message arg.
	inv := ExtractCommit("git commit -m \"feat: x\nbody\"")
	if len(inv.MessageArgs) != 1 {
		t.Fatalf("got %v", inv.MessageArgs)
	}
	if inv.MessageArgs[0] != "feat: x\nbody" {
		t.Errorf("got %q", inv.MessageArgs[0])
	}
}
