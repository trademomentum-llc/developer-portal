// Package main: score2openchoreo Score schema validator. Schema embedded at
// build time -- no network fetch at runtime. The file at assets/score.schema.json
// is vendored with an explicit SHA256 pin recorded in assets/SCHEMA_PROVENANCE.md;
// TestScoreSchemaPin guards against silent drift.
package main

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

// scoreSchemaJSON is pinned by SHA256 (see assets/SCHEMA_PROVENANCE.md).
// Bump both the file and the documented SHA256 in the same commit.
//
//go:embed assets/score.schema.json
var scoreSchemaJSON []byte

var compiledScoreSchema *jsonschema.Schema

func init() {
	compiler := jsonschema.NewCompiler()
	compiler.Draft = jsonschema.Draft2020
	if err := compiler.AddResource("score.json", strings.NewReader(string(scoreSchemaJSON))); err != nil {
		panic(err)
	}
	s, err := compiler.Compile("score.json")
	if err != nil {
		panic(err)
	}
	compiledScoreSchema = s
}

// ValidateScore validates raw YAML bytes against the embedded Score schema.
func ValidateScore(raw []byte) error {
	var node any
	if err := yaml.Unmarshal(raw, &node); err != nil {
		return fmt.Errorf("yaml parse: %w", err)
	}
	if err := compiledScoreSchema.Validate(node); err != nil {
		return fmt.Errorf("score schema: %w", err)
	}
	return nil
}
