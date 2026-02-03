package rules

import (
	"regexp"
	"strings"

	"github.com/makalin/preexec/internal/types"
)

// PipeToShellRule detects curl|bash, wget|sh, etc.
type PipeToShellRule struct{}

func (r *PipeToShellRule) ID() string { return "pipe-to-shell" }

var pipeToShellRe = regexp.MustCompile(`(?i)(curl|wget)\s+[^|]*\s*\|\s*(bash|sh|zsh|dash|ksh|csh|tcsh)\b`)

func (r *PipeToShellRule) Check(cmd string, cfg RuleConfig) []types.Finding {
	if cfg != nil && !cfg.RuleEnabled("pipe_to_shell") {
		return nil
	}
	var out []types.Finding
	for _, loc := range pipeToShellRe.FindAllStringIndex(cmd, -1) {
		token := cmd[loc[0]:loc[1]]
		out = append(out, types.Finding{
			RuleID:     r.ID(),
			Severity:   types.Warn,
			Token:      token,
			Issue:      "pattern: " + token,
			Position:   loc[0],
			Suggestion: "download, inspect, then execute",
		})
	}
	return out
}

// DotfileWriteRule detects writes to .bashrc, .zshrc, .profile, .ssh, etc.
type DotfileWriteRule struct{}

func (r *DotfileWriteRule) ID() string { return "dotfile-write" }

var dotfileTargets = []string{
	".bashrc", ".zshrc", ".profile", ".bash_profile",
	".ssh/", ">>", ">",
}

func (r *DotfileWriteRule) Check(cmd string, cfg RuleConfig) []types.Finding {
	if cfg != nil && !cfg.RuleEnabled("dotfile_write") {
		return nil
	}
	lower := strings.ToLower(cmd)
	var out []types.Finding
	for _, target := range dotfileTargets {
		idx := strings.Index(lower, target)
		if idx < 0 {
			continue
		}
		// Only flag if it looks like a write (redirect or echo/cat to file)
		before := strings.TrimSpace(lower[:idx])
		if strings.HasSuffix(before, ">>") || strings.HasSuffix(before, ">") ||
			strings.Contains(before, "echo") || strings.Contains(before, "cat") {
			token := tokenAround(cmd, idx, len(target))
			out = append(out, types.Finding{
				RuleID:     r.ID(),
				Severity:   types.Warn,
				Token:      token,
				Issue:      "writes to: " + target,
				Position:   idx,
				Suggestion: "review before writing to dotfiles or .ssh",
			})
		}
	}
	return out
}

func tokenAround(s string, start, width int) string {
	from := start
	for from > 0 && s[from-1] != ' ' && s[from-1] != '\n' {
		from--
	}
	to := start + width
	for to < len(s) && s[to] != ' ' && s[to] != '\n' {
		to++
	}
	return s[from:to]
}
