// Package main: score2openchoreo Score schema validator. Schema embedded at
// build time -- no network fetch at runtime.
package main

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

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
