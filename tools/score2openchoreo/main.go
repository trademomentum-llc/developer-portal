// Package main: score2openchoreo entrypoint. Parses flags, validates Score,
// converts to OpenChoreo resources, writes YAML to stdout.
package main

import (
	"bytes"
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
	resources, err := Convert(doc, ConvertOptions{
		Environment: opts.Environment,
		Namespace:   opts.Namespace,
		ImageRef:    opts.Image,
		Project:     opts.Project,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "score2openchoreo: %s\n", err)
		os.Exit(1)
	}
	out, err := marshalDocuments(resources)
	if err != nil {
		fmt.Fprintf(os.Stderr, "score2openchoreo: marshal: %s\n", err)
		os.Exit(2)
	}
	_, _ = os.Stdout.Write(out)
}

func marshalDocuments(resources []OpenChoreoResource) ([]byte, error) {
	var out bytes.Buffer
	for i, res := range resources {
		if i > 0 {
			out.WriteString("---\n")
		}
		doc, err := yaml.Marshal(res)
		if err != nil {
			return nil, err
		}
		out.Write(doc)
	}
	return out.Bytes(), nil
}
