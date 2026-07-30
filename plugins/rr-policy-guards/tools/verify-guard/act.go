// act.go -- act CLI integration for workflow grammar validation
// (`act --list`) and full local execution (`act --rm`).
//
// All exec.CommandContext calls in this file pass "act" as a literal
// constant -- no caller-controlled binary string can reach the OS.

package main

import (
	"bytes"
	"context"
	"errors"
	osexec "os/exec"
	"time"
)

// ActAvailable reports whether `act` is on PATH.
func ActAvailable() bool { return commandOnPath("act") }

// DockerAvailable reports whether `docker info` succeeds (a precondition
// for `act --rm` full runs).
func DockerAvailable() bool {
	if !commandOnPath("docker") {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := osexec.CommandContext(ctx, "docker", "info")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

// ActListWorkflows runs `act --list -W <file>` per workflow.
// Returns one Result per workflow; the first failing Result is also
// returned as a pointer for convenient block dispatch.
func ActListWorkflows(workflows []Workflow) ([]Result, *Result) {
	var results []Result
	for _, w := range workflows {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		cmd := osexec.CommandContext(ctx, "act", "--list", "-W", w.Path)
		var buf bytes.Buffer
		cmd.Stdout = &buf
		cmd.Stderr = &buf
		err := cmd.Run()
		cancel()
		exit := 0
		if err != nil {
			var ee *osexec.ExitError
			if errors.As(err, &ee) {
				exit = ee.ExitCode()
			} else {
				exit = 1
			}
		}
		r := Result{
			Toolchain: "act",
			StepName:  "list",
			Cmd:       "act",
			Args:      []string{"--list", "-W", w.Path},
			ExitCode:  exit,
			Truncated: firstLines(buf.String(), 50),
			Err:       err,
		}
		results = append(results, r)
		if exit != 0 {
			rr := r
			return results, &rr
		}
	}
	return results, nil
}

// ActRunWorkflows runs `act -W <file> --rm` per workflow.
func ActRunWorkflows(workflows []Workflow) ([]Result, *Result) {
	var results []Result
	for _, w := range workflows {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
		cmd := osexec.CommandContext(ctx, "act", "-W", w.Path, "--rm")
		var buf bytes.Buffer
		cmd.Stdout = &buf
		cmd.Stderr = &buf
		err := cmd.Run()
		cancel()
		exit := 0
		if err != nil {
			var ee *osexec.ExitError
			if errors.As(err, &ee) {
				exit = ee.ExitCode()
			} else {
				exit = 1
			}
		}
		r := Result{
			Toolchain: "act",
			StepName:  "run",
			Cmd:       "act",
			Args:      []string{"-W", w.Path, "--rm"},
			ExitCode:  exit,
			Truncated: firstLines(buf.String(), 50),
			Err:       err,
		}
		results = append(results, r)
		if exit != 0 {
			rr := r
			return results, &rr
		}
	}
	return results, nil
}
