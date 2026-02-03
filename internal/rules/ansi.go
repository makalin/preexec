package rules

import (
	"github.com/makalin/preexec/internal/types"
)

// ANSIEscapeRule detects ANSI escape sequences (ESC, \x1b).
type ANSIEscapeRule struct{}

func (r *ANSIEscapeRule) ID() string { return "ansi-escape" }

func (r *ANSIEscapeRule) Check(cmd string, cfg RuleConfig) []types.Finding {
	if cfg != nil && !cfg.RuleEnabled("ansi_escape") {
		return nil
	}
	var out []types.Finding
	for i := 0; i < len(cmd); i++ {
		if cmd[i] == 0x1B {
			// Consume full CSI/OSC sequence for token
			end := i + 1
			for end < len(cmd) {
				c := cmd[end]
				if c >= 0x40 && c <= 0x7E {
					end++
					break
				}
				if c == 0x07 || c == '\n' {
					end++
					break
				}
				end++
				if end-i > 64 {
					break
				}
			}
			token := cmd[i:end]
			out = append(out, types.Finding{
				RuleID:     r.ID(),
				Severity:   types.Warn,
				Token:      token,
				Issue:      "ESC (U+001B) control sequence detected",
				Position:   i,
				Suggestion: "paste as plain text",
			})
			i = end - 1
		}
	}
	return out
}
