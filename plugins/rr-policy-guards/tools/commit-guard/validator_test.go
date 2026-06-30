package main

import (
	"strings"
	"testing"
)

func TestVerdict_EmptySubject(t *testing.T) {
	v := VerdictFromMessage([]byte(""))
	if v.Decision != DecisionBlock || v.RuleCode != "IN-M-001" {
		t.Fatalf("got %+v, want block IN-M-001", v)
	}
}

func TestVerdict_SubjectTooLong(t *testing.T) {
	long := "feat: " + strings.Repeat("x", 80)
	v := VerdictFromMessage([]byte(long + "\n\nbody"))
	if v.Decision != DecisionBlock || v.RuleCode != "IN-M-001" {
		t.Fatalf("got %+v, want block IN-M-001", v)
	}
}

func TestVerdict_NonConventional(t *testing.T) {
	v := VerdictFromMessage([]byte("did stuff today\n\nbody line"))
	if v.Decision != DecisionBlock || v.RuleCode != "IN-M-002" {
		t.Fatalf("got %+v, want block IN-M-002", v)
	}
}

func TestVerdict_MissingBody(t *testing.T) {
	v := VerdictFromMessage([]byte("feat: add login"))
	if v.Decision != DecisionBlock || v.RuleCode != "IN-M-003" {
		t.Fatalf("got %+v, want block IN-M-003", v)
	}
}

func TestVerdict_SubjectOnlyWithTrailingBlanks(t *testing.T) {
	// Trailing blank lines without body content should still block.
	v := VerdictFromMessage([]byte("feat: add login\n\n\n"))
	if v.Decision != DecisionBlock || v.RuleCode != "IN-M-003" {
		t.Fatalf("got %+v, want block IN-M-003", v)
	}
}

func TestVerdict_HappyPath(t *testing.T) {
	v := VerdictFromMessage([]byte("feat(auth): add OIDC sign-in\n\nReplaces the mock cookie\nwith a real OIDC flow."))
	if v.Decision != DecisionAllow {
		t.Fatalf("got %+v, want allow", v)
	}
	if v.Subject != "feat(auth): add OIDC sign-in" {
		t.Errorf("subject = %q", v.Subject)
	}
}

func TestVerdict_BangBreakingChange(t *testing.T) {
	// "feat!:" and "feat(scope)!:" are conventional-commits breaking-change forms.
	v := VerdictFromMessage([]byte("feat!: rotate API surface\n\nBreaking change.\nMigration guide in docs."))
	if v.Decision != DecisionAllow {
		t.Fatalf("got %+v, want allow", v)
	}
}

func TestVerdict_MergeExempt(t *testing.T) {
	v := VerdictFromMessage([]byte("Merge branch 'feat/long-named-branch-that-would-exceed-72-chars-easily' into main"))
	if v.Decision != DecisionAllow {
		t.Fatalf("got %+v, want allow (merge exempt)", v)
	}
}

func TestVerdict_RevertExempt(t *testing.T) {
	v := VerdictFromMessage([]byte("Revert \"feat: thing\"\n\nThis reverts commit abc123."))
	if v.Decision != DecisionAllow {
		t.Fatalf("got %+v, want allow (revert exempt)", v)
	}
}

func TestVerdict_StripsComments(t *testing.T) {
	raw := "# editor will fill this in\n# Please enter the commit message\nfeat: add\n\nbody"
	v := VerdictFromMessage([]byte(raw))
	if v.Decision != DecisionAllow {
		t.Fatalf("got %+v, want allow after comment strip", v)
	}
}
