package main

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

// Score-5: the embedded Score schema is pinned by SHA256. Any change to
// assets/score.schema.json must be accompanied by a matching update to
// assets/SCHEMA_PROVENANCE.md and to the constant below. This guards
// against silent drift from a moving upstream branch.
const scoreSchemaSHA256 = "633c4394dfc03977c86932f5d10c77e2c356a38b63f162c281104fade5c9863c"

func TestScoreSchemaPin(t *testing.T) {
	sum := sha256.Sum256(scoreSchemaJSON)
	got := hex.EncodeToString(sum[:])
	if got != scoreSchemaSHA256 {
		t.Fatalf("embedded score schema SHA256 drift\n  got:  %s\n  want: %s\nBump the constant and SCHEMA_PROVENANCE.md in the same commit if the change is intentional.", got, scoreSchemaSHA256)
	}
}
