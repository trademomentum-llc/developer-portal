// Package main: score2openchoreo CLI flag parsing and input reading.
package main

import (
	"errors"
	"flag"
	"io"
	"os"
)

type CLIOptions struct {
	Input        string
	Environment  string
	Namespace    string
	Project      string
	Image        string
	ValidateOnly bool
}

func parseFlags(args []string) (CLIOptions, error) {
	fs := flag.NewFlagSet("score2openchoreo", flag.ContinueOnError)
	var o CLIOptions
	fs.StringVar(&o.Input, "input", "", "path to score.yaml (default: stdin)")
	fs.StringVar(&o.Environment, "environment", "", "target OpenChoreo environment (dev|staging)")
	fs.StringVar(&o.Namespace, "namespace", "openchoreo-data-plane", "target namespace")
	fs.StringVar(&o.Project, "project", "openchoreo", "OpenChoreo project name")
	fs.StringVar(&o.Image, "image", "", "override container image reference")
	fs.BoolVar(&o.ValidateOnly, "validate-only", false, "validate schema and exit without emitting output")
	if err := fs.Parse(args); err != nil {
		return o, err
	}
	if !o.ValidateOnly && o.Environment == "" {
		return o, errors.New("--environment required unless --validate-only")
	}
	return o, nil
}

func readInput(opts CLIOptions) ([]byte, error) {
	if opts.Input != "" {
		return os.ReadFile(opts.Input)
	}
	return io.ReadAll(os.Stdin)
}
