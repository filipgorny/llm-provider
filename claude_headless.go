package llm

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// ClaudeHeadless runs Claude Code in headless ("print") mode via the claude CLI.
type ClaudeHeadless struct {
	// Bin is the path to the claude binary. If empty, "claude" is used.
	Bin string

	// Model optionally selects a specific model via the --model flag.
	Model string

	// runner executes the command. It is injectable for testing; if nil, the
	// real claude binary is invoked.
	runner commandRunner
}

// commandRunner runs a command and returns its stdout.
type commandRunner func(ctx context.Context, name string, args ...string) ([]byte, error)

var _ Llm = (*ClaudeHeadless)(nil)

// NewClaudeHeadless returns a ClaudeHeadless backend using the claude CLI.
func NewClaudeHeadless() *ClaudeHeadless {
	return &ClaudeHeadless{
		Bin:    "claude",
		runner: execRunner,
	}
}

func execRunner(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	}

	return stdout.Bytes(), nil
}

// Prompt sends prompt to Claude Code in headless mode and returns its output.
func (c *ClaudeHeadless) Prompt(ctx context.Context, prompt string) (string, error) {
	bin := c.Bin

	if bin == "" {
		bin = "claude"
	}

	args := []string{"-p", prompt}

	if c.Model != "" {
		args = append(args, "--model", c.Model)
	}

	run := c.runner

	if run == nil {
		run = execRunner
	}

	out, err := run(ctx, bin, args...)

	if err != nil {
		return "", fmt.Errorf("claude-headless: %w", err)
	}

	return strings.TrimSpace(string(out)), nil
}
