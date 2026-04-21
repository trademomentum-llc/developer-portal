package main

import "testing"

func TestValidateTofuCommand(t *testing.T) {
	tests := []struct {
		name   string
		cmd    string
		allow  bool
		action string
	}{
		{"version", "tofu version", true, "allow"},
		{"init", "tofu init", true, "allow"},
		{"plan", "tofu plan", true, "allow"},
		{"validate", "tofu validate", true, "allow"},
		{"fmt", "tofu fmt", true, "allow"},
		{"state list", "tofu state list", true, "allow"},
		{"state show", "tofu state show aws_instance.foo", true, "allow"},
		{"output", "tofu output", true, "allow"},
		{"not tofu", "ls -la", true, "not-applicable"},

		{"apply", "tofu apply", false, "block"},
		{"apply auto", "tofu apply -auto-approve", false, "block"},
		{"destroy", "tofu destroy", false, "block"},
		{"import", "tofu import aws_instance.foo i-123", false, "block"},
		{"state rm", "tofu state rm aws_instance.foo", false, "block"},
		{"state mv", "tofu state mv a b", false, "block"},
		{"state push", "tofu state push state.tfstate", false, "block"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toks, err := Tokenize(tt.cmd)
			if err != nil {
				t.Fatalf("tokenize: %v", err)
			}
			d := ValidateTofuCommand(toks)
			if d.Allow != tt.allow {
				t.Errorf("allow=%v want %v (reason=%q)", d.Allow, tt.allow, d.Reason)
			}
			if d.Action != tt.action {
				t.Errorf("action=%q want %q", d.Action, tt.action)
			}
		})
	}
}

func TestTokenizeQuotes(t *testing.T) {
	got, err := Tokenize(`tofu state show "aws_instance.my name"`)
	if err != nil {
		t.Fatalf("tokenize: %v", err)
	}
	want := []string{"tofu", "state", "show", "aws_instance.my name"}
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("got[%d]=%q want %q", i, got[i], want[i])
		}
	}
}

func TestTokenizeUnterminatedQuote(t *testing.T) {
	if _, err := Tokenize(`tofu plan "unterm`); err == nil {
		t.Fatal("expected error on unterminated quote")
	}
}
