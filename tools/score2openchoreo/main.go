// Package main: score2openchoreo entrypoint. Parses flags, validates Score,
// converts to OpenChoreo Component, writes YAML to stdout.
package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func main() {
	opts, err := parseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "score2openchoreo: %s\n", err)
		os.Exit(2)
	}
	raw, err := readInput(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "score2openchoreo: %s\n", err)
		os.Exit(2)
	}
	if err := ValidateScore(raw); err != nil {
		fmt.Fprintf(os.Stderr, "score2openchoreo: %s\n", err)
		os.Exit(1)
	}
	if opts.ValidateOnly {
		return
	}
	var doc ScoreDocument
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		fmt.Fprintf(os.Stderr, "score2openchoreo: yaml: %s\n", err)
		os.Exit(1)
	}
	comp, err := Convert(doc, ConvertOptions{
		Environment: opts.Environment,
		Namespace:   opts.Namespace,
		ImageRef:    opts.Image,
		Project:     opts.Project,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "score2openchoreo: %s\n", err)
		os.Exit(1)
	}
	out, err := yaml.Marshal(comp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "score2openchoreo: marshal: %s\n", err)
		os.Exit(2)
	}
	_, _ = os.Stdout.Write(out)
}
