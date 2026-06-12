package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

var _ Chatter = (*ClaudeHeadless)(nil)

// claudeSession is a stateful conversation backed by Claude Code's own session
// store. Only the session id is tracked client-side; the transcript lives in
// ~/.claude and is reconstructed via --resume, so history is NOT re-injected.
// The local mirror exists only so History() can return the conversation.
type claudeSession struct {
	claude    *ClaudeHeadless
	sessionID string
	system    string
	mirror    []Message
}

// claudeJSONResult is the subset of `claude --output-format json` we consume.
type claudeJSONResult struct {
	Result    string `json:"result"`
	SessionID string `json:"session_id"`
}

// NewSession starts a Claude conversation. The first Send creates the session;
// subsequent Sends resume it by id.
func (c *ClaudeHeadless) NewSession(opts ...SessionOption) Session {
	cfg := newSessionConfig(opts)

	return &claudeSession{
		claude: c,
		system: cfg.system,
		mirror: append([]Message(nil), cfg.history...),
	}
}

func (s *claudeSession) History() []Message {
	return s.mirror
}

func (s *claudeSession) Send(ctx context.Context, text string) (string, error) {
	bin := s.claude.Bin

	if bin == "" {
		bin = "claude"
	}

	args := []string{"-p", text, "--output-format", "json"}

	if s.sessionID != "" {
		args = append(args, "--resume", s.sessionID)
	}

	if s.claude.Model != "" {
		args = append(args, "--model", s.claude.Model)
	}

	if s.system != "" {
		args = append(args, "--append-system-prompt", s.system)
	}

	run := s.claude.runner

	if run == nil {
		run = execRunner
	}

	out, err := run(ctx, bin, args...)

	if err != nil {
		return "", fmt.Errorf("claude-headless: %w", err)
	}

	var res claudeJSONResult

	if err := json.Unmarshal(out, &res); err != nil {
		return "", fmt.Errorf("claude-headless: decode json: %w", err)
	}

	s.sessionID = res.SessionID

	reply := strings.TrimSpace(res.Result)

	s.mirror = append(s.mirror,
		Message{Role: "user", Content: text},
		Message{Role: "assistant", Content: reply},
	)

	return reply, nil
}
