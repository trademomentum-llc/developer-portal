// Package main: score2openchoreo CLI flag parsing and input reading.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

// envSlice is a flag.Value that collects repeated --extra-env KEY=VALUE pairs.
type envSlice map[string]string

func (e envSlice) String() string {
	if len(e) == 0 {
		return ""
	}
	var parts []string
	for k, v := range e {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(parts, ",")
}

func (e envSlice) Set(value string) error {
	idx := strings.IndexByte(value, '=')
	if idx <= 0 {
		return fmt.Errorf("--extra-env must be KEY=VALUE, got %q", value)
	}
	key := value[:idx]
	val := value[idx+1:]
	if e == nil {
		return fmt.Errorf("envSlice not initialized")
	}
	e[key] = val
	return nil
}

type CLIOptions struct {
	Input        string
	Environment  string
	Namespace    string
	Project      string
	Image        string
	ValidateOnly bool
	ExtraEnv     map[string]string
}

func parseFlags(args []string) (CLIOptions, error) {
	fs := flag.NewFlagSet("score2openchoreo", flag.ContinueOnError)
	var o CLIOptions
	o.ExtraEnv = make(map[string]string)
	fs.StringVar(&o.Input, "input", "", "path to score.yaml (default: stdin)")
	fs.StringVar(&o.Environment, "environment", "", "target OpenChoreo environment (dev|staging)")
	fs.StringVar(&o.Namespace, "namespace", "default", "OpenChoreo project namespace")
	fs.StringVar(&o.Project, "project", "default", "OpenChoreo project name")
	fs.StringVar(&o.Image, "image", "", "override container image reference")
	fs.BoolVar(&o.ValidateOnly, "validate-only", false, "validate schema and exit without emitting output")
	fs.Var(envSlice(o.ExtraEnv), "extra-env", "extra container env var (repeatable, KEY=VALUE)")
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
