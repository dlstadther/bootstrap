// Package exec provides the canonical command executors used across the bs CLI.
//
// Every internal package depends on one of these interfaces so command logic
// stays testable (tests pass fakes), and every cmd package wires in one of the
// concrete implementations. Centralizing them here gives a single place to add
// cross-cutting behavior later (logging, timeouts, tracing).
package exec

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Executor runs a command and returns its trimmed combined output.
type Executor interface {
	Run(cmd string, args ...string) (string, error)
}

// LookPathExecutor is an Executor that can also resolve a binary on $PATH.
type LookPathExecutor interface {
	Executor
	LookPath(name string) (string, error)
}

// Real shells out to real commands, capturing combined output. Use for
// commands whose output is parsed. It also satisfies LookPathExecutor.
type Real struct{}

func (Real) Run(cmd string, args ...string) (string, error) {
	out, err := exec.Command(cmd, args...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func (Real) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}

// Streaming streams stdout/stderr to the terminal while also capturing combined
// output so callers can include it in error messages.
type Streaming struct{}

func (Streaming) Run(cmd string, args ...string) (string, error) {
	var buf bytes.Buffer
	c := exec.Command(cmd, args...)
	c.Stdout = io.MultiWriter(os.Stdout, &buf)
	c.Stderr = io.MultiWriter(os.Stderr, &buf)
	err := c.Run()
	return strings.TrimSpace(buf.String()), err
}

// CMux shells out to real commands with the auto-set cmux context vars stripped,
// so programmatic calls aren't scoped to the calling terminal's
// workspace/surface when run from inside cmux.
type CMux struct{}

func (CMux) Run(cmd string, args ...string) (string, error) {
	c := exec.Command(cmd, args...)
	c.Env = stripCmuxContext(os.Environ())
	out, err := c.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func stripCmuxContext(env []string) []string {
	skip := map[string]bool{
		"CMUX_WORKSPACE_ID": true,
		"CMUX_TAB_ID":       true,
		"CMUX_SURFACE_ID":   true,
	}
	result := make([]string, 0, len(env))
	for _, e := range env {
		key, _, _ := strings.Cut(e, "=")
		if !skip[key] {
			result = append(result, e)
		}
	}
	return result
}
